package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/functions"
	"github.com/tesh254/stick/internal/utils"
)

const gap = "\n\n"

type errMsg error

// handleWindowSizeMsg adjusts the layout when the terminal window is resized
func (m *model) handleWindowSizeMsg(v tea.WindowSizeMsg) {
	// Fullscreen: use the whole terminal. Viewport height is terminal minus input + gap.
	m.viewport.Width = v.Width
	m.textarea.SetWidth(v.Width)
	m.viewport.Height = v.Height - m.textarea.Height() - lipgloss.Height(gap)

	// Update wrap style to current viewport width
	m.wrapStyle = lipgloss.NewStyle().Width(m.viewport.Width)

	if len(m.messages) > 0 {
		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
	}
	m.viewport.GotoBottom()
}

// handleKeyMsg processes keyboard inputs
func (m *model) handleKeyMsg(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch v.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		// Print last textarea value then quit
		fmt.Println(m.textarea.Value())
		return *m, tea.Quit
	case tea.KeyEnter:
		m.handleEnterKey()
	}

	return *m, nil
}

// handleEnterKey processes the Enter key press in the textarea
func (m *model) handleEnterKey() {
	input := m.textarea.Value()
	if input != "" {
		username := utils.GetUser()
		prefix := m.senderStyle.Render("{" + username + "}: ")

		// Check if input looks like a function call by trying to parse it
		result, err := m.processFunctionCall(input)

		if err != nil {
			// Function call failed, display the input and error
			m.messages = append(m.messages, prefix+input)
			m.messages = append(m.messages, "Error: "+err.Error())
		} else if result != "" {
			// Function call succeeded, display input and result
			m.messages = append(m.messages, prefix+input)
			m.messages = append(m.messages, "Function call result: "+result)
		} else {
			// Regular message, not a function call
			m.messages = append(m.messages, prefix+input)
		}

		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
		m.textarea.Reset()
		m.viewport.GotoBottom()
	}
}

// processFunctionCall processes function calls in the input string
func (m *model) processFunctionCall(input string) (string, error) {
	p := functions.Parser{}
	name, args, err := p.Parse(input)

	// If parsing succeeds and no error, it's a function call
	if err == nil && name != "" {
		// Attempt to call the function
		return m.functionRegistry.Call(name, args)
	}

	// Not a function call, return empty string and no error
	return "", nil
}

// formatMessages joins messages with newlines for display in the viewport
func formatMessages(messages []string) string {
	return strings.Join(messages, "\n")
}