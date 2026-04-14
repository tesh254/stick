package db

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"

    _ "github.com/tursodatabase/libsql-client-go/libsql"
    _ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	dbPath string
}

// New creates a new database connection to a Turso database
func New() (*DB, error) {
	// Get the home directory to store the database file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create a database file path in the home directory
	dbPath := filepath.Join(homeDir, ".stick", "stick.db")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

    // Open a connection to the database using sqlite driver (local file). For Turso remote,
    // configure LIBSQL connection string if available.
    dbConn, err := sql.Open("sqlite", fmt.Sprintf("file:%s", dbPath))
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

	// Set up connection pool settings
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(25)
	dbConn.SetConnMaxLifetime(0) // Connections never expire

    // Create tables and run lightweight migrations
    if err := createTables(dbConn); err != nil {
        return nil, fmt.Errorf("failed to create tables: %w", err)
    }
    if err := migrateSchema(dbConn); err != nil {
        return nil, fmt.Errorf("failed to migrate schema: %w", err)
    }

	return &DB{
		DB:     dbConn,
		dbPath: dbPath,
	}, nil
}

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	// Create conversations table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			working_directory TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create conversations table: %w", err)
	}

    // Create messages table (base columns; migrations add seq/parent)
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            content TEXT NOT NULL,
            role TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations (id)
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create messages table: %w", err)
    }

    // Create usage table
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS usage (
            id TEXT PRIMARY KEY,
            prompt_tokens INTEGER DEFAULT 0,
            completion_tokens INTEGER DEFAULT 0,
            total_tokens INTEGER DEFAULT 0,
            model TEXT,
            conversation_id TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations (id)
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create usage table: %w", err)
    }

    // Create calls table for function/tool metadata
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS calls (
            id TEXT PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            parent_message_id TEXT,
            type TEXT NOT NULL,
            name TEXT NOT NULL,
            params_json TEXT NOT NULL,
            status TEXT NOT NULL,
            result_raw TEXT,
            error TEXT,
            started_at DATETIME,
            completed_at DATETIME,
            duration_ms INTEGER,
            FOREIGN KEY (conversation_id) REFERENCES conversations (id),
            FOREIGN KEY (parent_message_id) REFERENCES messages (id)
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create calls table: %w", err)
    }

    // Create settings table for AI provider configurations
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS settings (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            provider_name TEXT NOT NULL UNIQUE,
            api_key TEXT NOT NULL,
            model TEXT NOT NULL,
            endpoint TEXT NOT NULL,
            extra_params TEXT,  -- JSON for additional parameters
            is_default BOOLEAN DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create settings table: %w", err)
    }

    return nil
}

// migrateSchema performs lightweight ALTERs to add columns and indexes without full migrations.
func migrateSchema(db *sql.DB) error {
    // Add seq column to messages if missing
    _, err := db.Exec(`ALTER TABLE messages ADD COLUMN seq INTEGER`)
    if err != nil {
        // ignore error if column exists
        _ = err
    }
    // Add parent_message_id if missing
    _, err = db.Exec(`ALTER TABLE messages ADD COLUMN parent_message_id TEXT`)
    if err != nil {
        _ = err
    }
    // Create indexes (idempotent in SQLite)
    _, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_conv_seq ON messages (conversation_id, seq)`)
    _, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_calls_conv_started ON calls (conversation_id, started_at)`)
    return nil
}

// SaveProviderSettings saves or updates the settings for a provider
func (d *DB) SaveProviderSettings(ctx context.Context, settings *ProviderSettings) error {
	stmt, err := d.PrepareContext(ctx, `
		INSERT OR REPLACE INTO settings (provider_name, api_key, model, endpoint, extra_params, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare save settings statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, settings.ProviderName, settings.APIKey, settings.Model, settings.Endpoint, settings.ExtraParams, settings.IsDefault, settings.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}
	return nil
}

// LoadProviderSettings loads the settings for a specific provider
func (d *DB) LoadProviderSettings(ctx context.Context, providerName string) (*ProviderSettings, error) {
	row := d.QueryRowContext(ctx, `
		SELECT provider_name, api_key, model, endpoint, extra_params, is_default, created_at, updated_at
		FROM settings
		WHERE provider_name = ?
	`, providerName)

	settings := &ProviderSettings{}
	err := row.Scan(&settings.ProviderName, &settings.APIKey, &settings.Model, &settings.Endpoint, &settings.ExtraParams, &settings.IsDefault, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No settings found
		}
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	return settings, nil
}

// GetDefaultProvider returns the name of the default provider
func (d *DB) GetDefaultProvider(ctx context.Context) (string, error) {
	var providerName string
	err := d.QueryRowContext(ctx, `
		SELECT provider_name
		FROM settings
		WHERE is_default = 1
		LIMIT 1
	`).Scan(&providerName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No default set
		}
		return "", fmt.Errorf("failed to get default provider: %w", err)
	}
	return providerName, nil
}

// SetDefaultProvider sets a provider as default and unsets others
func (d *DB) SetDefaultProvider(ctx context.Context, providerName string) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all defaults
	_, err = tx.ExecContext(ctx, `UPDATE settings SET is_default = 0`)
	if err != nil {
		return fmt.Errorf("failed to unset defaults: %w", err)
	}

	// Set the new default
	_, err = tx.ExecContext(ctx, `UPDATE settings SET is_default = 1 WHERE provider_name = ?`, providerName)
	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return tx.Commit()
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}
