package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dombox/uuidv7"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func TestDBIntegration(t *testing.T) {
	// Create a temporary database for testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Override the database creation to use the temp path
	dbConn, err := setupTestDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbConn.Close()

	rm := NewRepositoryManager(dbConn)
	ctx := context.Background()

	// Test conversation operations
	t.Run("Conversation CRUD", func(t *testing.T) {
		convID, err := uuidv7.New()
		if err != nil {
			t.Fatalf("Failed to create UUID: %v", err)
		}
		conversation := &Conversation{
			ID:              convID,
			Title:           "Test Conversation",
			WorkingDirectory: "/test/working/directory",
			CreatedAt:       time.Now(),
		}

		// Create
		err = rm.Conversations().Create(ctx, conversation)
		if err != nil {
			t.Fatalf("Failed to create conversation: %v", err)
		}

		// Get by ID
		retrieved, err := rm.Conversations().GetByID(ctx, convID)
		if err != nil {
			t.Fatalf("Failed to get conversation: %v", err)
		}

		if retrieved.Title != conversation.Title {
			t.Errorf("Expected title %s, got %s", conversation.Title, retrieved.Title)
		}

		// Update
		newTitle := "Updated Conversation"
		err = rm.Conversations().Update(ctx, convID, newTitle)
		if err != nil {
			t.Fatalf("Failed to update conversation: %v", err)
		}

		// Verify update
		updated, err := rm.Conversations().GetByID(ctx, convID)
		if err != nil {
			t.Fatalf("Failed to get updated conversation: %v", err)
		}

		if updated.Title != newTitle {
			t.Errorf("Expected updated title %s, got %s", newTitle, updated.Title)
		}

		// Get all (should have 1)
		all, err := rm.Conversations().GetAll(ctx, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get all conversations: %v", err)
		}

		if len(all) != 1 {
			t.Errorf("Expected 1 conversation, got %d", len(all))
		}
	})

	// Test message operations
	t.Run("Message CRUD", func(t *testing.T) {
		convID, err := uuidv7.New()
		if err != nil {
			t.Fatalf("Failed to create conversation UUID: %v", err)
		}
		msgID, err := uuidv7.New()
		if err != nil {
			t.Fatalf("Failed to create message UUID: %v", err)
		}
		
		// Create a conversation first
		conversation := &Conversation{
			ID:              convID,
			Title:           "Test Conversation for Messages",
			WorkingDirectory: "/test/messages/directory",
			CreatedAt:       time.Now(),
		}
		err = rm.Conversations().Create(ctx, conversation)
		if err != nil {
			t.Fatalf("Failed to create conversation for message test: %v", err)
		}

		message := &Message{
			ID:           msgID,
			Conversation: convID,
			Content:      "Test message content",
			Role:         User,
			CreatedAt:    time.Now(),
		}

		// Create message
		err = rm.Messages().Create(ctx, message)
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}

		// Get by ID
		retrieved, err := rm.Messages().GetByID(ctx, msgID)
		if err != nil {
			t.Fatalf("Failed to get message: %v", err)
		}

		if retrieved.Content != message.Content {
			t.Errorf("Expected content %s, got %s", message.Content, retrieved.Content)
		}

		// Get by conversation ID
		messages, err := rm.Messages().GetByConversationID(ctx, convID)
		if err != nil {
			t.Fatalf("Failed to get messages by conversation: %v", err)
		}

		if len(messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(messages))
		}
	})

	// Test usage operations
	t.Run("Usage CRUD", func(t *testing.T) {
		convID, err := uuidv7.New()
		if err != nil {
			t.Fatalf("Failed to create conversation UUID: %v", err)
		}
		usageID, err := uuidv7.New()
		if err != nil {
			t.Fatalf("Failed to create usage UUID: %v", err)
		}
		
		// Create a conversation first
		conversation := &Conversation{
			ID:              convID,
			Title:           "Test Conversation for Usage",
			WorkingDirectory: "/test/usage/directory",
			CreatedAt:       time.Now(),
		}
		err = rm.Conversations().Create(ctx, conversation)
		if err != nil {
			t.Fatalf("Failed to create conversation for usage test: %v", err)
		}

		usage := &Usage{
			ID:               usageID,
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
			Model:            "gpt-4",
			Conversation:     convID,
			CreatedAt:        time.Now(),
		}

		// Create usage
		err = rm.Usage().Create(ctx, usage)
		if err != nil {
			t.Fatalf("Failed to create usage: %v", err)
		}

		// Get by ID
		retrieved, err := rm.Usage().GetByID(ctx, usageID)
		if err != nil {
			t.Fatalf("Failed to get usage: %v", err)
		}

		if retrieved.TotalTokens != usage.TotalTokens {
			t.Errorf("Expected total tokens %d, got %d", usage.TotalTokens, retrieved.TotalTokens)
		}

		// Get by conversation ID
		usages, err := rm.Usage().GetByConversationID(ctx, convID)
		if err != nil {
			t.Fatalf("Failed to get usage by conversation: %v", err)
		}

		if len(usages) != 1 {
			t.Errorf("Expected 1 usage record, got %d", len(usages))
		}
	})
}

// setupTestDB creates a test database with the same setup as New() but using a specific path
func setupTestDB(dbPath string) (*DB, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	// Open a connection to the database using sqlite driver
	dbConn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return nil, err
	}

	// Set up connection pool settings
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(25)
	dbConn.SetConnMaxLifetime(0) // Connections never expire

	// Create tables if they don't exist
	if err := createTables(dbConn); err != nil {
		return nil, err
	}

	// Apply schema migrations
	if err := migrateSchema(dbConn); err != nil {
		return nil, err
	}

	return &DB{
		DB:     dbConn,
		dbPath: dbPath,
	}, nil
}