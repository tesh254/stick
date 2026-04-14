package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tesh254/stick/internal/agent/tools"
	"github.com/tesh254/stick/internal/provider"
)

func (m *model) callAgent(userInput string) tea.Cmd {
	return func() tea.Msg {
		content, toolCalls, err := m.agentSession.Chat(context.Background(), userInput)
		if err != nil {
			return errMsg(err)
		}
		return agentResponseMsg{Content: content, ToolCalls: toolCalls}
	}
}

func (m *model) executePendingTool() tea.Cmd {
	if m.pendingToolCall == nil {
		return nil
	}
	return m.executeTool(*m.pendingToolCall)
}

func (m *model) executeTool(call provider.ToolCall) tea.Cmd {
	return func() tea.Msg {
		var output string
		var err error

		switch call.Function.Name {
		case "execute_command":
			var args struct {
				Command string `json:"command"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				out, cmdErr := tools.ExecuteCommand(args.Command)
				output = string(out)
				if cmdErr != nil {
					output += fmt.Sprintf("\nError: %v", cmdErr)
				}
			}

		case "read_file":
			var args struct {
				Path string `json:"path"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				output, err = tools.ReadFile(args.Path)
				if err != nil {
					output = fmt.Sprintf("Error reading file: %v", err)
				}
			}

		case "write_file":
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				output, err = tools.WriteFile(args.Path, args.Content)
				if err != nil {
					output = fmt.Sprintf("Error writing file: %v", err)
				}
			}

		case "list_files":
			var args struct {
				Path string `json:"path"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				if args.Path == "" {
					args.Path = "."
				}
				output, err = tools.ListFiles(args.Path)
				if err != nil {
					output = fmt.Sprintf("Error listing files: %v", err)
				}
			}

		case "search_files":
			var args struct {
				Path    string `json:"path"`
				Pattern string `json:"pattern"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				output, err = tools.SearchFiles(args.Path, args.Pattern)
				if err != nil {
					output = fmt.Sprintf("Error searching files: %v", err)
				}
			}

		case "fetch_url":
			var args struct {
				URL string `json:"url"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				output, err = tools.FetchURL(args.URL)
				if err != nil {
					output = fmt.Sprintf("Error fetching URL: %v", err)
				}
			}

		case "create_task_slice":
			var args struct {
				Tasks []string `json:"tasks"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				m.agentSession.AddTasks(args.Tasks)
				output = "Tasks created successfully."
			}

		case "update_task_status":
			var args struct {
				Index  int  `json:"index"`
				IsDone bool `json:"is_done"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				err := m.agentSession.UpdateTaskStatus(args.Index, args.IsDone)
				if err != nil {
					output = fmt.Sprintf("Error updating task: %v", err)
				} else {
					output = "Task updated successfully."
				}
			}

		case "task_complete":
			var args struct {
				Message string `json:"message"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				m.agentSession.MarkAllTasksAsComplete()
				output = "All tasks marked as complete. " + args.Message
			}

		case "switch_mode":
			var args struct {
				Mode string `json:"mode"`
			}
			if err = json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				output = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				m.agentSession.SetMode(args.Mode)
				output = fmt.Sprintf("Switched to mode: %s", args.Mode)
			}

		default:
			output = fmt.Sprintf("Unknown tool: %s", call.Function.Name)
		}

		// Send output back to agent
		content, toolCalls, err := m.agentSession.HandleToolOutput(context.Background(), call.ID, output)
		if err != nil {
			return errMsg(err)
		}
		return toolOutputMsg{
			ToolCallID: call.ID,
			Output:     output,
			ToolCalls:  toolCalls,
			Content:    content,
		}
	}
}

func (m model) handleAgentResponse(content string, toolCalls []provider.ToolCall) (model, tea.Cmd) {
	if content != "" {
		m.messages = append(m.messages, content)
	}

	m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
	m.viewport.GotoBottom()

	if len(toolCalls) > 0 {
		call := toolCalls[0]

		// Notify user about tool usage
		m.messages = append(m.messages, fmt.Sprintf("🤖 Calling tool: %s", call.Function.Name))
		m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
		m.viewport.GotoBottom()

		m.pendingToolCall = &call

		switch call.Function.Name {
		case "execute_command":
			m.isAskUserMode = true
			var args struct {
				Command string `json:"command"`
			}
			_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
			m.askUserModel = NewAskUserModel(fmt.Sprintf("Execute command: %s?", args.Command), "")
			return m, nil

		case "write_file":
			m.isAskUserMode = true
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			_ = json.Unmarshal([]byte(call.Function.Arguments), &args)

			// Generate a simple diff view
			diff := generateDiffView(args.Path, args.Content)

			m.askUserModel = NewAskUserModel(fmt.Sprintf("Write file: %s?", args.Path), diff)
			return m, nil

		default:
			// For safe tools (read, list, etc.), execute immediately
			return m, m.executeTool(call)
		}
	}

	m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
	m.viewport.GotoBottom()
	return m, nil
}

func generateDiffView(path, newContent string) string {
	oldContent, err := tools.ReadFile(path)
	if err != nil {
		// New file
		return fmt.Sprintf("--- /dev/null\n+++ %s\n@@ New File @@\n%s", path, newContent)
	}

	// Very naive line-based diff
	// Real diffing is complex, but we can show a summary
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	if len(oldLines) > 50 || len(newLines) > 50 {
		return fmt.Sprintf("--- %s (existing)\n+++ %s (new)\n[File too large to diff, overwriting %d bytes with %d bytes]", path, path, len(oldContent), len(newContent))
	}

	// Just show side-by-side or unified-ish?
	// Let's just return a standard-looking block
	return fmt.Sprintf("--- %s\n+++ %s\n\n<<<< OLD CONTENT\n%s\n==== NEW CONTENT\n%s\n>>>>", path, path, oldContent, newContent)
}
