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
