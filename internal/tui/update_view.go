package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(v)
	case tea.KeyMsg:
		return m.handleKeyMsg(v)
	case errMsg:
		m.err = v
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// View renders the UI components
func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}