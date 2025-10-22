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
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	wrapStyle        lipgloss.Style
	err              error
	functionRegistry *functions.Registry
}

// NewProgram creates a Bubble Tea program configured for fullscreen (AltScreen).
// Call this instead of tea.NewProgram(initialModel()) in your main.
func NewProgram() *tea.Program {
	return tea.NewProgram(initialModel(), tea.WithAltScreen())
}