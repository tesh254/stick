// package tui/model.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/utils"
)

const gap = "\n\n"

type (
	errMsg error
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	wrapStyle   lipgloss.Style
	err         error
}

// NewProgram creates a Bubble Tea program configured for fullscreen (AltScreen).
// Call this instead of tea.NewProgram(initialModel()) in your main.
func NewProgram() *tea.Program {
	return tea.NewProgram(initialModel(), tea.WithAltScreen())
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "

	// Remove character limit so pasted long text isn't truncated.
	ta.CharLimit = 0

	// Initial sizes; will be overridden by WindowSizeMsg on start
	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter sends, not newline

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to stick!
Type a message and press Enter to send.`)

	wrap := lipgloss.NewStyle().Width(vp.Width)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		wrapStyle:   wrap,
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	username := utils.GetUser()

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		// Fullscreen: use the whole terminal. Viewport height is terminal minus input + gap.
		m.viewport.Width = v.Width
		m.textarea.SetWidth(v.Width)
		m.viewport.Height = v.Height - m.textarea.Height() - lipgloss.Height(gap)

		// Update wrap style to current viewport width
		m.wrapStyle = lipgloss.NewStyle().Width(m.viewport.Width)

		if len(m.messages) > 0 {
			m.viewport.SetContent(m.wrapStyle.Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch v.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Print last textarea value then quit
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			// Build label first; Render takes a single string
			prefix := m.senderStyle.Render("{" + username + "}: ")
			m.messages = append(m.messages, prefix+m.textarea.Value())
			m.viewport.SetContent(m.wrapStyle.Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = v
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
