package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/functions"
)

// initialModel creates the initial state of the TUI application
func initialModel() model {
	ta := setupTextarea()
	vp := setupViewport()
	wrap := setupWrapStyle(vp)
	registry := setupFunctionRegistry()

	return model{
		textarea:         ta,
		messages:         []string{},
		commandHistory:   []string{},
		historyIndex:     -1,
		viewport:         vp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		wrapStyle:        wrap,
		err:              nil,
		functionRegistry: registry,
		showSearchModal:  false,
		searchInput:      "",
		allSlashCommands: []string{},
		filteredCommands: []string{},
		selectedIndex:    0,
		isInSlashMode:    false,
		viewportFocused:  false,
	}
}

// setupTextarea configures the text input area
func setupTextarea() textarea.Model {
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

	return ta
}

// setupViewport configures the message display area
func setupViewport() viewport.Model {
	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to stick!
Type a message and press Enter to send.`)

	return vp
}

// setupWrapStyle creates a style for wrapping content based on viewport width
func setupWrapStyle(vp viewport.Model) lipgloss.Style {
	return lipgloss.NewStyle().Width(vp.Width)
}

// setupFunctionRegistry initializes and registers functions in the registry
func setupFunctionRegistry() *functions.Registry {
	registry := functions.NewRegistry()
	registry.Register("add", functions.Add, 0, 2)
	registry.Register("echo", functions.Echo, 0, -1) // -1 means unlimited arguments
	registry.Register("print_statement", functions.Echo, 0, -1)

	// crawl functions
	registry.Register("get_llm_text", functions.GetLLMText, 1, 1)
	registry.Register("get_page_content", functions.GetPageHTMLContentToMarkdown, 1, 1)

	return registry
}

// Init initializes the TUI model
func (m model) Init() tea.Cmd {
	return textarea.Blink
}
