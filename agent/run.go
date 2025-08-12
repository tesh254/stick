package agent

import (
	"encoding/json"
	"fmt"

	"github.com/tesh254/stick/internal/config"
)

func getTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: &Function{
				Name:        "run_tool",
				Description: "run a command in the terminal",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"command": {
							Type:        "string",
							Description: "the command to run",
						},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "create_file",
				Description: "creates a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to create the file",
						},
						"content": {
							Type:        "string",
							Description: "the content of the file",
						},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "create_dir",
				Description: "creates a directory in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to create the directory",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "delete_dir",
				Description: "deletes a directory in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to delete the directory",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "delete_file",
				Description: "deletes a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to delete the file",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "read_file",
				Description: "reads a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to read the file",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "dir_tree",
				Description: "returns the directory tree of the current directory in json",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to read the directory tree",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "patch_file",
				Description: "patches a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to patch the file",
						},
						"edits": {
							Type:        "array",
							Description: `list of edit operations to perform on the file.`,
							Items: Item{
								Type: "object",
								Properties: map[string]Prop{
									"action": {
										Type:        "string",
										Enum:        []string{"replace", "insert"},
										Description: `The type of edit operation: 'replace' to overwrite a line, 'insert' to add a new line.`,
									},
									"line_number": {
										Type:        "integer",
										Description: `The 1-based line number where the edit should occur.`,
									},
									"new_content": {
										Type:        "string",
										Description: `The content to insert or replace. Supports multi-line content with escaped newlines.`,
									},
								},
								Required: []string{"action", "line", "content"},
							},
						},
					},
					Required: []string{"path", "edits"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "request_user_input",
				Description: "requests user input",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"prompt": {
							Type:        "string",
							Description: "the prompt to display to the user",
						},
					},
					Required: []string{"prompt"},
				},
			},
		},
	}
}

type UserInputRequest struct {
	Prompt string
}

func RunAgent(prompt string, provider string, responseChan chan string, userInputChan chan string) {
	defer close(responseChan)

	client, err := NewAIClient(provider)
	if err != nil {
		responseChan <- fmt.Sprintf("Error creating AI client: %v", err)
		return
	}

	providerConfig, _, err := config.GetProviderConfig(provider)
	if err != nil {
		responseChan <- fmt.Sprintf("Error getting provider config: %v", err)
		return
	}

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	for {
		req := ChatCompletionRequest{
			Model:     providerConfig.Model,
			Messages:  messages,
			Tools:     getTools(),
			MaxTokens: 4096,
		}

		resp, err := client.Create(req)
		if err != nil {
			responseChan <- fmt.Sprintf("Error creating chat completion: %v", err)
			return
		}

		if len(resp.Choices) == 0 {
			responseChan <- "No response from AI"
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

		messages = append(messages, Message{
			Role:      respMessage.Role,
			Content:   respMessage.Content,
			ToolCalls: toolCalls,
		})

		if len(respMessage.ToolCalls) > 0 {
			for _, toolCall := range respMessage.ToolCalls {
				var toolArgs map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &toolArgs); err != nil {
					responseChan <- fmt.Sprintf("Failed to unmarshal tool arguments: %v", err)
					continue
				}

				if toolCall.Function.Name == "request_user_input" {
					prompt, ok := toolArgs["prompt"].(string)
					if !ok {
						responseChan <- "Invalid prompt for user input"
						continue
					}
					// Signal TUI to ask for input
					responseChan <- fmt.Sprintf("USER_INPUT_REQUEST:%s", prompt)
					// Wait for user input
					userInput, ok := <-userInputChan
					if !ok {
						// Channel closed, terminate
						return
					}
					// Add user input to messages
					messages = append(messages, Message{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Name:       toolCall.Function.Name,
						Content:    userInput,
					})
					continue
				}

				result, err := ExecuteTool(toolCall.Function.Name, toolArgs)
				if err != nil {
					responseChan <- fmt.Sprintf("Error executing tool: %v", err)
					// Potentially add error message to conversation history
					messages = append(messages, Message{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Name:       toolCall.Function.Name,
						Content:    fmt.Sprintf("Error: %v", err),
					})
					continue
				}
				responseChan <- fmt.Sprintf("Tool Result: %s", result)
				messages = append(messages, Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Name:       toolCall.Function.Name,
					Content:    result,
				})
			}
		} else {
			// If no tool calls, send the message content to the user and exit the loop.
			responseChan <- respMessage.Content
			return
		}
	}
}
