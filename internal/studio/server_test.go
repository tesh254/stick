package studio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dombox/uuidv7"
	"github.com/gofiber/fiber/v2"
	"github.com/tesh254/stick/internal/db"
)

// --- Mocks ---

type mockRepoManager struct{}

func (m *mockRepoManager) Conversations() db.ConversationRepository { return &mockConvRepo{} }
func (m *mockRepoManager) Messages() db.MessageRepository           { return &mockMsgRepo{} }
func (m *mockRepoManager) Usage() db.UsageRepository                { return &mockUsageRepo{} }
func (m *mockRepoManager) Calls() db.CallRepository                 { return &mockCallRepo{} }
func (m *mockRepoManager) LoadProviderSettings(_ context.Context, _ string) (*db.ProviderSettings, error) {
	return nil, nil
}
func (m *mockRepoManager) SaveProviderSettings(_ context.Context, _ *db.ProviderSettings) error {
	return nil
}
func (m *mockRepoManager) SetDefaultProvider(_ context.Context, _ string) error    { return nil }
func (m *mockRepoManager) BeginTx(_ context.Context) (db.RepositoryManager, error) { return m, nil }
func (m *mockRepoManager) Commit() error                                           { return nil }
func (m *mockRepoManager) Rollback() error                                         { return nil }

type mockConvRepo struct{}

func (r *mockConvRepo) Create(_ context.Context, _ *db.Conversation) error { return nil }
func (r *mockConvRepo) GetByID(_ context.Context, id uuidv7.UUID) (*db.Conversation, error) {
	return &db.Conversation{ID: id, Title: "Test", WorkingDirectory: "/tmp", CreatedAt: time.Now()}, nil
}
func (r *mockConvRepo) GetAll(_ context.Context, limit, offset int) ([]*db.Conversation, error) {
	uid, err := uuidv7.New()
	if err != nil {
		uid = uuidv7.UUID{}
	}
	return []*db.Conversation{{ID: uid, Title: "A", WorkingDirectory: "/tmp", CreatedAt: time.Now()}}, nil
}
func (r *mockConvRepo) Update(_ context.Context, _ uuidv7.UUID, _ string) error { return nil }
func (r *mockConvRepo) Delete(_ context.Context, _ uuidv7.UUID) error           { return nil }
func (r *mockConvRepo) GetWithMessages(_ context.Context, id uuidv7.UUID) (*db.Conversation, []*db.Message, error) {
	return &db.Conversation{ID: id, Title: "Test", WorkingDirectory: "/tmp", CreatedAt: time.Now()}, []*db.Message{}, nil
}

type mockMsgRepo struct{}

func (r *mockMsgRepo) Create(_ context.Context, _ *db.Message) error { return nil }
func (r *mockMsgRepo) GetByConversationID(_ context.Context, _ uuidv7.UUID) ([]*db.Message, error) {
	return []*db.Message{}, nil
}
func (r *mockMsgRepo) GetByID(_ context.Context, _ uuidv7.UUID) (*db.Message, error) {
	return &db.Message{}, nil
}
func (r *mockMsgRepo) Update(_ context.Context, _ uuidv7.UUID, _ string) error       { return nil }
func (r *mockMsgRepo) Delete(_ context.Context, _ uuidv7.UUID) error                 { return nil }
func (r *mockMsgRepo) DeleteByConversationID(_ context.Context, _ uuidv7.UUID) error { return nil }
func (r *mockMsgRepo) GetCountByConversationID(_ context.Context, _ uuidv7.UUID) (int, error) {
	return 0, nil
}

type mockUsageRepo struct{}

func (r *mockUsageRepo) Create(_ context.Context, _ *db.Usage) error { return nil }
func (r *mockUsageRepo) GetByID(_ context.Context, _ uuidv7.UUID) (*db.Usage, error) {
	return &db.Usage{}, nil
}
func (r *mockUsageRepo) GetByConversationID(_ context.Context, _ uuidv7.UUID) ([]*db.Usage, error) {
	return []*db.Usage{}, nil
}
func (r *mockUsageRepo) GetTotalByConversationID(_ context.Context, _ uuidv7.UUID) (*db.Usage, error) {
	return &db.Usage{}, nil
}
func (r *mockUsageRepo) Update(_ context.Context, _ uuidv7.UUID, _ *db.Usage) error { return nil }
func (r *mockUsageRepo) Delete(_ context.Context, _ uuidv7.UUID) error              { return nil }
func (r *mockUsageRepo) GetUsageSummary(_ context.Context, limit, offset int) ([]*db.Usage, error) {
	return []*db.Usage{}, nil
}

type mockCallRepo struct{}

func (r *mockCallRepo) Create(_ context.Context, _ *db.CallEvent) error { return nil }
func (r *mockCallRepo) GetByConversationID(_ context.Context, _ uuidv7.UUID) ([]*db.CallEvent, error) {
	return []*db.CallEvent{}, nil
}
func (r *mockCallRepo) GetByParentMessageID(_ context.Context, _ uuidv7.UUID) ([]*db.CallEvent, error) {
	return []*db.CallEvent{}, nil
}
func (r *mockCallRepo) UpdateStatus(_ context.Context, _ uuidv7.UUID, _ db.CallStatus, _, _ string, _ time.Time, _ int64) error {
	return nil
}
func (r *mockCallRepo) DeleteByConversationID(_ context.Context, _ uuidv7.UUID) error { return nil }

func TestHealthEndpoint(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	registerRoutes(app, Config{Env: "test"}, func() (db.RepositoryManager, error) { return &mockRepoManager{}, nil }, DefaultFuncRegistryFactory())

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFunctionsList(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	registerRoutes(app, Config{Env: "test"}, func() (db.RepositoryManager, error) { return &mockRepoManager{}, nil }, DefaultFuncRegistryFactory())

	req := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("functions list request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
