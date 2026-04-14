package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/tesh254/stick/internal/agent/tools"
	"github.com/tesh254/stick/internal/prompts"
	"github.com/tesh254/stick/internal/provider"
)

// Agent Modes
const (
	ModeCoding    = "coding"
	ModePlanning  = "planning"
	ModeDebugging = "debugging"
	ModeArchitect = "architect"
)

func NewSession() *Session {
	// Resolve Provider
	prov := viper.GetString("provider")
	if prov == "" {
		prov = viper.GetString("ai.platform")
	}
	if prov == "" {
		prov = provider.ProviderOpenAI
	}

	// Resolve API Key
	apiKey := viper.GetString(fmt.Sprintf("providers.%s.apikey", prov))
	if apiKey == "" {
		apiKey = viper.GetString("api_key")
	}
	if apiKey == "" {
		apiKey = viper.GetString("ai.apikey")
	}
	if apiKey == "" {
		apiKey = os.Getenv("STICK_API_KEY")
	}
	apiKey = strings.TrimSpace(apiKey)

	// Resolve Model
	model := viper.GetString(fmt.Sprintf("providers.%s.model", prov))
	if model == "" {
		model = viper.GetString("model")
	}
	if model == "" {
		model = viper.GetString("ai.model")
	}

	client := provider.NewClient(prov, apiKey, "", model)

	sysPrompt := prompts.NewPrompts().SystemPrompt

	return &Session{
		Tasks:  []Task{},
		Client: client,
		Mode:   ModeCoding,
		History: []provider.Message{
			{Role: "system", Content: sysPrompt},
		},
	}
}

// SetMode changes the agent's operating mode and updates the system prompt context
func (s *Session) SetMode(mode string) {
	s.Mode = mode
	// Inject mode-specific instruction
	instruction := ""
	switch mode {
	case ModePlanning:
		instruction = "You are now in PLANNING mode. Focus on creating a comprehensive task list using 'create_task_slice'. Do not write code yet. Analyze the requirements and break them down."
	case ModeDebugging:
		instruction = "You are now in DEBUGGING mode. Focus on reading files, running tests ('execute_command'), and analyzing errors. Use 'search_files' to find relevant code."
	case ModeArchitect:
		instruction = "You are now in ARCHITECT mode. Focus on high-level design, system structure, and explaining trade-offs. Use 'list_files' to understand the current project structure."
	case ModeCoding:
		instruction = "You are now in CODING mode. Focus on implementing the tasks. Use 'write_file' to create/edit files and 'execute_command' to verify."
	}

	if instruction != "" {
		s.History = append(s.History, provider.Message{
			Role:    "system",
			Content: instruction,
		})
	}
}

// PruneHistory keeps the history within a reasonable limit to save tokens
func (s *Session) PruneHistory() {
	const maxHistoryMessages = 20 // Keep system prompt + last N messages
	if len(s.History) > maxHistoryMessages {
		// Keep the first message (System Prompt)
		// And the last (maxHistoryMessages - 1)
		newHistory := make([]provider.Message, 0, maxHistoryMessages)
		newHistory = append(newHistory, s.History[0])

		offset := len(s.History) - (maxHistoryMessages - 1)
		newHistory = append(newHistory, s.History[offset:]...)

		s.History = newHistory
	}
}

// Chat sends a user message and returns the response + tool calls
func (s *Session) Chat(ctx context.Context, userInput string) (string, []provider.ToolCall, error) {
	// Append user message
	if userInput != "" {
		s.History = append(s.History, provider.Message{Role: "user", Content: userInput})
	}

	// Prune before sending
	s.PruneHistory()

	req := provider.ChatCompletionRequest{
		Messages: s.History,
		Tools:    tools.GetTools(),
	}

	resp, err := s.Client.CreateChatCompletion(req)
	if err != nil {
		return "", nil, err
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	msg := choice.Message

	// Append assistant message
	s.History = append(s.History, msg)

	return msg.Content, msg.ToolCalls, nil
}

// HandleToolOutput processes the result of a tool execution and continues the conversation
func (s *Session) HandleToolOutput(ctx context.Context, toolCallID, output string) (string, []provider.ToolCall, error) {
	s.History = append(s.History, provider.Message{
		Role:       "tool",
		Content:    output,
		ToolCallID: toolCallID,
	})

	// Prune before sending
	s.PruneHistory()

	// Continue conversation after tool output
	req := provider.ChatCompletionRequest{
		Messages: s.History,
		Tools:    tools.GetTools(),
	}

	resp, err := s.Client.CreateChatCompletion(req)
	if err != nil {
		return "", nil, err
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	msg := choice.Message
	s.History = append(s.History, msg)

	return msg.Content, msg.ToolCalls, nil
}

// Task Management Methods

func (s *Session) AddTasks(descriptions []string) {
	for _, desc := range descriptions {
		s.Tasks = append(s.Tasks, Task{
			Description: desc,
			IsDone:      false,
		})
	}
}

func (s *Session) UpdateTaskStatus(index int, isDone bool) error {
	if index < 0 || index >= len(s.Tasks) {
		return fmt.Errorf("task index out of bounds")
	}
	s.Tasks[index].IsDone = isDone
	return nil
}

func (s *Session) MarkAllTasksAsComplete() {
	for i := range s.Tasks {
		s.Tasks[i].IsDone = true
	}
}
