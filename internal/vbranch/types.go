package vbranch

import "time"

// VirtualBranch represents a virtual branch with its changes
type VirtualBranch struct {
	Name         string            `json:"name"`
	ID           string            `json:"id"`
	Files        map[string]string `json:"files"`         // filename -> content for added/modified files
	DeletedFiles []string          `json:"deleted_files"` // list of deleted files
	Hunks        []Hunk            `json:"hunks"`         // individual change hunks
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Description  string            `json:"description"`
	Active       bool              `json:"active"`
}

// Hunk represents an individual change that can be moved between branches
type Hunk struct {
	ID        string    `json:"id"`         // Unique identifier for the hunk
	File      string    `json:"file"`       // Path to the file this hunk affects
	StartLine int       `json:"start_line"` // Starting line number in the file
	EndLine   int       `json:"end_line"`   // Ending line number in the file
	Content   string    `json:"content"`    // The actual content of the change
	Type      string    `json:"type"`       // "add", "remove", "modify"
	Context   string    `json:"context"`    // Surrounding lines for context
	CreatedAt time.Time `json:"created_at"` // When this hunk was created
}

// StickState manages the overall state of virtual branches
type StickState struct {
	Branches      map[string]*VirtualBranch `json:"branches"`
	CurrentBranch string                    `json:"current_branch"`
	WorkingDir    string                    `json:"working_dir"`
	GitRoot       string                    `json:"git_root"`
	LastSync      time.Time                 `json:"last_sync"`
}
