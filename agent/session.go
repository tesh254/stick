package agent

import (
	"encoding/json"
	"fmt"

	"github.com/tesh254/stick/agent/message"
	"github.com/tesh254/stick/internal/config"
)

// AgentSession manages the conversation with the AI model.
type AgentSession struct {
	provider      string
	client        AIClient
	messages      []Message
	responseChan  chan<- any
	userInputChan <-chan string
	promptChan    chan string
}

// NewAgentSession creates a new agent session.
func NewAgentSession(provider string, responseChan chan<- any, userInputChan <-chan string) (*AgentSession, error) {
	client, err := NewAIClient(provider)
	if err != nil {
		return nil, fmt.Errorf("error creating AI client: %v", err)
	}

	session := &AgentSession{
		provider:      provider,
		client:        client,
		messages:      []Message{},
		responseChan:  responseChan,
		userInputChan: userInputChan,
		promptChan:    make(chan string),
	}

	session.messages = append(session.messages, Message{
		Role:    "system",
		Content: systemPrompt,
	})

	return session, nil
}

// Run starts the agent session's main loop.
func (s *AgentSession) Run() {
	for prompt := range s.promptChan {
		s.processPrompt(prompt)
	}
}

// ProcessPrompt sends a prompt to the AI and handles the response.
func (s *AgentSession) ProcessPrompt(prompt string) {
	s.promptChan <- prompt
}

func (s *AgentSession) processPrompt(prompt string) {
	providerConfig, _, err := config.GetProviderConfig(s.provider)
	if err != nil {
		s.responseChan <- fmt.Sprintf("Error getting provider config: %v", err)
		return
	}

	s.messages = append(s.messages, Message{
		Role:    "user",
		Content: prompt,
	})

	for {
		req := ChatCompletionRequest{
			Model:     providerConfig.Model,
			Messages:  s.messages,
			Tools:     getTools(),
			MaxTokens: 4096,
		}

		resp, err := s.client.Create(req)
		if err != nil {
			s.responseChan <- fmt.Sprintf("Error creating chat completion: %v", err)
			return
		}

		if len(resp.Choices) == 0 {
			s.responseChan <- "No response from AI"
			return
		}

		respMessage := resp.Choices[0].Message
		// Convert resp.Choices[0].Message.ToolCalls to []agent.ToolCall
		var toolCalls []ToolCall
		for _, tc := range respMessage.ToolCalls {
			toolCalls = append(toolCalls, ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		s.messages = append(s.messages, Message{
			Role:      respMessage.Role,
			Content:   respMessage.Content,
			ToolCalls: toolCalls,
		})

		if len(respMessage.ToolCalls) > 0 {
			for _, toolCall := range respMessage.ToolCalls {
				var toolArgs map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &toolArgs); err != nil {
					s.responseChan <- fmt.Sprintf("Failed to unmarshal tool arguments: %v", err)
					continue
				}

				if toolCall.Function.Name == "request_user_input" {
					prompt, ok := toolArgs["prompt"].(string)
					if !ok {
						s.responseChan <- "Invalid prompt for user input"
						continue
					}
					// Signal TUI to ask for input
					s.responseChan <- fmt.Sprintf("USER_INPUT_REQUEST:%s", prompt)
					// Wait for user input
					userInput, ok := <-s.userInputChan
					if !ok {
						// Channel closed, terminate
						return
					}
					// Add user input to messages
					s.messages = append(s.messages, Message{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Name:       toolCall.Function.Name,
						Content:    userInput,
					})
					continue
				}

				s.responseChan <- message.AgentToolCallMsg{
					ToolID: toolCall.ID,
					Name:   toolCall.Function.Name,
					Args:   toolCall.Function.Arguments,
				}

				result, err := ExecuteTool(toolCall.Function.Name, toolArgs)
				if err != nil {
					s.responseChan <- message.AgentToolResultMsg{
						ToolID:  toolCall.ID,
						Name:    toolCall.Function.Name,
						Result:  fmt.Sprintf("Error executing tool: %v", err),
						IsError: true,
					}
					// Potentially add error message to conversation history
					s.messages = append(s.messages, Message{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Name:       toolCall.Function.Name,
						Content:    fmt.Sprintf("Error: %v", err),
					})
					continue
				}

				if toolCall.Function.Name == "task_complete" {
					s.responseChan <- result
					s.responseChan <- "AGENT_DONE"
					return
				}

				s.responseChan <- message.AgentToolResultMsg{
					ToolID: toolCall.ID,
					Name:   toolCall.Function.Name,
					Result: result,
				}
				s.messages = append(s.messages, Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Name:       toolCall.Function.Name,
					Content:    result,
				})
			}
		} else {
			// If no tool calls, send the message content to the user and exit the loop.
			s.responseChan <- respMessage.Content
			s.responseChan <- "AGENT_DONE"
			return
		}
	}
}
