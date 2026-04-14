package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/db"
)

const (
	providerOpenAI      = "OpenAI"
	providerOpenRouter  = "OpenRouter"
	providerAnthropic   = "Anthropic"
	providerZAI         = "Z.AI"
	providerGrok        = "Grok.com"
)

var (
	// Styles
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#7D56F4"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

type SettingsModel struct {
	providerList    list.Model
	textInputs      []textinput.Model
	focusedInput    int
	selectedProvider string
	err             error
	quitting        bool
	repoManager     db.RepositoryManager
}

// NewSettingsModel creates a new settings model with repository manager for DB access
func NewSettingsModel(rm db.RepositoryManager) SettingsModel {
	// Provider list items
	items := []list.Item{
		item{title: providerOpenAI, desc: "Configure OpenAI provider"},
		item{title: providerOpenRouter, desc: "Configure OpenRouter provider"},
		item{title: providerAnthropic, desc: "Configure Anthropic (Claude) provider"},
		item{title: providerZAI, desc: "Configure Z.AI provider"},
		item{title: providerGrok, desc: "Configure Grok.com (xAI) provider"},
	}

	// Create the list
	providerList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	providerList.Title = "Select AI Provider"
	providerList.SetShowStatusBar(false)
	providerList.SetFilteringEnabled(false)
	providerList.Styles.Title = titleStyle

	// Text inputs for configuration
	inputs := make([]textinput.Model, 6)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "API Key"
	inputs[0].EchoMode = textinput.EchoPassword // Mask API key
	inputs[0].EchoCharacter = '•'

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Model (e.g., gpt-4o)"

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Endpoint (default: https://api.openai.com/v1)"

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Temperature (default: 1.0)"

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Max Tokens (default: 4096)"

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Set as default (y/n)"

	for i := range inputs {
		inputs[i].CharLimit = 256
		inputs[i].Width = 50
	}

	return SettingsModel{
		providerList: providerList,
		textInputs:   inputs,
		focusedInput: -1, // -1 means provider selection mode
		repoManager:  rm,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.focusedInput == -1 {
				// Selected a provider
				selectedItem := m.providerList.SelectedItem().(item)
				m.selectedProvider = selectedItem.title
				m.focusedInput = 0
				m.textInputs[0].Focus()
				// Set default values based on provider
				m.setDefaults()
				// Load existing settings if available
				m.loadProviderSettings()
			} else if m.focusedInput < len(m.textInputs)-1 {
				m.textInputs[m.focusedInput].Blur()
				m.focusedInput++
				m.textInputs[m.focusedInput].Focus()
			} else {
				// Save settings
				err := m.saveProviderSettings()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.quitting = true
				return m, tea.Quit
			}
		case "tab":
			if m.focusedInput >= 0 {
				m.textInputs[m.focusedInput].Blur()
				m.focusedInput = (m.focusedInput + 1) % len(m.textInputs)
				m.textInputs[m.focusedInput].Focus()
			}
		}
	case tea.WindowSizeMsg:
		m.providerList.SetWidth(msg.Width)
		m.providerList.SetHeight(msg.Height - 2)
		return m, nil
	}

	if m.focusedInput == -1 {
		var cmd tea.Cmd
		m.providerList, cmd = m.providerList.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.textInputs[m.focusedInput], cmd = m.textInputs[m.focusedInput].Update(msg)
		return m, cmd
	}
}

func (m SettingsModel) View() string {
	if m.quitting {
		return "Settings saved. Quitting...\n"
	}

	if m.focusedInput == -1 {
		return m.providerList.View()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Configure %s", m.selectedProvider)) + "\n\n")

	for i, input := range m.textInputs {
		if i == m.focusedInput {
			b.WriteString(input.View() + "\n")
		} else {
			b.WriteString(input.View() + "\n")
		}
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()) + "\n")
	}

	b.WriteString("\nPress enter to proceed, tab to switch fields, q to quit.\n")

	return b.String()
}

func (m *SettingsModel) setDefaults() {
	switch m.selectedProvider {
	case providerOpenAI:
		m.textInputs[1].SetValue("gpt-4o")
		m.textInputs[2].SetValue("https://api.openai.com/v1")
	case providerOpenRouter:
		m.textInputs[1].SetValue("mixtral-8x7b")
		m.textInputs[2].SetValue("https://openrouter.ai/api/v1")
	case providerAnthropic:
		m.textInputs[1].SetValue("claude-3-opus")
		m.textInputs[2].SetValue("https://api.anthropic.com/v1")
	case providerZAI:
		m.textInputs[1].SetValue("default-model")
		m.textInputs[2].SetValue("https://api.z.ai/v1")
	case providerGrok:
		m.textInputs[1].SetValue("grok-4")
		m.textInputs[2].SetValue("https://api.grok.com/v1")
	}
	m.textInputs[3].SetValue("1.0")
	m.textInputs[4].SetValue("4096")
	m.textInputs[5].SetValue("n")
}

func (m *SettingsModel) loadProviderSettings() {
	if m.repoManager == nil {
		return
	}
	ctx := context.Background()
	settings, err := m.repoManager.LoadProviderSettings(ctx, m.selectedProvider)
	if err != nil || settings == nil {
		return // No settings found, use defaults
	}

	m.textInputs[0].SetValue(settings.APIKey)
	m.textInputs[1].SetValue(settings.Model)
	m.textInputs[2].SetValue(settings.Endpoint)

	var extra map[string]interface{}
	if err := json.Unmarshal([]byte(settings.ExtraParams), &extra); err == nil {
		if temp, ok := extra["temperature"].(float64); ok {
			m.textInputs[3].SetValue(fmt.Sprintf("%f", temp))
		}
		if maxTokens, ok := extra["max_tokens"].(float64); ok {
			m.textInputs[4].SetValue(fmt.Sprintf("%d", int(maxTokens)))
		}
	}

	if settings.IsDefault {
		m.textInputs[5].SetValue("y")
	} else {
		m.textInputs[5].SetValue("n")
	}
}

func (m *SettingsModel) saveProviderSettings() error {
	apiKey := m.textInputs[0].Value()
	model := m.textInputs[1].Value()
	endpoint := m.textInputs[2].Value()
	tempStr := m.textInputs[3].Value()
	maxTokensStr := m.textInputs[4].Value()
	setDefaultStr := strings.ToLower(m.textInputs[5].Value())

	// Validation
	if apiKey == "" {
		return fmt.Errorf("API Key is required")
	}
	if model == "" {
		return fmt.Errorf("Model is required")
	}
	if endpoint == "" {
		return fmt.Errorf("Endpoint is required")
	}
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return fmt.Errorf("Invalid temperature: %v", err)
	}
	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil {
		return fmt.Errorf("Invalid max tokens: %v", err)
	}

	extraParams, err := json.Marshal(map[string]interface{}{
		"temperature": temp,
		"max_tokens":  maxTokens,
	})
	if err != nil {
		return err
	}

	settings := &db.ProviderSettings{
		ProviderName: m.selectedProvider,
		APIKey:       apiKey,
		Model:        model,
		Endpoint:     endpoint,
		ExtraParams:  string(extraParams),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	ctx := context.Background()
	if err := m.repoManager.SaveProviderSettings(ctx, settings); err != nil {
		return err
	}

	if setDefaultStr == "y" {
		if err := m.repoManager.SetDefaultProvider(ctx, m.selectedProvider); err != nil {
			return err
		}
	}

	return nil
}

// item implements list.Item
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }