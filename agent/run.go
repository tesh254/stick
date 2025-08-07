package agent

import (
	"encoding/json"
	"fmt"
	"log"
)

func RunAgent(prompt string) {
	client := Init()
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
	}

	req := ChatCompletionRequest{
		Model:     "Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8",
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
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
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
