// package tui/model.go
package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/functions"
)

// model represents the state of the TUI application
type model struct {
	viewport         viewport.Model
	messages         []string
	commandHistory   []string
	historyIndex     int
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	wrapStyle        lipgloss.Style
	err              error
	functionRegistry *functions.Registry
	// Search modal fields
	showSearchModal  bool
	searchInput      string
	allSlashCommands []string // Full list of commands for filtering
	filteredCommands []string // Filtered list displayed in modal
	selectedIndex    int
	// Auto-slash modal behavior
	isInSlashMode bool // Tracks if we're in slash command mode
}

// NewProgram creates a Bubble Tea program configured for fullscreen (AltScreen).
// Call this instead of tea.NewProgram(initialModel()) in your main.
func NewProgram() *tea.Program {
	return tea.NewProgram(initialModel(), tea.WithAltScreen())
}
