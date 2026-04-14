package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/utils"
)

// Update handles messages and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	if m.isAskUserMode {
		newAskModel, cmd := m.askUserModel.Update(msg)
		m.askUserModel = newAskModel.(AskUserModel)

		if m.askUserModel.quitting || m.askUserModel.choice != "" {
			m.isAskUserMode = false
			if m.askUserModel.choice == "Yes" {
				return m, m.executePendingTool()
			} else if m.askUserModel.choice == "No" {
				m.messages = append(m.messages, "Action cancelled.")
				// Send cancelled output to agent
				pendingID := m.pendingToolCall.ID
				m.pendingToolCall = nil

				return m, func() tea.Msg {
					content, toolCalls, err := m.agentSession.HandleToolOutput(context.Background(), pendingID, "User cancelled execution.")
					if err != nil {
						return errMsg(err)
					}
					return toolOutputMsg{ToolCallID: pendingID, Output: "User cancelled execution.", ToolCalls: toolCalls, Content: content}
				}
			}
			m.pendingToolCall = nil
		}
		return m, cmd
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// Check if we need to handle slash auto-completion after the update
	newValue := m.textarea.Value()

	// Check if we should start slash mode or file mode
	if !m.isInSlashMode && !m.isInAtMode {
		// Start slash mode only if user types a slash at the beginning of text area content
		if len(newValue) > 0 && newValue[0] == '/' {
			m.startSlashMode()
		} else {
			// Check for @ to start file mode
			words := strings.Fields(newValue)
			if len(words) > 0 {
				lastWord := words[len(words)-1]
				if strings.HasPrefix(lastWord, "@") {
					m.startFileMode()
					m.fileSearchInput = strings.TrimPrefix(lastWord, "@")
					m.updateFilteredFiles()
				}
			} else if len(newValue) > 0 && strings.HasSuffix(newValue, "@") {
				// Handle case where @ is the first char or after space but Fields didn't catch it yet (unlikely but possible if just "@")
				// strings.Fields("@") gives ["@"], so it is caught above.
				// But if I just typed "@", newValue is "@". Fields is ["@"]. lastWord is "@".
				// So the above block handles it.
			}
		}
	} else if m.isInSlashMode {
		// Update search input and filter commands as user types in slash mode
		if len(newValue) > 0 && strings.HasPrefix(newValue, "/") {
			// Update search input for filtering
			m.searchInput = newValue
			m.updateFilteredCommands()

			// Check if user has added a space after a slash command, which may indicate they're done with the command
			// For example, if they type "/help" and add a space, we could close the modal
			if strings.Contains(newValue, " ") {
				// Extract the command part before the space
				parts := strings.Split(newValue, " ")
				commandPart := parts[0]

				// Check if this is a complete slash command that doesn't need more args in the modal
				if commandPart == "/functions" || strings.HasPrefix(commandPart, "/help_") {
					// This is a complete command, close the modal
					m.endSlashMode()
				}
			}
		} else {
			// If user removes the slash or it's no longer a slash command, exit slash mode
			m.endSlashMode()
		}
	} else if m.isInAtMode {
		// Update file search input
		words := strings.Fields(newValue)
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			if strings.HasPrefix(lastWord, "@") {
				m.fileSearchInput = strings.TrimPrefix(lastWord, "@")
				m.updateFilteredFiles()
			} else {
				// If user typed space, we might want to exit
				if strings.HasSuffix(newValue, " ") {
					m.endFileMode()
				}
			}
		} else {
			m.endFileMode()
		}
	}

	// Handle mouse events for viewport scrolling
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		switch mouseMsg.Type {
		case tea.MouseWheelUp:
			// Scroll viewport when mouse wheel is used, regardless of explicit focus
			// This allows mouse wheel scrolling without having to specifically focus the viewport
			m.viewport.LineUp(3)     // Scroll 3 lines up
			m.viewportFocused = true // Mark that viewport is being interacted with
			return m, nil
		case tea.MouseWheelDown:
			// Scroll viewport when mouse wheel is used
			m.viewport.LineDown(3)   // Scroll 3 lines down
			m.viewportFocused = true // Mark that viewport is being interacted with
			return m, nil
		}
	}

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(v)
	case tea.KeyMsg:
		return m.handleKeyMsg(v)
	case agentResponseMsg:
		m.isLoading = false
		return m.handleAgentResponse(v.Content, v.ToolCalls)
	case toolOutputMsg:
		m.isLoading = false
		return m.handleAgentResponse(v.Content, v.ToolCalls)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case submitMsg:
		// We need to trigger the submission, but handleEnterKey is in messages.go
		// Let's use a command that will trigger the enter key handling in the next cycle
		// For now, let's handle it inline with access to all necessary functions
		messages := m.messages
		commandHistory := m.commandHistory
		historyIndex := m.historyIndex
		viewport := m.viewport
		senderStyle := m.senderStyle
		wrapStyle := m.wrapStyle
		textarea := m.textarea
		functionRegistry := m.functionRegistry

		input := textarea.Value()
		if input != "" {
			username := utils.GetUser()
			prefix := senderStyle.Render("{" + username + "}: ")

			// Check if input is a slash command first (copied logic from processSlashCommand)
			if strings.HasPrefix(input, "/") {
				input = strings.TrimSpace(input)

				var response string
				// Handle /functions command
				if input == "/functions" {
					if functionRegistry == nil {
						response = "No functions are currently registered."
					} else {
						functions := functionRegistry.GetFunctions()
						if len(functions) == 0 {
							response = "No functions are currently registered."
						} else {
							var functionsList []string
							for name := range functions {
								functionsList = append(functionsList, name)
							}
							response = "Available functions: " + strings.Join(functionsList, ", ")
						}
					}
				} else if strings.HasPrefix(input, "/help_") {
					// Handle /help_{function} command
					functionName := strings.TrimPrefix(input, "/help_")
					functionName = strings.ToLower(functionName)

					// Check if the function exists in the registry
					functions := functionRegistry.GetFunctions()
					if _, exists := functions[functionName]; !exists {
						response = fmt.Sprintf("Function '%s' not found. Use /functions to see available functions.", functionName)
					} else {
						// Check if the function has min/max arg constraints
						var argInfo string
						if min, minExists := functionRegistry.GetMinArgs(functionName); minExists {
							if max, maxExists := functionRegistry.GetMaxArgs(functionName); maxExists {
								if max == -1 {
									argInfo = fmt.Sprintf(" (%d+ args)", min)
								} else {
									if min == max {
										argInfo = fmt.Sprintf(" (%d args)", min)
									} else {
										argInfo = fmt.Sprintf(" (%d-%d args)", min, max)
									}
								}
							} else {
								argInfo = fmt.Sprintf(" (%d args)", min)
							}
						} else {
							argInfo = " (0+ args)"
						}
						response = fmt.Sprintf("Function: %s%s", functionName, argInfo)
					}
				} else {
					// Unknown slash command
					response = fmt.Sprintf("Unknown command: %s. Available commands: /functions, /help_{function}", input)
				}

				messages = append(messages, prefix+input)
				if response != "" {
					messages = append(messages, response)
				}
			} else {
				// Regular text input
				messages = append(messages, prefix+input)
			}

			// Add the input to command history
			commandHistory = append(commandHistory, input)
			historyIndex = -1 // Reset history index to current (empty) state

			viewport.SetContent(wrapStyle.Render(formatMessagesHelper(messages)))
			textarea.Reset()
			viewport.GotoBottom()

			// Update the model with new values
			m.messages = messages
			m.commandHistory = commandHistory
			m.historyIndex = historyIndex
			m.viewport = viewport
			m.textarea = textarea

			// Exit slash mode after command is executed if it's still active
			if m.isInSlashMode {
				m.endSlashMode()
			}

			// Trigger agent if it was a regular input
			if !strings.HasPrefix(input, "/") {
				m.isLoading = true
				utils.LogDebug("Sending request to agent (submitMsg): %s", input)
				return m, tea.Batch(
					m.callAgent(input),
					m.spinner.Tick,
				)
			}
		}
		return m, nil
	case errMsg:
		m.err = v
		m.isLoading = false
		// Append error to messages so user sees it
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
		m.messages = append(m.messages, errorStyle.Render(fmt.Sprintf("Error: %v", v)))
		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
		m.viewport.GotoBottom()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// View renders the UI components
func (m model) View() string {
	// Render the viewport (main content area)
	viewportContent := m.viewport.View()

	// Conditionally render the search modal above the textarea if active
	var modalContent string
	if m.showSearchModal {
		modalContent = m.renderSearchModal() + "\n"
	} else if m.showFileModal {
		modalContent = m.renderFileModal() + "\n"
	}

	// Render the textarea
	textareaContent := m.textarea.View()

	// Render spinner if loading
	var spinnerContent string
	if m.isLoading {
		spinnerContent = fmt.Sprintf("\n%s Thinking...", m.spinner.View())
	}

	// Combine everything: viewport, optional modal, gap, textarea
	if m.showSearchModal || m.showFileModal {
		return fmt.Sprintf(
			"%s%s\n%s%s%s",
			viewportContent,
			spinnerContent,
			modalContent, // Render modal between viewport and textarea
			gap,
			textareaContent,
		)
	} else {
		return fmt.Sprintf(
			"%s%s%s%s%s",
			viewportContent,
			spinnerContent,
			gap,
			textareaContent,
		)
	}
}

// renderSearchModal renders the search modal UI
func (m model) renderSearchModal() string {
	// Define modal styles
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		MaxHeight(15).
		Background(lipgloss.Color("235"))

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Bold(true).
		MarginBottom(1)

	searchBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("12"))

	// Create search input display
	searchBox := searchBoxStyle.Render("> " + m.searchInput)

	// Create the command list
	var commandItems []string

	// Limit visible items
	start := 0
	end := len(m.filteredCommands)
	maxVisible := 8 // Fits in MaxHeight 15 comfortably with title and search box

	if end > maxVisible {
		if m.selectedIndex < maxVisible/2 {
			start = 0
			end = maxVisible
		} else if m.selectedIndex >= end-maxVisible/2 {
			start = end - maxVisible
			end = end
		} else {
			start = m.selectedIndex - maxVisible/2
			end = start + maxVisible
		}
	}

	for i := start; i < end; i++ {
		cmd := m.filteredCommands[i]
		itemStyle := lipgloss.NewStyle()
		if i == m.selectedIndex {
			itemStyle = itemStyle.
				Background(lipgloss.Color("5")).
				Foreground(lipgloss.Color("0"))
		}
		commandItems = append(commandItems, itemStyle.Render("  "+cmd+"  "))
	}

	commandList := strings.Join(commandItems, "\n")

	// Combine all modal elements
	modalContent := titleStyle.Render("Search Slash Commands") + "\n" +
		searchBox + "\n\n" +
		commandList

	return modalStyle.Render(modalContent)
}

// renderFileModal renders the file search modal UI
func (m model) renderFileModal() string {
	// Define modal styles
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		MaxHeight(15).
		Background(lipgloss.Color("235"))

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Bold(true).
		MarginBottom(1)

	searchBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("12"))

	// Create search input display
	searchBox := searchBoxStyle.Render("> " + m.fileSearchInput)

	// Create the file list
	var fileItems []string

	// Limit visible items
	start := 0
	end := len(m.filteredFiles)
	maxVisible := 8

	if end > maxVisible {
		if m.fileSelectedIndex < maxVisible/2 {
			start = 0
			end = maxVisible
		} else if m.fileSelectedIndex >= end-maxVisible/2 {
			start = end - maxVisible
			end = end
		} else {
			start = m.fileSelectedIndex - maxVisible/2
			end = start + maxVisible
		}
	}

	for i := start; i < end; i++ {
		f := m.filteredFiles[i]
		itemStyle := lipgloss.NewStyle()
		if i == m.fileSelectedIndex {
			itemStyle = itemStyle.
				Background(lipgloss.Color("5")).
				Foreground(lipgloss.Color("0"))
		}
		fileItems = append(fileItems, itemStyle.Render("  "+f+"  "))
	}

	fileList := strings.Join(fileItems, "\n")

	// Combine all modal elements
	modalContent := titleStyle.Render("Search Files") + "\n" +
		searchBox + "\n\n" +
		fileList

	return modalStyle.Render(modalContent)
}
