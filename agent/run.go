package agent

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tesh254/stick/internal/config"
)

func RunAgent(prompt string, provider string) {
	client, err := NewAIClient(provider)
	if err != nil {
		log.Fatal(err)
	}

	providerConfig, _, err := config.GetProviderConfig(provider)
	if err != nil {
		log.Fatal(err)
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

	tools := []Tool{
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
	}

	req := ChatCompletionRequest{
		Model:     providerConfig.Model,
		Messages:  messages,
		Tools:     tools,
		MaxTokens: 4096,
	}

	resp, err := client.Create(req)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response content
	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", string(respBytes))

	// Handle any tool calls in the response
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.ToolCalls) > 0 {
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			// Unmarshal the arguments from the tool call
			var toolArgs map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &toolArgs); err != nil {
				log.Fatalf("Failed to unmarshal tool arguments: %v", err)
			}

			result, err := ExecuteTool(toolCall.Function.Name, toolArgs)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Tool Result:", result)
		}
	}
}
