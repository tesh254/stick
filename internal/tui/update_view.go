package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/functions"
	"github.com/tesh254/stick/internal/utils"
)

// Update handles messages and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// Check if we need to handle slash auto-completion after the update
	newValue := m.textarea.Value()

	// Check if we should start slash mode
	if !m.isInSlashMode {
		// Start slash mode only if user types a slash at the beginning of text area content
		if len(newValue) > 0 && newValue[0] == '/' {
			m.startSlashMode()
		}
		// Don't start slash mode if slash appears anywhere else (e.g. after a space)
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
	}

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(v)
	case tea.KeyMsg:
		return m.handleKeyMsg(v)
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
				// Check if input looks like a function call by trying to parse it
				p := functions.Parser{}
				name, args, err := p.Parse(input)

				// If parsing succeeds and no error, it's a function call
				if err == nil && name != "" {
					// Attempt to call the function
					result, err := functionRegistry.Call(name, args)

					if err != nil {
						// Function call failed, display the input and error
						messages = append(messages, prefix+input)
						messages = append(messages, "Error: "+err.Error())
					} else if result != "" {
						// Function call succeeded, display input and result
						messages = append(messages, prefix+input)
						messages = append(messages, "Function call result: "+result)
					} else {
						// Regular message, not a function call
						messages = append(messages, prefix+input)
					}
				} else {
					// Not a function call, regular message
					messages = append(messages, prefix+input)
				}
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
		}
		return m, nil
	case errMsg:
		m.err = v
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// formatMessagesHelper joins messages with newlines for display in the viewport
func formatMessagesHelper(messages []string) string {
	return strings.Join(messages, "\n")
}

// View renders the UI components
func (m model) View() string {
	// Render the viewport (main content area)
	viewportContent := m.viewport.View()

	// Conditionally render the search modal above the textarea if active
	var modalContent string
	if m.showSearchModal {
		modalContent = m.renderSearchModal() + "\n"
	}

	// Render the textarea
	textareaContent := m.textarea.View()

	// Combine everything: viewport, optional modal, gap, textarea
	if m.showSearchModal {
		return fmt.Sprintf(
			"%s\n%s%s%s",
			viewportContent,
			modalContent, // Render modal between viewport and textarea
			gap,
			textareaContent,
		)
	} else {
		return fmt.Sprintf(
			"%s%s%s",
			viewportContent,
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
	for i, cmd := range m.filteredCommands {
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
