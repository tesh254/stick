package tui

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/ffs/core"
	"github.com/tesh254/stick/agent"
	"github.com/tesh254/stick/agent/message"
	"github.com/tesh254/stick/internal/constants"
	"github.com/tesh254/stick/render"
)

const (
	minInputHeight = 3
	maxInputHeight = 8
	dropdownHeight = 15
	maxTreeDepth   = 5
)

type agentResponseMsg string
type agentDoneMsg struct{}
type agentWaitingForInputMsg struct {
	prompt string
}

type agentStartMsg struct{}

type Model struct {
	viewport            viewport.Model
	input               textarea.Model
	messages            []string
	status              string
	showCommandDropdown bool
	showFileDropdown    bool
	commandList         list.Model
	fileList            list.Model
	currentDir          string
	dirTree             *core.DirectoryTree
	fullCommandItems    []list.Item
	inputHeight         int
	provider            string
	agentSession        *agent.AgentSession
	agentResponses      chan any
	userInputChan       chan string
}

func NewModel(provider string, currentDir string, dirTree *core.DirectoryTree) *Model {
	ti := textarea.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.SetHeight(minInputHeight)
	ti.SetValue("")

	vp := viewport.New(0, 0)

	agentResponses := make(chan any)
	userInputChan := make(chan string)

	agentSession, err := agent.NewAgentSession(provider, agentResponses, userInputChan)
	if err != nil {
		// Handle error appropriately, maybe return an error from NewModel
		panic(err)
	}

	m := &Model{
		input:          ti,
		viewport:       vp,
		messages:       []string{},
		currentDir:     currentDir,
		dirTree:        dirTree,
		inputHeight:    minInputHeight,
		provider:       provider,
		agentSession:   agentSession,
		agentResponses: agentResponses,
		userInputChan:  userInputChan,
	}

	m.initCommandList()
	m.updateFileList()
	return m
}

func (m *Model) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) setStatus(status string) {
	m.status = status
}

func (m *Model) waitForAgentResponse() tea.Cmd {
	return func() tea.Msg {
		response, ok := <-m.agentResponses
		if !ok {
			return agentDoneMsg{}
		}

		switch msg := response.(type) {
		case string:
			if strings.HasPrefix(msg, "USER_INPUT_REQUEST:") {
				prompt := strings.TrimPrefix(msg, "USER_INPUT_REQUEST:")
				return agentWaitingForInputMsg{prompt: prompt}
			}
			if msg == "AGENT_DONE" {
				return agentDoneMsg{}
			}
			return agentResponseMsg(msg)
		case message.AgentToolCallMsg, message.AgentToolResultMsg:
			return msg
		default:
			return agentResponseMsg(fmt.Sprintf("unknown response type: %T", msg))
		}
	}
}

func (m *Model) Init() tea.Cmd {
	m.addMessage("\n\n\n" + constants.STICK_ASCII)
	m.addMessage(render.GreenText.Render("provider: " + m.provider))
	m.addMessage(render.GrayText.Render("stick version: " + constants.VERSION()))
	return tea.Batch(textarea.Blink, m.input.Focus(), func() tea.Msg {
		return agentStartMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.input.SetWidth(msg.Width - 6)
		m.commandList.SetWidth(msg.Width - 2)
		m.fileList.SetWidth(msg.Width - 2)
		h := msg.Height - m.inputHeight
		if m.showCommandDropdown || m.showFileDropdown {
			h -= dropdownHeight
		}
		m.viewport.Height = h
		m.commandList.SetHeight(dropdownHeight)
		m.fileList.SetHeight(dropdownHeight)
		m.updateViewportContent()
		return m, nil
	case agentStartMsg:
		go m.agentSession.Run()
		return m, nil

	case agentResponseMsg:
		m.addMessage(render.CreamHighlight.Render(render.BoldText.Render(render.PurpleText.Render("stick:"))) + " " + string(msg))
		return m, m.waitForAgentResponse()
	case message.AgentToolCallMsg:
		m.setStatus(fmt.Sprintf("running tool: %s", msg.Name))
		m.addMessage(render.RenderToolCall(msg.Name, msg.Args))
		return m, m.waitForAgentResponse()
	case message.AgentToolResultMsg:
		m.setStatus("thinking...")
		if msg.Name == "print_code_block" {
			m.addMessage(render.RenderCodeBlock(msg.Result))
		} else if msg.Name == "render_markdown" {
			m.addMessage(render.RenderMarkdown(msg.Result))
		} else {
			m.addMessage(render.RenderToolResult(msg.Name, msg.Result, msg.IsError))
		}
		return m, m.waitForAgentResponse()
	case agentWaitingForInputMsg:
		m.setStatus(fmt.Sprintf("waiting for input: %s", msg.prompt))
		m.input.Placeholder = msg.prompt
		m.input.Focus()
		return m, nil
	case agentDoneMsg:
		m.setStatus("")
		m.addMessage("what's next")
		m.input.Focus()
		return m, nil

	case tea.KeyMsg:
		if m.status != "" {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if m.status != "waiting for input" {
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		var inputCmd, listCmd, vpCmd tea.Cmd
		inputUpdated := false
		listUpdated := false

		text := m.input.Value()
		lastSlashIndex := strings.LastIndex(text, "/")
		lastAtIndex := strings.LastIndex(text, "@")

		isFileTrigger := lastAtIndex != -1
		isCommandTrigger := !isFileTrigger && lastSlashIndex != -1

		if isCommandTrigger {
			sliceAfterSlash := text[lastSlashIndex+1:]
			hasSpaceAfterSlash := strings.Contains(sliceAfterSlash, " ")
			wasShowingCommand := m.showCommandDropdown
			m.showCommandDropdown = !hasSpaceAfterSlash
			if !wasShowingCommand && m.showCommandDropdown {
				m.input.SetCursor(len(text))
			}
		} else {
			m.showCommandDropdown = false
		}

		if isFileTrigger {
			sliceAfterAt := text[lastAtIndex+1:]
			hasSpaceAfterAt := strings.Contains(sliceAfterAt, " ")
			wasShowingFile := m.showFileDropdown
			m.showFileDropdown = !hasSpaceAfterAt
			if !wasShowingFile && m.showFileDropdown {
				m.input.SetCursor(len(text))
			}
		} else {
			m.showFileDropdown = false
		}

		if msg.String() == "enter" && !m.showCommandDropdown && !m.showFileDropdown {
			return m.execute()
		}

		if m.showCommandDropdown {
			switch msg.String() {
			case "up", "down", "k", "j":
				m.commandList, listCmd = m.commandList.Update(msg)
				listUpdated = true
			case "enter":
				item := m.commandList.SelectedItem()
				if item != nil {
					prefix := text[:lastSlashIndex]
					text = prefix + item.FilterValue()
					m.input.SetValue(text)
					m.showCommandDropdown = false
					return m.execute()
				}
				m.input.Focus()
				return m, tea.Batch(m.input.Focus(), nil)
			case "esc":
				m.input.SetValue(text[:lastSlashIndex])
				m.showCommandDropdown = false
				m.input.Focus()
				return m, tea.Batch(m.input.Focus(), nil)
			default:
				m.input, inputCmd = m.input.Update(msg)
				inputUpdated = true
			}
		} else if m.showFileDropdown {
			switch msg.String() {
			case "up", "down", "k", "j":
				m.fileList, listCmd = m.fileList.Update(msg)
				listUpdated = true
			case "enter":
				item := m.fileList.SelectedItem()
				if item != nil {
					fi, ok := item.(fileItem)
					if ok {
						sliceAfterAt := text[lastAtIndex+1:]
						parts := strings.Split(sliceAfterAt, "/")
						prefix := strings.Join(parts[:len(parts)-1], "/")
						if prefix != "" {
							prefix += "/"
						}
						newVal := prefix + fi.name
						if fi.isDir {
							newVal += "/"
						}
						text = text[:lastAtIndex] + "@" + newVal
						m.input.SetValue(text)
						m.showFileDropdown = false
					}
				}
				m.input.Focus()
				return m, tea.Batch(m.input.Focus(), nil)
			case "esc":
				m.input.SetValue(text[:lastAtIndex])
				m.showFileDropdown = false
				m.input.Focus()
				return m, tea.Batch(m.input.Focus(), nil)
			default:
				m.input, inputCmd = m.input.Update(msg)
				inputUpdated = true
			}
		} else {
			m.input, inputCmd = m.input.Update(msg)
			inputUpdated = true
		}

		if inputUpdated {
			text = m.input.Value()
			lines := m.input.LineCount()
			if text == "" {
				lines = 1
			}
			m.inputHeight = minInputHeight + (lines - 1)
			if m.inputHeight > maxInputHeight {
				m.inputHeight = maxInputHeight
			}
			m.input.SetHeight(m.inputHeight)

			isCommandTrigger = strings.Contains(text, "/")
			isFileTrigger = strings.Contains(text, "@")

			if m.showCommandDropdown {
				lastSlashIndex = strings.LastIndex(text, "/")
				filter := strings.ToLower(text[lastSlashIndex+1:])
				var filtered []list.Item
				for _, it := range m.fullCommandItems {
					if strings.Contains(strings.ToLower(it.FilterValue()[1:]), filter) {
						filtered = append(filtered, it)
					}
				}
				sort.Slice(filtered, func(i, j int) bool {
					return filtered[i].FilterValue() < filtered[j].FilterValue()
				})
				m.commandList.SetItems(filtered)
				if len(filtered) > 0 {
					m.commandList.Select(0)
				}
			} else if m.showFileDropdown {
				lastAtIndex = strings.LastIndex(text, "@")
				val := text[lastAtIndex+1:]
				var parts []string
				var lastPart string
				var prefixPath string
				var current *core.DirectoryTree

				val = strings.TrimSpace(val)
				if val == "" {
					current = m.dirTree
					prefixPath = ""
					lastPart = ""
				} else {
					parts = strings.Split(val, "/")
					lastPart = parts[len(parts)-1]
					prefixParts := parts[:len(parts)-1]
					current = m.dirTree
					prefixPath = ""
					valid := true
					for _, p := range prefixParts {
						found := false
						for _, child := range current.Children {
							if child.Name == p && !child.IsFile {
								current = &child
								prefixPath = filepath.Join(prefixPath, p)
								found = true
								break
							}
						}
						if !found {
							valid = false
							break
						}
					}
					if !valid {
						current = nil
					}
					if len(prefixParts) > 0 {
						prefixPath += "/"
					}
				}

				if current == nil {
					m.fileList.SetItems([]list.Item{})
					m.fileList.Title = "Invalid path"
				} else {
					var filtered []list.Item
					lowerLast := strings.ToLower(lastPart)
					for _, child := range current.Children {
						if strings.Contains(strings.ToLower(child.Name), lowerLast) {
							filtered = append(filtered, fileItem{name: child.Name, isDir: !child.IsFile})
						}
					}
					sort.Slice(filtered, func(i, j int) bool {
						fi := filtered[i].(fileItem)
						fj := filtered[j].(fileItem)
						if fi.isDir == fj.isDir {
							return fi.name < fj.name
						}
						return fi.isDir
					})
					m.fileList.SetItems(filtered)
					if len(filtered) > 0 {
						m.fileList.Select(0)
					}
					m.fileList.Title = "Directory: " + filepath.Join(m.currentDir, prefixPath)
				}
			}
		}

		if listUpdated {
			if m.showCommandDropdown {
				item := m.commandList.SelectedItem()
				if item != nil {
					prefix := text[:strings.LastIndex(text, "/")]
					m.input.SetValue(prefix + item.FilterValue())
				}
			} else if m.showFileDropdown {
				item := m.fileList.SelectedItem()
				if item != nil {
					fi, ok := item.(fileItem)
					if ok {
						val := text[strings.LastIndex(text, "@")+1:]
						parts := strings.Split(val, "/")
						prefix := strings.Join(parts[:len(parts)-1], "/")
						if prefix != "" {
							prefix += "/"
						}
						newVal := prefix + fi.name
						if fi.isDir {
							newVal += "/"
						}
						m.input.SetValue(text[:strings.LastIndex(text, "@")] + "@" + newVal)
					}
				}
			}
		}

		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, tea.Batch(inputCmd, listCmd, vpCmd)
	}

	return m, cmd
}

func (m *Model) execute() (tea.Model, tea.Cmd) {
	text := m.input.Value()
	if text == "" {
		return m, nil
	}

	if m.status != "" {
		m.userInputChan <- text
		m.setStatus("thinking...")
		m.input.SetValue("")
		m.input.Placeholder = ""
		return m, m.waitForAgentResponse()
	}

	m.addMessage(render.OrangeHighlight.Render(render.BoldText.Render(render.BlueText.Render("you:"))) + " " + text)

	var cmd tea.Cmd
	if strings.HasPrefix(text, "/") {
		commandParts := strings.Fields(text)
		cmdName := commandParts[0]
		switch cmdName {
		case "/help":
			m.addMessage("System: Help commands: /help, /exit")
		case "/exit":
			cmd = tea.Quit
		default:
			m.addMessage("System: Unknown command " + cmdName)
		}
		m.input.SetValue("")
		m.inputHeight = minInputHeight
		m.input.Focus()
		return m, cmd
	}

	m.setStatus("thinking...")
	m.sendPromptToAgent(text)

	m.input.SetValue("")
	m.inputHeight = minInputHeight
	m.input.Focus()

	return m, m.waitForAgentResponse()
}

func (m *Model) View() string {
	content := m.viewport.View()
	bottom := ""
	if m.status != "" {
		bottom = lipgloss.NewStyle().Bold(true).Render(m.status)
	} else {
		prefix := "▲ "
		inputView := m.input.View()

		// Join the prefix and the input view horizontally
		prompt := lipgloss.JoinHorizontal(lipgloss.Top, prefix, inputView)

		// Style for the container with a border
		containerStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

		bottom = containerStyle.Render(prompt)
	}

	dropdown := ""
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Margin(1, 0, 0, 0)

	if m.showCommandDropdown {
		dropdown = style.Render(m.commandList.View())
	} else if m.showFileDropdown {
		dropdown = style.Render(m.fileList.View())
	}

	if dropdown != "" {
		bottom = lipgloss.JoinVertical(lipgloss.Left, dropdown, bottom)
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, bottom)
}

// sendPromptToAgent kicks off the agent execution in a goroutine and streams results back
func (m *Model) sendPromptToAgent(prompt string) {
	m.agentSession.ProcessPrompt(prompt)
}

func (m *Model) initCommandList() {
	items := []list.Item{
		commandItem("/help"),
		commandItem("/exit"),
	}
	m.fullCommandItems = items
	del := commandDelegate{}
	m.commandList = list.New(append([]list.Item(nil), items...), del, 0, 0)
	m.commandList.Title = "Commands"
	m.commandList.SetShowHelp(false)
	m.commandList.SetShowStatusBar(false)
	m.commandList.SetShowPagination(false)
}

func (m *Model) updateFileList() {
	items := []list.Item{}
	current := m.dirTree
	if current == nil {
		return
	}
	for _, child := range current.Children {
		items = append(items, fileItem{child.Name, !child.IsFile})
	}
	del := fileDelegate{}
	m.fileList = list.New(items, del, 0, 0)
	m.fileList.Title = "Directory: " + current.Path
	m.fileList.SetShowHelp(false)
	m.fileList.SetShowStatusBar(false)
	m.fileList.SetShowPagination(false)
}

type commandItem string

func (i commandItem) Title() string       { return string(i) }
func (i commandItem) Description() string { return "" }
func (i commandItem) FilterValue() string { return string(i) }

type fileItem struct {
	name  string
	isDir bool
}

func (i fileItem) Title() string {
	if i.isDir {
		return i.name + "/"
	}
	return i.name
}
func (i fileItem) Description() string { return "" }
func (i fileItem) FilterValue() string { return i.name }

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("63")).Bold(true)
)

type commandDelegate struct{}

func (d commandDelegate) Height() int                             { return 1 }
func (d commandDelegate) Spacing() int                            { return 0 }
func (d commandDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d commandDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(commandItem)
	if !ok {
		return
	}

	str := string(i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type fileDelegate struct{}

func (d fileDelegate) Height() int                             { return 1 }
func (d fileDelegate) Spacing() int                            { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fileDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fileItem)
	if !ok {
		return
	}

	str := i.Title()

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
