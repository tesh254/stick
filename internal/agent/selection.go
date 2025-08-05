package agent

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/config"
	"github.com/tesh254/stick/internal/constants"
)

type modelInfo struct {
	id           string
	name         string
	inputCostPM  float32
	outputCostPM float64
	context      int
	description  string
}

var models = []modelInfo{
	{
		id:           "qwen/qwen3-coder:free",
		name:         "Qwen3 Coder (Free)",
		inputCostPM:  0,
		outputCostPM: 0,
		context:      262144,
		description:  `optimised for agentic coding tasks such as function calling, tool use and longer context reasoning over repos`,
	},
	{
		id:           "qwen/qwen3-coder",
		name:         "Qwen3 Coder",
		inputCostPM:  0.30,
		outputCostPM: 1.20,
		context:      262144,
		description:  `optimised for agentic coding tasks such as function calling, tool use and longer context reasoning over repos`,
	},
	{
		id:           "google/gemini-2.0-flash-exp:free",
		name:         "Google: Gemini 2.0 Flash Experimental (free)",
		inputCostPM:  0,
		outputCostPM: 0,
		context:      1048576,
		description:  `faster time to first token and delivers more seamless and robust agentic experience`,
	},
	{
		id:           "openrouter/horizon-beta",
		name:         "OpenRouter: Horizon Beta",
		inputCostPM:  0,
		outputCostPM: 0,
		context:      256000,
		description:  `faster time to first token and delivers more seamless and robust agentic experience`,
	},
}

var (
	appStyle           = lipgloss.NewStyle().Padding(1, 2)
	titleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})
)

func (i modelInfo) Name() string {
	return i.name
}

func (i modelInfo) ID() string {
	return i.id
}

func (i modelInfo) Description() string {
	return i.description
}

func (i modelInfo) InputCostPerMillion() float32 {
	return i.inputCostPM
}

func (i modelInfo) OutputCostPerMillion() float64 {
	return i.outputCostPM
}

func (i modelInfo) FilterValue() string {
	return i.id
}

func (i modelInfo) Title() string {
	return i.name
}

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 3
}

func (d itemDelegate) Spacing() int {
	return 1
}

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	mi, ok := listItem.(modelInfo)
	if !ok {
		return
	}

	title := mi.name
	desc := mi.description
	costs := fmt.Sprintf("Input: $%.2f/M, Output: $%.2f/M", mi.inputCostPM, mi.outputCostPM)
	context := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("Context: %d", mi.context))

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render(title)
	}

	itemStr := fmt.Sprintf("%s\n  %s\n  %s | %s", title, desc, costs, context)

	if index == m.Index() {
		style := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#04B575")).
			PaddingLeft(2)
		fmt.Fprint(w, style.Render(itemStr))
	} else {
		style := lipgloss.NewStyle().PaddingLeft(3)
		fmt.Fprint(w, style.Render(itemStr))
	}
}

type sessionState int

const (
	modelSelection sessionState = iota
	apiKeyInput
)

type model struct {
	list         list.Model
	input        textinput.Model
	state        sessionState
	selectedItem modelInfo
	quitting     bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			switch m.state {
			case modelSelection:
				i, ok := m.list.SelectedItem().(modelInfo)
				if ok {
					m.selectedItem = i
					m.state = apiKeyInput
					m.input.Focus()
					return m, textinput.Blink
				}
			case apiKeyInput:
				apiKey := m.input.Value()
				if err := config.SetAPIKey(apiKey); err != nil {
					// In a real app, you might want to display this error to the user.
					// For now, we'll just quit.
					m.quitting = true
					return m, tea.Quit
				}
				if err := config.SetSelectedModel(m.selectedItem.id); err != nil {
					m.quitting = true
					return m, tea.Quit
				}
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case modelSelection:
		m.list, cmd = m.list.Update(msg)
	case apiKeyInput:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case apiKeyInput:
		return appStyle.Render(fmt.Sprintf(
			"Enter your OpenRouter API Key for %s: \n%s \n%s",
			m.selectedItem.name,
			m.input.View(),
			statusMessageStyle.Render("(press enter to save)"),
		))
	default:
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
		asciiArt := green.Render(constants.STICK_ASCII)

		availableModelsStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#04B575")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginBottom(1)
		availableModelsLabel := availableModelsStyle.Render("Available Models")

		return appStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				asciiArt,
				availableModelsLabel,
				m.list.View(),
			),
		)
	}
}

func SelectModel() {
	items := make([]list.Item, len(models))
	for i, mi := range models {
		items[i] = mi
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle().Margin(1, 0)
	l.Styles.HelpStyle = lipgloss.NewStyle().Margin(1, 0)

	ti := textinput.New()
	ti.Placeholder = "sk-or-..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	m := &model{
		list:  l,
		input: ti,
		state: modelSelection,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(*model); ok {
		if fm.selectedItem.id != "" {
			fmt.Printf("%s ", statusMessageStyle.Render(fmt.Sprintf("✓ Selected model set to: %s", fm.selectedItem.name)))
		}
	}
}
