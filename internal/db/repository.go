package db

import (
	"context"
	"database/sql"

	"github.com/dombox/uuidv7"
)

// ConversationRepository defines the interface for conversation operations
type ConversationRepository interface {
	Create(ctx context.Context, conversation *Conversation) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*Conversation, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Conversation, error)
	Update(ctx context.Context, id uuidv7.UUID, title string) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	GetWithMessages(ctx context.Context, id uuidv7.UUID) (*Conversation, []*Message, error)
}

// MessageRepository defines the interface for message operations
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*Message, error)
	GetByID(ctx context.Context, id uuidv7.UUID) (*Message, error)
	Update(ctx context.Context, id uuidv7.UUID, content string) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	DeleteByConversationID(ctx context.Context, conversationID uuidv7.UUID) error
	GetCountByConversationID(ctx context.Context, conversationID uuidv7.UUID) (int, error)
}

// UsageRepository defines the interface for usage operations
type UsageRepository interface {
	Create(ctx context.Context, usage *Usage) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*Usage, error)
	GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*Usage, error)
	GetTotalByConversationID(ctx context.Context, conversationID uuidv7.UUID) (*Usage, error)
	Update(ctx context.Context, id uuidv7.UUID, usage *Usage) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	GetUsageSummary(ctx context.Context, limit, offset int) ([]*Usage, error)
}

// RepositoryManager provides access to all repositories
type RepositoryManager interface {
	Conversations() ConversationRepository
	Messages() MessageRepository
	Usage() UsageRepository
	Calls() CallRepository
	BeginTx(ctx context.Context) (RepositoryManager, error)
	Commit() error
	Rollback() error
	SaveProviderSettings(ctx context.Context, settings *ProviderSettings) error
	LoadProviderSettings(ctx context.Context, providerName string) (*ProviderSettings, error)
	SetDefaultProvider(ctx context.Context, providerName string) error
}

// repositoryManager implements RepositoryManager
type repositoryManager struct {
	db *DB
	tx *sql.Tx
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *DB) RepositoryManager {
	return &repositoryManager{
		db: db,
	}
}

func (rm *repositoryManager) Conversations() ConversationRepository {
	if rm.tx != nil {
		return &conversationRepository{tx: rm.tx}
	}
	return &conversationRepository{db: rm.db.DB}
}

func (rm *repositoryManager) Messages() MessageRepository {
	if rm.tx != nil {
		return &messageRepository{tx: rm.tx}
	}
	return &messageRepository{db: rm.db.DB}
}

func (rm *repositoryManager) Usage() UsageRepository {
    if rm.tx != nil {
        return &usageRepository{tx: rm.tx}
    }
    return &usageRepository{db: rm.db.DB}
}

func (rm *repositoryManager) Calls() CallRepository {
    if rm.tx != nil {
        return &callRepository{tx: rm.tx}
    }
    return &callRepository{db: rm.db.DB}
}

func (rm *repositoryManager) BeginTx(ctx context.Context) (RepositoryManager, error) {
	tx, err := rm.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	
	return &repositoryManager{
		db: rm.db,
		tx: tx,
	}, nil
}

func (rm *repositoryManager) Commit() error {
	if rm.tx != nil {
		return rm.tx.Commit()
	}
	return nil
}

func (rm *repositoryManager) Rollback() error {
	if rm.tx != nil {
		return rm.tx.Rollback()
	}
	return nil
}

func (rm *repositoryManager) SaveProviderSettings(ctx context.Context, settings *ProviderSettings) error {
	return rm.db.SaveProviderSettings(ctx, settings)
}

func (rm *repositoryManager) LoadProviderSettings(ctx context.Context, providerName string) (*ProviderSettings, error) {
	return rm.db.LoadProviderSettings(ctx, providerName)
}

func (rm *repositoryManager) SetDefaultProvider(ctx context.Context, providerName string) error {
	return rm.db.SetDefaultProvider(ctx, providerName)
}