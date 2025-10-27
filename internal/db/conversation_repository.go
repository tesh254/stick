package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dombox/uuidv7"
)

// conversationRepository implements ConversationRepository
type conversationRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// getExecer returns the appropriate execer (transaction or regular)
func (r *conversationRepository) getExecer() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new conversation
func (r *conversationRepository) Create(ctx context.Context, conversation *Conversation) error {
	query := `
		INSERT INTO conversations (id, title, working_directory, created_at) 
		VALUES (?, ?, ?, ?)
	`
	
	_, err := r.getExecer().ExecContext(ctx, query, conversation.ID.String(), conversation.Title, conversation.WorkingDirectory, conversation.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	
	return nil
}

// GetByID retrieves a conversation by ID
func (r *conversationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Conversation, error) {
	query := `SELECT id, title, working_directory, created_at FROM conversations WHERE id = ?`
	
	var convID string
	var conv Conversation
	var createdAt time.Time
	
	err := r.getExecer().QueryRowContext(ctx, query, id.String()).Scan(&convID, &conv.Title, &conv.WorkingDirectory, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation with id %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	
	// Parse the ID string into UUID
	parsedID, err := uuidv7.Parse(convID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
	}
	conv.ID = parsedID
	conv.CreatedAt = createdAt
	
	return &conv, nil
}

// GetAll retrieves all conversations with pagination
func (r *conversationRepository) GetAll(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	query := `SELECT id, title, working_directory, created_at FROM conversations ORDER BY created_at DESC LIMIT ? OFFSET ?`
	
	rows, err := r.getExecer().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()
	
	var conversations []*Conversation
	for rows.Next() {
		var convID string
		var conv Conversation
		var createdAt time.Time
		
		err := rows.Scan(&convID, &conv.Title, &conv.WorkingDirectory, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		
		// Parse the ID string into UUID
		parsedID, err := uuidv7.Parse(convID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
		}
		conv.ID = parsedID
		conv.CreatedAt = createdAt
		conversations = append(conversations, &conv)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over conversations: %w", err)
	}
	
	return conversations, nil
}

// Update updates a conversation
func (r *conversationRepository) Update(ctx context.Context, id uuidv7.UUID, title string) error {
	// Only update the title field for now, as working directory is typically set during creation
	query := `UPDATE conversations SET title = ? WHERE id = ?`
	
	result, err := r.getExecer().ExecContext(ctx, query, title, id.String())
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("conversation with id %s not found", id.String())
	}
	
	return nil
}

// Delete deletes a conversation by ID
func (r *conversationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM conversations WHERE id = ?`
	
	result, err := r.getExecer().ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("conversation with id %s not found", id.String())
	}
	
	return nil
}

// GetWithMessages retrieves a conversation with all its messages
func (r *conversationRepository) GetWithMessages(ctx context.Context, id uuidv7.UUID) (*Conversation, []*Message, error) {
	// First get the conversation
	conversation, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	
	// Then get all messages for this conversation
	messageRepo := &messageRepository{db: r.db, tx: r.tx}
	messages, err := messageRepo.GetByConversationID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	
	return conversation, messages, nil
}