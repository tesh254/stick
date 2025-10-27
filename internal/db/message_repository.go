package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dombox/uuidv7"
)

// messageRepository implements MessageRepository
type messageRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// getExecer returns the appropriate execer (transaction or regular)
func (r *messageRepository) getExecer() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, content, role, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`
	
	_, err := r.getExecer().ExecContext(ctx, query, message.ID.String(), message.Conversation.String(), message.Content, message.Role.String(), message.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	
	return nil
}

// GetByConversationID retrieves all messages for a conversation
func (r *messageRepository) GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*Message, error) {
	query := `SELECT id, conversation_id, content, role, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`
	
	rows, err := r.getExecer().QueryContext(ctx, query, conversationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()
	
	var messages []*Message
	for rows.Next() {
		var msgID string
		var msg Message
		var createdAt time.Time
		var convID string
		var roleStr string
		
		err := rows.Scan(&msgID, &convID, &msg.Content, &roleStr, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		
		// Parse message ID
		msgUUID, err := uuidv7.Parse(msgID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse message ID: %w", err)
		}
		msg.ID = msgUUID
		
		// Parse conversation ID
		convUUID, err := uuidv7.Parse(convID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
		}
		msg.Conversation = convUUID
		
		// Parse role
		role, err := FromString(roleStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse message role: %w", err)
		}
		msg.Role = role
		
		msg.CreatedAt = createdAt
		messages = append(messages, &msg)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}
	
	return messages, nil
}

// GetByID retrieves a message by ID
func (r *messageRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Message, error) {
	query := `SELECT id, conversation_id, content, role, created_at FROM messages WHERE id = ?`
	
	var msgID string
	var msg Message
	var createdAt time.Time
	var convID string
	var roleStr string
	
	err := r.getExecer().QueryRowContext(ctx, query, id.String()).Scan(&msgID, &convID, &msg.Content, &roleStr, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message with id %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	// Parse message ID
	parsedMsgID, err := uuidv7.Parse(msgID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message ID: %w", err)
	}
	msg.ID = parsedMsgID
	
	// Parse conversation ID
	convUUID, err := uuidv7.Parse(convID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
	}
	msg.Conversation = convUUID
	
	// Parse role
	role, err := FromString(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message role: %w", err)
	}
	msg.Role = role
	
	msg.CreatedAt = createdAt
	
	return &msg, nil
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, id uuidv7.UUID, content string) error {
	query := `UPDATE messages SET content = ? WHERE id = ?`
	
	result, err := r.getExecer().ExecContext(ctx, query, content, id.String())
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("message with id %s not found", id.String())
	}
	
	return nil
}

// Delete deletes a message by ID
func (r *messageRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM messages WHERE id = ?`
	
	result, err := r.getExecer().ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("message with id %s not found", id.String())
	}
	
	return nil
}

// DeleteByConversationID deletes all messages for a conversation
func (r *messageRepository) DeleteByConversationID(ctx context.Context, conversationID uuidv7.UUID) error {
	query := `DELETE FROM messages WHERE conversation_id = ?`
	
	_, err := r.getExecer().ExecContext(ctx, query, conversationID.String())
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	
	return nil
}

// GetCountByConversationID retrieves the count of messages in a conversation
func (r *messageRepository) GetCountByConversationID(ctx context.Context, conversationID uuidv7.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = ?`
	
	var count int
	err := r.getExecer().QueryRowContext(ctx, query, conversationID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	
	return count, nil
}