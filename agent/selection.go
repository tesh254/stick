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

type providerInfo struct {
	id          string
	name        string
	description string
}

type modelInfo struct {
	id           string
	name         string
	inputCostPM  float32
	outputCostPM float64
	context      int
	description  string
}

var providers = []providerInfo{
	{id: "together", name: "Together AI", description: "A cloud provider for running, fine-tuning, and serving large AI models."},
	{id: "anthropic", name: "Anthropic", description: "AI safety and research company creating reliable, interpretable, and steerable AI systems."},
	{id: "google", name: "Google", description: "A suite of generative AI models developed by Google."},
	{id: "openai", name: "OpenAI", description: "An AI research and deployment company."},
	{id: "openrouter", name: "OpenRouter", description: "A unified API for accessing and using various large language models."},
	{id: "atlascloud", name: "Atlas Cloud", description: "A platform for running and fine-tuning AI models."},
}

var models = map[string][]modelInfo{
	"together": {
		{id: "Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8", name: "Qwen3 Coder", inputCostPM: 0.2, outputCostPM: 0.80, context: 262144, description: "Optimised for agentic coding tasks."},
	},
	"anthropic": {
		{id: "claude-3-opus-20240229", name: "Claude 3 Opus", inputCostPM: 15.00, outputCostPM: 75.00, context: 200000, description: "Most powerful model for highly complex tasks."},
	},
	"google": {
		{id: "gemini/gemini-1.5-pro-latest", name: "Gemini 1.5 Pro", inputCostPM: 7.00, outputCostPM: 21.00, context: 1000000, description: "Most capable generative model."},
	},
	"openai": {
		{id: "gpt-4-turbo", name: "GPT-4 Turbo", inputCostPM: 10.00, outputCostPM: 30.00, context: 128000, description: "The latest GPT-4 model with improved instruction-following."},
	},
	"openrouter": {
		{id: "openrouter/auto", name: "Auto (best-in-class)", inputCostPM: 1.00, outputCostPM: 1.00, context: 128000, description: "Automatically selects the best model for your request."},
	},
	"atlascloud": {
		{id: "atlascloud/llama-3-8b-instruct", name: "Llama 3 8B Instruct", inputCostPM: 0.2, outputCostPM: 0.2, context: 8192, description: "The most capable small model from Meta."},
	},
}

var (
	appStyle           = lipgloss.NewStyle().Padding(1, 2)
	titleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})
)

func (i providerInfo) FilterValue() string {
	return i.name
}

func (i providerInfo) Title() string {
	return i.name
}

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
	switch item := listItem.(type) {
	case providerInfo:
		title := item.name
		desc := item.description

		if index == m.Index() {
			title = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render(title)
		}

		itemStr := fmt.Sprintf("%s\n  %s", title, desc)

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
	case modelInfo:
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
}

type sessionState int

const (
	providerSelection sessionState = iota
	modelSelection
	apiKeyInput
)

type model struct {
	list             list.Model
	input            textinput.Model
	state            sessionState
	selectedProvider providerInfo
	selectedModel    modelInfo
	quitting         bool
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
			case providerSelection:
				i, ok := m.list.SelectedItem().(providerInfo)
				if ok {
					m.selectedProvider = i
					m.state = modelSelection
					items := make([]list.Item, len(models[i.id]))
					for j, mi := range models[i.id] {
						items[j] = mi
					}
					m.list.SetItems(items)
				}
			case modelSelection:
				i, ok := m.list.SelectedItem().(modelInfo)
				if ok {
					m.selectedModel = i
					m.state = apiKeyInput
					m.input.Focus()
					return m, textinput.Blink
				}
			case apiKeyInput:
				apiKey := m.input.Value()
				if err := config.Set(fmt.Sprintf("providers.%s.apiKey", m.selectedProvider.id), apiKey); err != nil {
					m.quitting = true
					return m, tea.Quit
				}
				if err := config.Set(fmt.Sprintf("providers.%s.model", m.selectedProvider.id), m.selectedModel.id); err != nil {
					m.quitting = true
					return m, tea.Quit
				}
				if err := config.Set("defaultProvider", m.selectedProvider.id); err != nil {
					m.quitting = true
					return m, tea.Quit
				}
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case apiKeyInput:
		return appStyle.Render(fmt.Sprintf(
			"Enter your %s API Key for %s: \n%s \n%s",
			m.selectedProvider.name,
			m.selectedModel.name,
			m.input.View(),
			statusMessageStyle.Render("(press enter to save)"),
		))
	case modelSelection:
		return m.list.View()
	default:
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
		asciiArt := green.Render(constants.STICK_ASCII)

		availableProvidersStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#04B575")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginBottom(1)
		availableProvidersLabel := availableProvidersStyle.Render("Available Providers")

		return appStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				asciiArt,
				availableProvidersLabel,
				m.list.View(),
			),
		)
	}
}

func SelectModel() {
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = p
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle().Margin(1, 0)
	l.Styles.HelpStyle = lipgloss.NewStyle().Margin(1, 0)

	ti := textinput.New()
	ti.Placeholder = "sk-..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	m := &model{
		list:  l,
		input: ti,
		state: providerSelection,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(*model); ok {
		if fm.selectedModel.id != "" {
			fmt.Printf("%s ", statusMessageStyle.Render(fmt.Sprintf("✓ Selected model set to: %s", fm.selectedModel.name)))
		}
	}
}
