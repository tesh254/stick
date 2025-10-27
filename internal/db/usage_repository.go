package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dombox/uuidv7"
)

// usageRepository implements UsageRepository
type usageRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// getExecer returns the appropriate execer (transaction or regular)
func (r *usageRepository) getExecer() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new usage record
func (r *usageRepository) Create(ctx context.Context, usage *Usage) error {
	query := `
		INSERT INTO usage (id, prompt_tokens, completion_tokens, total_tokens, model, conversation_id, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.getExecer().ExecContext(ctx, query, 
		usage.ID.String(), 
		usage.PromptTokens, 
		usage.CompletionTokens, 
		usage.TotalTokens, 
		usage.Model, 
		usage.Conversation.String(), 
		usage.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create usage: %w", err)
	}
	
	return nil
}

// GetByID retrieves a usage record by ID
func (r *usageRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Usage, error) {
	query := `SELECT id, prompt_tokens, completion_tokens, total_tokens, model, conversation_id, created_at FROM usage WHERE id = ?`
	
	var usageID string
	var usage Usage
	var createdAt time.Time
	var convID string
	
	err := r.getExecer().QueryRowContext(ctx, query, id.String()).Scan(
		&usageID, 
		&usage.PromptTokens, 
		&usage.CompletionTokens, 
		&usage.TotalTokens, 
		&usage.Model, 
		&convID, 
		&createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usage with id %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	
	// Parse usage ID
	parsedUsageID, err := uuidv7.Parse(usageID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse usage ID: %w", err)
	}
	usage.ID = parsedUsageID
	
	// Parse conversation ID
	convUUID, err := uuidv7.Parse(convID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
	}
	usage.Conversation = convUUID
	
	usage.CreatedAt = createdAt
	
	return &usage, nil
}

// GetByConversationID retrieves all usage records for a conversation
func (r *usageRepository) GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*Usage, error) {
	query := `SELECT id, prompt_tokens, completion_tokens, total_tokens, model, conversation_id, created_at FROM usage WHERE conversation_id = ? ORDER BY created_at ASC`
	
	rows, err := r.getExecer().QueryContext(ctx, query, conversationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}
	defer rows.Close()
	
	var usages []*Usage
	for rows.Next() {
		var usageID string
		var usage Usage
		var createdAt time.Time
		var convID string
		
		err := rows.Scan(
			&usageID,
			&usage.PromptTokens,
			&usage.CompletionTokens, 
			&usage.TotalTokens,
			&usage.Model,
			&convID,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}
		
		// Parse usage ID
		parsedUsageID, err := uuidv7.Parse(usageID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse usage ID: %w", err)
		}
		usage.ID = parsedUsageID
		
		// Parse conversation ID
		convUUID, err := uuidv7.Parse(convID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
		}
		usage.Conversation = convUUID
		
		usage.CreatedAt = createdAt
		usages = append(usages, &usage)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over usage records: %w", err)
	}
	
	return usages, nil
}

// GetTotalByConversationID calculates the total usage for a conversation
func (r *usageRepository) GetTotalByConversationID(ctx context.Context, conversationID uuidv7.UUID) (*Usage, error) {
	query := `
		SELECT 
			'00000000-0000-0000-0000-000000000000' as id,
			SUM(prompt_tokens) as prompt_tokens,
			SUM(completion_tokens) as completion_tokens,
			SUM(total_tokens) as total_tokens,
			MAX(model) as model,
			? as conversation_id,
			MAX(created_at) as created_at
		FROM usage 
		WHERE conversation_id = ?
		GROUP BY conversation_id
	`
	
	var usage Usage
	var createdAt time.Time
	var convID string
	
	var usageID string
	
	err := r.getExecer().QueryRowContext(ctx, query, conversationID.String(), conversationID.String()).Scan(
		&usageID,
		&usage.PromptTokens,
		&usage.CompletionTokens,
		&usage.TotalTokens,
		&usage.Model,
		&convID,
		&createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a zero usage if no records found
			return &Usage{
				ID:               uuidv7.UUID{}, // Empty UUID for the special aggregated case
				Conversation:     conversationID,
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
				CreatedAt:        time.Now(), // Set a default time for zero usage
			}, nil
		}
		return nil, fmt.Errorf("failed to get total usage: %w", err)
	}
	
	// Parse usage ID
	parsedUsageID, err := uuidv7.Parse(usageID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse usage ID: %w", err)
	}
	usage.ID = parsedUsageID
	
	// Parse conversation ID
	convUUID, err := uuidv7.Parse(convID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
	}
	usage.Conversation = convUUID
	
	usage.CreatedAt = createdAt
	
	return &usage, nil
}

// Update updates a usage record
func (r *usageRepository) Update(ctx context.Context, id uuidv7.UUID, usage *Usage) error {
	query := `
		UPDATE usage 
		SET prompt_tokens = ?, completion_tokens = ?, total_tokens = ?, model = ? 
		WHERE id = ?
	`
	
	result, err := r.getExecer().ExecContext(ctx, query, 
		usage.PromptTokens, 
		usage.CompletionTokens, 
		usage.TotalTokens, 
		usage.Model, 
		id.String())
	if err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("usage with id %s not found", id.String())
	}
	
	return nil
}

// Delete deletes a usage record by ID
func (r *usageRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM usage WHERE id = ?`
	
	result, err := r.getExecer().ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete usage: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("usage with id %s not found", id.String())
	}
	
	return nil
}

// GetUsageSummary retrieves usage summary with pagination
func (r *usageRepository) GetUsageSummary(ctx context.Context, limit, offset int) ([]*Usage, error) {
	query := `
		SELECT 
			'00000000-0000-0000-0000-000000000000' as id,
			SUM(prompt_tokens) as prompt_tokens,
			SUM(completion_tokens) as completion_tokens,
			SUM(total_tokens) as total_tokens,
			MAX(model) as model,
			conversation_id,
			MAX(created_at) as created_at
		FROM usage 
		GROUP BY conversation_id
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.getExecer().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}
	defer rows.Close()
	
	var usages []*Usage
	for rows.Next() {
		var usageID string
		var usage Usage
		var createdAt time.Time
		var convID string
		
		err := rows.Scan(
			&usageID,
			&usage.PromptTokens,
			&usage.CompletionTokens,
			&usage.TotalTokens,
			&usage.Model,
			&convID,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage summary: %w", err)
		}
		
		// Parse usage ID
		parsedUsageID, err := uuidv7.Parse(usageID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse usage ID: %w", err)
		}
		usage.ID = parsedUsageID
		
		// Parse conversation ID
		convUUID, err := uuidv7.Parse(convID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
		}
		usage.Conversation = convUUID
		
		usage.CreatedAt = createdAt
		usages = append(usages, &usage)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over usage summary: %w", err)
	}
	
	return usages, nil
}