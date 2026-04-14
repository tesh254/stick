// package tui/model.go
package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dombox/uuidv7"
	"github.com/tesh254/stick/internal/agent"
	"github.com/tesh254/stick/internal/db"
	"github.com/tesh254/stick/internal/functions"
	"github.com/tesh254/stick/internal/provider"
)

// model represents the state of the TUI application
type model struct {
	viewport viewport.Model
	messages []string
	// storageMessages holds the canonical messages intended for DB persistence.
	// This is kept in lockstep with the display slice to ensure consistency.
	storageMessages  []*db.Message
	commandHistory   []string
	historyIndex     int
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	wrapStyle        lipgloss.Style
	err              error
	functionRegistry *functions.Registry
	// DB integration fields (non-blocking writes via background worker).
	dbConn           *db.DB
	repoManager      db.RepositoryManager
	conversationID   uuidv7.UUID
	storageQueue     chan *db.Message
	callStorageQueue chan *db.CallEvent
	// Search modal fields
	showSearchModal  bool
	searchInput      string
	allSlashCommands []string // Full list of commands for filtering
	filteredCommands []string // Filtered list displayed in modal
	selectedIndex    int
	// Auto-slash modal behavior
	isInSlashMode bool // Tracks if we're in slash command mode
	// File search modal fields
	showFileModal     bool
	fileSearchInput   string
	allFiles          []string
	filteredFiles     []string
	fileSelectedIndex int
	isInAtMode        bool

	// Viewport focus state
	viewportFocused bool // Tracks if the viewport has focus for scrolling
	isAskUserMode   bool
	askUserModel    AskUserModel
	askUserQuestion string

	// Agent
	agentSession    *agent.Session
	pendingToolCall *provider.ToolCall // To store the tool call waiting for confirmation

	// Loading state
	isLoading bool
	spinner   spinner.Model
}

// NewProgram creates a Bubble Tea program configured for fullscreen (AltScreen).
// Call this instead of tea.NewProgram(initialModel()) in your main.
func NewProgram() *tea.Program {
	return tea.NewProgram(initialModel(), tea.WithAltScreen())
}
