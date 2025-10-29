package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tesh254/stick/internal/functions"
	"github.com/tesh254/stick/internal/utils"
)

const gap = "\n\n"

type errMsg error

// submitMsg is a message to trigger message submission
type submitMsg struct{}

// handleWindowSizeMsg adjusts the layout when the terminal window is resized
func (m *model) handleWindowSizeMsg(v tea.WindowSizeMsg) {
	// Fullscreen: use the whole terminal. Viewport height is terminal minus input + gap.
	m.viewport.Width = v.Width
	m.textarea.SetWidth(v.Width)

	// Calculate the height needed for non-viewport components
	textAreaHeight := m.textarea.Height()
	gapHeight := lipgloss.Height(gap)

	// If modal is shown, we need to reserve space for it as well
	var modalReservedHeight int
	if m.showSearchModal {
		// Reserve approximate height for modal (this is a fixed size modal)
		// Based on renderSearchModal settings: MaxHeight(15)
		modalReservedHeight = 15
	} else {
		modalReservedHeight = 0
	}

	totalNonViewportHeight := textAreaHeight + gapHeight + modalReservedHeight
	m.viewport.Height = v.Height - totalNonViewportHeight

	// Update wrap style to current viewport width
	m.wrapStyle = lipgloss.NewStyle().Width(m.viewport.Width)

	if len(m.messages) > 0 {
		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
	}
	// Only go to bottom if viewport is not focused, to preserve scroll position when user is reading history
	if !m.viewportFocused {
		m.viewport.GotoBottom()
	}
}

// handleKeyMsg processes keyboard inputs
func (m *model) handleKeyMsg(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search modal keys if modal is open
	if m.showSearchModal {
		return m.handleSearchModalKeys(v)
	}

	// Handle viewport-specific keys when viewport is focused
	if m.viewportFocused {
		switch v.Type {
		case tea.KeyUp:
			m.viewport.ScrollUp(1)
			return *m, nil
		case tea.KeyDown:
			m.viewport.ScrollDown(1)
			return *m, nil
		case tea.KeyPgUp:
			m.viewport.PageUp()
			return *m, nil
		case tea.KeyPgDown:
			m.viewport.PageDown()
			return *m, nil
		case tea.KeyHome:
			m.viewport.GotoTop()
			return *m, nil
		case tea.KeyEnd:
			m.viewport.GotoBottom()
			return *m, nil
		case tea.KeyEsc:
			// Unfocus viewport and return to normal mode
			m.viewportFocused = false
			return *m, nil
		}
	}

	switch v.Type {
	case tea.KeyCtrlC:
		// Print last textarea value then quit
		fmt.Println(m.textarea.Value())
		return *m, tea.Quit
	case tea.KeyEsc:
		// If viewport is focused, unfocus it; otherwise handle normally
		if m.viewportFocused {
			m.viewportFocused = false
			return *m, nil
		}
		// Exit search modal if in slash mode, otherwise quit
		if m.isInSlashMode {
			m.endSlashMode()
			return *m, nil
		} else {
			// Print last textarea value then quit
			fmt.Println(m.textarea.Value())
			return *m, tea.Quit
		}
	case tea.KeyEnter:
		// If in slash mode, user may want to submit what they have typed
		if m.isInSlashMode {
			// Let the modal handle Enter key, which will populate the textarea and exit slash mode
			// The message will be submitted in the next update cycle
		}
		// Handle enter normally (submit message)
		m.handleEnterKey()
	case tea.KeyUp:
		// Only navigate command history if viewport is not focused
		if !m.viewportFocused {
			m.handleUpArrow()
		} else {
			// If viewport is focused, scroll up
			m.viewport.ScrollUp(1)
		}
	case tea.KeyDown:
		// Only navigate command history if viewport is focused
		if !m.viewportFocused {
			m.handleDownArrow()
		} else {
			// If viewport is focused, scroll down
			m.viewport.ScrollDown(1)
		}
	case tea.KeyPgUp:
		// If viewport is focused, page up; otherwise, treat as up arrow for history
		if m.viewportFocused {
			m.viewport.PageUp()
		} else {
			m.handleUpArrow()
		}
		return *m, nil
	case tea.KeyPgDown:
		// If viewport is focused, page down; otherwise, treat as down arrow for history
		if m.viewportFocused {
			m.viewport.PageDown()
		} else {
			m.handleDownArrow()
		}
		return *m, nil
	case tea.KeyCtrlR: // Trigger search modal (fallback if needed)
		m.toggleSearchModal()
	case tea.KeyCtrlF: // Focus the viewport for scrolling
		m.viewportFocused = true
		return *m, nil
	}

	return *m, nil
}

// handleUpArrow handles the up arrow key to navigate command history
func (m *model) handleUpArrow() {
	if len(m.commandHistory) == 0 {
		return // No history to navigate
	}

	// If we're at the beginning of the history, start from the most recent command
	if m.historyIndex == -1 {
		m.historyIndex = len(m.commandHistory) - 1
	} else if m.historyIndex > 0 {
		m.historyIndex--
	}

	// Set the textarea value to the current history item
	m.textarea.SetValue(m.commandHistory[m.historyIndex])
	m.textarea.SetCursor(len(m.commandHistory[m.historyIndex]))
}

// handleDownArrow handles the down arrow key to navigate forward in command history
func (m *model) handleDownArrow() {
	if m.historyIndex == -1 || len(m.commandHistory) == 0 {
		return // No history to navigate or we're at the current position
	}

	m.historyIndex++

	// If we've gone past the end of history, reset to empty input
	if m.historyIndex >= len(m.commandHistory) {
		m.historyIndex = -1
		m.textarea.SetValue("")
	} else {
		// Set the textarea value to the current history item
		m.textarea.SetValue(m.commandHistory[m.historyIndex])
		m.textarea.SetCursor(len(m.commandHistory[m.historyIndex]))
	}
}

// handleEnterKey processes the Enter key press in the textarea
func (m *model) handleEnterKey() {
	input := m.textarea.Value()
	if input != "" {
		username := utils.GetUser()
		prefix := m.senderStyle.Render("{" + username + "}: ")

		// Check if input is a slash command first
		if strings.HasPrefix(input, "/") {
			response := m.processSlashCommand(input)
			m.messages = append(m.messages, prefix+input)
			if response != "" {
				m.messages = append(m.messages, response)
			}
		} else {
			// Check if input looks like a function call by trying to parse it
			result, err := m.processFunctionCall(input)

			if err != nil {
				// Function call failed, display the input and error
				m.messages = append(m.messages, prefix+input)
				m.messages = append(m.messages, "Error: "+err.Error())
			} else if result != "" {
				// Function call succeeded, display input and result
				m.messages = append(m.messages, prefix+input)
				m.messages = append(m.messages, "Function call result: "+result)
			} else {
				// Regular message, not a function call
				m.messages = append(m.messages, prefix+input)
			}
		}

		// Add the input to command history
		m.commandHistory = append(m.commandHistory, input)
		m.historyIndex = -1 // Reset history index to current (empty) state

		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
		m.textarea.Reset()
		m.viewport.GotoBottom()

		// Exit slash mode after command is executed
		if m.isInSlashMode {
			m.endSlashMode()
		}
	}
}

// processFunctionCall processes function calls in the input string
func (m *model) processFunctionCall(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", nil
	}

	openIdx := strings.Index(s, "(")
	if openIdx == -1 {
		// With no parentheses, always treat input as plain text (single or multi-word),
		// even if it matches a registered function name. Detection only occurs with '('.
		return "", nil
	}

	// If there is whitespace immediately before '(', treat as plain text (no function)
	if openIdx > 0 {
		prev := s[openIdx-1]
		if prev == ' ' || prev == '\t' || prev == '\n' || prev == '\r' {
			return "", nil
		}
	}

	p := functions.Parser{}
	parsed := p.ParseDetailed(s)

	// Enhance error messages with exact positions where possible
	if parsed.Error != nil {
		// If we have an opening parenthesis but no matching closing parenthesis
		closeIdx := strings.LastIndex(s, ")")
		if openIdx != -1 && (closeIdx == -1 || closeIdx < openIdx) {
			return "", fmt.Errorf("syntax error: missing closing ')' for '(' at position %d", openIdx)
		}
		// Propagate original parser error for other cases
		return "", parsed.Error
	}

	// If a function call was recognized
	if parsed.HasFunction && parsed.FunctionName != "" {
		// Enforce case-sensitive function name detection in the TUI layer
		funcs := m.functionRegistry.GetFunctions()
		if _, exists := funcs[parsed.FunctionName]; !exists {
			// Maintain current error reporting for truly unknown functions in proper call syntax
			return "", fmt.Errorf("unknown function: %s", parsed.FunctionName)
		}

		// Valid function call; support empty and parameterized calls
		result, err := m.functionRegistry.Call(parsed.FunctionName, parsed.Arguments)
		if err != nil {
			return "", err
		}
		return result, nil
	}

	// Not a function call -> treat as regular text
	return "", nil
}

// processSlashCommand processes slash commands like /functions and /help_{function}
func (m *model) processSlashCommand(input string) string {
	input = strings.TrimSpace(input)

	// Handle /functions command
	if input == "/functions" {
		return m.listFunctions()
	}

	// Handle /help_{function} command
	if strings.HasPrefix(input, "/help_") {
		functionName := strings.TrimPrefix(input, "/help_")
		return m.getFunctionHelp(functionName)
	}

	// Unknown slash command
	return fmt.Sprintf("Unknown command: %s. Available commands: /functions, /help_{function}", input)
}

// listFunctions returns a list of all registered functions
func (m *model) listFunctions() string {
	if m.functionRegistry == nil {
		return "No functions are currently registered."
	}

	functions := m.functionRegistry.GetFunctions()
	if len(functions) == 0 {
		return "No functions are currently registered."
	}

	var functionsList []string
	for name := range functions {
		functionsList = append(functionsList, name)
	}

	return "Available functions: " + strings.Join(functionsList, ", ")
}

// getFunctionHelp returns help information for a specific function
func (m *model) getFunctionHelp(functionName string) string {
	functionName = strings.ToLower(functionName)

	// Check if the function exists in the registry
	functions := m.functionRegistry.GetFunctions()
	if _, exists := functions[functionName]; !exists {
		return fmt.Sprintf("Function '%s' not found. Use /functions to see available functions.", functionName)
	}

	// Check if the function has min/max arg constraints
	var argInfo string
	if min, minExists := m.functionRegistry.GetMinArgs(functionName); minExists {
		if max, maxExists := m.functionRegistry.GetMaxArgs(functionName); maxExists {
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

	return fmt.Sprintf("Function: %s%s", functionName, argInfo)
}

// toggleSearchModal toggles the search modal visibility and initializes it
func (m *model) toggleSearchModal() {
	if m.isInSlashMode {
		// Close the modal
		m.endSlashMode()
	} else {
		// Open the modal and populate with slash commands
		m.startSlashMode()
	}
}

// handleSearchModalKeys handles keyboard input when the search modal is active
func (m *model) handleSearchModalKeys(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch v.Type {
	case tea.KeyEscape:
		// Close the modal
		m.endSlashMode()
		return *m, nil
	case tea.KeyCtrlC:
		// Close the modal if it's open, but still allow exit
		m.endSlashMode()
		// Let the original Ctrl+C behavior continue (exit the app)
		fmt.Println(m.textarea.Value())
		return *m, tea.Quit
	case tea.KeyEnter:
		// If there are filtered commands and a selection, insert the command and submit it
		if len(m.filteredCommands) > 0 && m.selectedIndex < len(m.filteredCommands) {
			selectedCommand := m.filteredCommands[m.selectedIndex]
			// Insert the selected command into the textarea
			m.textarea.SetValue(selectedCommand)
			m.textarea.SetCursor(len(selectedCommand))
			// Close the modal
			m.endSlashMode()
			// Return a submit command to process the message
			return *m, func() tea.Msg {
				return submitMsg{}
			}
		} else {
			// No command selected, just close the modal
			m.endSlashMode()
			// Continue processing this Enter key in the main handler to submit the message
			return *m, nil
		}
	case tea.KeyUp:
		if len(m.filteredCommands) > 0 {
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.filteredCommands) - 1
			}
		}
	case tea.KeyDown:
		if len(m.filteredCommands) > 0 {
			m.selectedIndex++
			if m.selectedIndex >= len(m.filteredCommands) {
				m.selectedIndex = 0
			}
		}
	case tea.KeyTab: // Use tab to select command without submitting
		if len(m.filteredCommands) > 0 && m.selectedIndex < len(m.filteredCommands) {
			selectedCommand := m.filteredCommands[m.selectedIndex]
			// Insert the selected command into the textarea
			m.textarea.SetValue(selectedCommand)
			m.textarea.SetCursor(len(selectedCommand))
		}
		return *m, nil
	}

	return *m, nil
}

// startSlashMode starts the slash command mode with modal
func (m *model) startSlashMode() {
	m.isInSlashMode = true
	m.showSearchModal = true
	m.searchInput = m.textarea.Value() // Use current textarea value as search input
	m.selectedIndex = 0
	m.populateSlashCommands()
	m.updateFilteredCommands()

	// Adjust viewport height to account for modal appearance
	m.adjustViewportForModal()
}

// endSlashMode ends the slash command mode and closes the modal
func (m *model) endSlashMode() {
	m.isInSlashMode = false
	m.showSearchModal = false
	m.searchInput = ""
	m.filteredCommands = []string{}
	m.selectedIndex = 0

	// Adjust viewport height to account for modal disappearance
	m.adjustViewportForModal()
}

// adjustViewportForModal adjusts the viewport height based on modal visibility
func (m *model) adjustViewportForModal() {
	textAreaHeight := m.textarea.Height()
	gapHeight := lipgloss.Height(gap)

	// Calculate the height needed for non-viewport components based on modal state
	var modalReservedHeight int
	if m.showSearchModal {
		// Reserve approximate height for modal
		modalReservedHeight = 15
	} else {
		modalReservedHeight = 0
	}

	totalNonViewportHeight := textAreaHeight + gapHeight + modalReservedHeight
	// Calculate terminal height from current state
	terminalHeight := m.viewport.Height + textAreaHeight + gapHeight
	// If modal was already being accounted for, we need to adjust accordingly
	if m.showSearchModal {
		terminalHeight += modalReservedHeight // Add back the modal height we were already accounting for
	}

	m.viewport.Height = terminalHeight - totalNonViewportHeight
}

// populateSlashCommands populates the list of available slash commands
func (m *model) populateSlashCommands() {
	// Define the available slash commands
	slashCommands := []string{
		"/functions",
		"/help_",
	}

	// Add help commands for each registered function
	if m.functionRegistry != nil {
		functions := m.functionRegistry.GetFunctions()
		for name := range functions {
			helpCmd := fmt.Sprintf("/help_%s", name)
			slashCommands = append(slashCommands, helpCmd)
		}
	}

	// Store the full list of commands
	m.allSlashCommands = slashCommands
	// Set filtered commands to the full list initially
	m.filteredCommands = slashCommands
}

// updateFilteredCommands updates the list of filtered commands based on search input
func (m *model) updateFilteredCommands() {
	if m.searchInput == "" {
		// Reset to all commands
		m.filteredCommands = m.allSlashCommands
		return
	}

	var filtered []string
	search := strings.ToLower(m.searchInput)

	// First, prioritize commands that start with the search term
	var exactStartMatches []string
	var containsMatches []string

	for _, cmd := range m.allSlashCommands {
		cmdLower := strings.ToLower(cmd)
		if strings.HasPrefix(cmdLower, search) {
			exactStartMatches = append(exactStartMatches, cmd)
		} else if strings.Contains(cmdLower, search) {
			containsMatches = append(containsMatches, cmd)
		}
	}

	// Combine results with exact start matches first
	filtered = append(filtered, exactStartMatches...)
	filtered = append(filtered, containsMatches...)

	m.filteredCommands = filtered
}

// formatMessages joins messages with newlines for display in the viewport
func formatMessages(messages []string) string {
	return strings.Join(messages, "\n")
}

// formatMessagesHelper joins messages with newlines for display in the viewport
func formatMessagesHelper(messages []string) string {
	return strings.Join(messages, "\n")
}
