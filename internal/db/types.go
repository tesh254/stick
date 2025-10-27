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
