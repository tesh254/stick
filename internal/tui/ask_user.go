// package tui/ask_user.go
package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const askListHeight = 5 // Smaller height for simple yes/no

var (
	askTitleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	askItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	askSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	askPaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	askHelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	askQuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type askItem string

func (i askItem) FilterValue() string { return "" }

type askItemDelegate struct{}

func (d askItemDelegate) Height() int                               { return 1 }
func (d askItemDelegate) Spacing() int                              { return 0 }
func (d askItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d askItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(askItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := askItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return askSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// AskUserModel is a Bubble Tea model for yes/no confirmation
type AskUserModel struct {
	list     list.Model
	choice   string
	quitting bool
	prompt   string // The question to ask the user
	diff     string // Optional diff or content to show
}

func NewAskUserModel(prompt string, diff string) AskUserModel {
	items := []list.Item{
		askItem("Yes"),
		askItem("No"),
	}

	const defaultWidth = 60 // Increased width for diff visibility

	l := list.New(items, askItemDelegate{}, defaultWidth, askListHeight)
	l.Title = prompt // Use the provided prompt as title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = askTitleStyle
	l.Styles.PaginationStyle = askPaginationStyle
	l.Styles.HelpStyle = askHelpStyle

	return AskUserModel{list: l}
}

func InitialAskUserModel() AskUserModel {
	return NewAskUserModel("", "")
}

func (m AskUserModel) Init() tea.Cmd {
	return nil
}

func (m AskUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			m.choice = "No" // Default to No on quit
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(askItem)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AskUserModel) View() string {
	if m.choice != "" {
		return askQuitTextStyle.Render(fmt.Sprintf("You selected: %s", m.choice))
	}
	if m.quitting {
		return askQuitTextStyle.Render("Action cancelled.")
	}
	view := ""
	if m.diff != "" {
		view += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(m.diff) + "\n\n"
	}
	view += m.list.View()
	return "\n" + view
}

// AskUser runs the yes/no prompt and returns the user's choice ("Yes", "No", or "" on error)
func AskUser(prompt string) (string, error) {
	m := NewAskUserModel(prompt, "")
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	return finalModel.(AskUserModel).choice, nil
}
