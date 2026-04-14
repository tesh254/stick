package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/dombox/uuidv7"
)

type Conversation struct {
	ID              uuidv7.UUID `json:"id"`
	Title           string      `json:"title"`
	WorkingDirectory string     `json:"working_directory"`
	CreatedAt       time.Time   `json:"created_at"`
}

type Message struct {
    ID           uuidv7.UUID `json:"id"`
    Conversation uuidv7.UUID `json:"conversation"`
    Content      string      `json:"content"`
    Role         Role        `json:"role"`
    CreatedAt    time.Time   `json:"created_at"`
    // Seq maintains per-conversation ordering for fast reconstruction.
    Seq          int64       `json:"seq"`
    // ParentMessage optionally links replies or assistant outputs to a triggering message.
    ParentMessage *uuidv7.UUID `json:"parent_message,omitempty"`
}

type Usage struct {
	ID               uuidv7.UUID `json:"id"`
	PromptTokens     int64       `json:"prompt_tokens"`
	CompletionTokens int64       `json:"completion_tokens"`
	TotalTokens      int64       `json:"total_tokens"`
	Model            string      `json:"model"`
	Conversation     uuidv7.UUID `json:"conversation"`
	CreatedAt        time.Time   `json:"created_at"`
}

type Role string

const (
	User      Role = "user"
	Assistant Role = "assistant"
	System    Role = "system"
)

// String returns the canonical lowercase string.
func (r Role) String() string {
	return string(r)
}

// Validate checks if r is one of the allowed roles.
func (r Role) Validate() error {
	switch r {
	case User, Assistant, System:
		return nil
	default:
		return fmt.Errorf("invalid role: %q", r)
	}
}

// FromString parses a role string (case-insensitive) into a Role.
func FromString(s string) (Role, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "user":
		return User, nil
	case "assistant":
		return Assistant, nil
	case "system":
		return System, nil
	default:
		return "", fmt.Errorf("invalid role: %q", s)
	}
}

// CallType defines the kind of execution recorded.
type CallType string

const (
    CallTypeFunction CallType = "function"
    CallTypeTool     CallType = "tool"
)

// CallStatus records execution outcomes.
type CallStatus string

const (
    CallStatusPending CallStatus = "pending"
    CallStatusSuccess CallStatus = "success"
    CallStatusError   CallStatus = "error"
)

// CallEvent captures function/tool invocation metadata for reconstruction and replay.
type CallEvent struct {
    ID              uuidv7.UUID  `json:"id"`
    Conversation    uuidv7.UUID  `json:"conversation"`
    ParentMessage   *uuidv7.UUID `json:"parent_message,omitempty"`
    Type            CallType     `json:"type"`
    Name            string       `json:"name"`
    ParamsJSON      string       `json:"params_json"`
    Status          CallStatus   `json:"status"`
    ResultRaw       string       `json:"result_raw"`
    Error           string       `json:"error"`
    StartedAt       time.Time    `json:"started_at"`
    CompletedAt     time.Time    `json:"completed_at"`
    DurationMS      int64        `json:"duration_ms"`
}

// ProviderSettings stores configuration for an AI provider
type ProviderSettings struct {
	ProviderName string    `json:"provider_name"`
	APIKey       string    `json:"api_key"`
	Model        string    `json:"model"`
	Endpoint     string    `json:"endpoint"`
	ExtraParams  string    `json:"extra_params"` // JSON string
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
