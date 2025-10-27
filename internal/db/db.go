package db

import (
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

	// Open a connection to the database using sqlite driver
	dbConn, err := sql.Open("sqlite", fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set up connection pool settings
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(25)
	dbConn.SetConnMaxLifetime(0) // Connections never expire

	// Create tables if they don't exist
	if err := createTables(dbConn); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
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

	// Create messages table
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

	return nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}
