package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dombox/uuidv7"
	"github.com/tesh254/stick/internal/agent"
	"github.com/tesh254/stick/internal/db"
	"github.com/tesh254/stick/internal/functions"
)

// initialModel creates the initial state of the TUI application
func initialModel() model {
	ta := setupTextarea()
	vp := setupViewport()
	wrap := setupWrapStyle(vp)
	registry := setupFunctionRegistry()
	dbConn, rm, convID, storageQueue, callQueue := setupDBAndConversation()

	agentSession := agent.NewSession()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		textarea:         ta,
		messages:         []string{},
		storageMessages:  []*db.Message{},
		commandHistory:   []string{},
		historyIndex:     -1,
		viewport:         vp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		wrapStyle:        wrap,
		err:              nil,
		functionRegistry: registry,
		dbConn:           dbConn,
		repoManager:      rm,
		conversationID:   convID,
		storageQueue:     storageQueue,
		callStorageQueue: callQueue,
		showSearchModal:  false,
		searchInput:      "",
		allSlashCommands: []string{},
		filteredCommands: []string{},
		selectedIndex:    0,
		isInSlashMode:    false,
		showFileModal:    false,
		fileSearchInput:  "",
		allFiles:         []string{},
		filteredFiles:    []string{},
		fileSelectedIndex: 0,
		isInAtMode:       false,
		viewportFocused:  false,
		agentSession:     agentSession,
		askUserModel:     InitialAskUserModel(),
		spinner:          s,
		isLoading:        false,
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
	vp.SetContent(fmt.Sprintf("%s\n Type a message and press Enter to send.", AGENT_ASCII))

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
	registry.Register("set_provider", functions.SetProvider, 3, 3)

	return registry
}

// Init initializes the TUI model
func (m model) Init() tea.Cmd {
	// Start background storage worker to persist messages without blocking UI.
	startStorageWorker(&m)
	// Start background storage worker to persist call events and metadata.
	startCallStorageWorker(&m)
	return textarea.Blink
}

// setupDBAndConversation initializes the DB connection and creates a conversation for this session.
func setupDBAndConversation() (*db.DB, db.RepositoryManager, uuidv7.UUID, chan *db.Message, chan *db.CallEvent) {
	// Open local sqlite DB
	dbConn, err := db.New()
	if err != nil {
		// In case of DB failure, return nils; the UI will still function without persistence.
		return nil, nil, uuidv7.UUID{}, make(chan *db.Message), make(chan *db.CallEvent)
	}
	rm := db.NewRepositoryManager(dbConn)

	// Create a new conversation record
	convID, _ := uuidv7.New()
	wd, _ := os.Getwd()
	conv := &db.Conversation{
		ID:               convID,
		Title:            fmt.Sprintf("TUI Session %s", time.Now().Format("2006-01-02 15:04:05")),
		WorkingDirectory: wd,
		CreatedAt:        time.Now(),
	}
	// Best-effort create; if it fails, we keep going but will skip storage.
	_ = rm.Conversations().Create(context.Background(), conv)

	// Buffered queue to avoid blocking UI
	storageQueue := make(chan *db.Message, 256)
	callQueue := make(chan *db.CallEvent, 256)
	return dbConn, rm, convID, storageQueue, callQueue
}
