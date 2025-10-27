// Package dbtui contains TUI components for database interactions
package dbtui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/db"
)

// ConversationItem represents a conversation in the list
type ConversationItem struct {
	ID              string
	TitleValue      string
	WorkingDirectory string
	CreatedAt       string
}

// FilterValue implements the list.Item interface
func (c ConversationItem) FilterValue() string {
	return c.TitleValue
}

// Title implements the list.Item interface
func (c ConversationItem) Title() string {
	return c.TitleValue
}

// Description implements the list.Item interface
func (c ConversationItem) Description() string {
	return fmt.Sprintf("ID: %s | Directory: %s | Created: %s", c.ID[:8], c.WorkingDirectory, c.CreatedAt)
}

// ConversationListModel holds the state for the conversation list view
type ConversationListModel struct {
	list     list.Model
	db       *db.DB
	repo     db.ConversationRepository
	quitting bool
}

// NewConversationListModel creates a new conversation list model
func NewConversationListModel() (*ConversationListModel, error) {
	// Initialize database connection
	dbConn, err := db.New()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repoManager := db.NewRepositoryManager(dbConn)

	// Get conversations and convert to items
	conversations, err := repoManager.Conversations().GetAll(context.Background(), 50, 0)
	if err != nil {
		dbConn.Close() // Clean up on error
		return nil, fmt.Errorf("failed to retrieve conversations: %w", err)
	}

	// Convert conversations to list items
	var items []list.Item
	for _, conv := range conversations {
		item := ConversationItem{
			ID:              conv.ID.String(),
			TitleValue:      conv.Title,
			WorkingDirectory: conv.WorkingDirectory,
			CreatedAt:       conv.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		items = append(items, item)
	}

	// Create the Bubble Tea list model
	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Conversations"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	model := &ConversationListModel{
		list: l,
		db:   dbConn,
	}

	return model, nil
}

// Init initializes the model
func (m *ConversationListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *ConversationListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keypress := msg.String(); keypress == "ctrl+c" || keypress == "q" {
			m.quitting = true
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m *ConversationListModel) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	return m.list.View()
}

// Define styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true)
)

// Run starts the Bubble Tea program
func (m *ConversationListModel) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}