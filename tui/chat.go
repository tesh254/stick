package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/tesh254/ffs/core"
	"github.com/tesh254/stick/internal/constants"
)

const (
	minInputHeight = 3
	maxInputHeight = 8
	dropdownHeight = 50
	maxTreeDepth   = 5
)

type Model struct {
	viewport            viewport.Model
	input               textarea.Model
	messages            []string
	loading             bool
	showCommandDropdown bool
	showFileDropdown    bool
	commandList         list.Model
	fileList            list.Model
	currentDir          string
	dirTree             *core.DirectoryTree
	fullCommandItems    []list.Item
	inputHeight         int
}

var purpleText = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
var boldText = lipgloss.NewStyle().Bold(true)

func (m *Model) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) toggleLoading() {
	m.loading = !m.loading
}

type loadingDoneMsg struct{}

func (m *Model) Init() tea.Cmd {
	// Add ASCII art to messages to ensure it persists
	m.addMessage(constants.STICK_ASCII)
	return tea.Batch(textarea.Blink, m.input.Focus())
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

	case tea.KeyMsg:
		if m.loading {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
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
		isCommandTrigger := lastSlashIndex >= 0 && (lastAtIndex == -1 || lastSlashIndex > lastAtIndex)
		isFileTrigger := lastAtIndex >= 0 && (lastSlashIndex == -1 || lastAtIndex > lastSlashIndex)

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
			if text == "" {
				return m, nil
			}
			m.addMessage(boldText.Render(purpleText.Render("you: ")) + text)
			if strings.HasPrefix(text, "/") {
				commandParts := strings.Fields(text)
				cmdName := commandParts[0]
				switch cmdName {
				case "/help":
					m.addMessage("System: Help commands: /help, /run, /exit")
				case "/run":
					m.loading = true
					m.addMessage("System: Running...")
					return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return loadingDoneMsg{} })
				case "/exit":
					return m, tea.Quit
				default:
					m.addMessage("System: Unknown command " + cmdName)
				}
			}
			m.input.SetValue("")
			m.inputHeight = minInputHeight
			m.input.Focus()
			return m, tea.Batch(m.input.Focus(), nil)
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

	case loadingDoneMsg:
		m.loading = false
		m.addMessage("System: Run complete.")
		m.input.Focus()
		return m, tea.Batch(m.input.Focus(), nil)
	}

	return m, cmd
}

func (m *Model) View() string {
	content := m.viewport.View()
	bottom := ""
	if m.loading {
		bottom = lipgloss.NewStyle().Bold(true).Render("Loading...")
	} else {
		inputContent := m.input.View()
		lines := strings.Split(inputContent, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimPrefix(line, ">")
			lines[i] = strings.TrimPrefix(lines[i], "> ")
			lines[i] = strings.TrimSpace(line)
		}
		// Prepend ▲ to the first non-empty line, ensuring input stays beside it
		if len(lines) > 0 {
			lines[0] = "▲ " + lines[0]
		}
		inputContent = strings.Join(lines, "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			Width(m.input.Width() + 2)
		bottom = inputStyle.Render(inputContent)
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

func (m *Model) initCommandList() {
	items := []list.Item{
		commandItem("/help"),
		commandItem("/run"),
		commandItem("/exit"),
	}
	m.fullCommandItems = items
	del := list.NewDefaultDelegate()
	m.commandList = list.New(append([]list.Item(nil), items...), del, 0, 0)
	m.commandList.Title = "Commands"
	m.commandList.SetShowHelp(false)
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
	del := list.NewDefaultDelegate()
	m.fileList = list.New(items, del, 0, 0)
	m.fileList.Title = "Directory: " + current.Path
	m.fileList.SetShowHelp(false)
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

func Chat() {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("This command requires an interactive terminal.")
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ti := textarea.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.SetHeight(minInputHeight)
	ti.SetValue("")

	vp := viewport.New(0, 0)
	vp.SetContent(constants.STICK_ASCII)

	dirTree, err := core.WorkingDirectoryTree(nil, []string{".git"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m := &Model{
		input:       ti,
		viewport:    vp,
		messages:    []string{constants.STICK_ASCII},
		currentDir:  cwd,
		dirTree:     &dirTree,
		inputHeight: minInputHeight,
	}

	m.initCommandList()
	m.updateFileList()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
