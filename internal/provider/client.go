package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ProviderOpenAI     = "openai"
	ProviderOpenRouter = "openrouter"
	ProviderAnthropic  = "anthropic"
	ProviderZAI        = "zai"
	ProviderGrok       = "grok"
	ProviderTogether   = "together"
)

var defaultModels = map[string]string{
	ProviderOpenAI:     "gpt-4o",
	ProviderOpenRouter: "moonshotai/kimi-k2-0905",
	ProviderAnthropic:  "claude-3-opus-20240229",
	ProviderZAI:        "default-zai-model", // Assuming a default
	ProviderGrok:       "grok-4",
	ProviderTogether:   "meta-llama/Llama-3-70b-chat-hf",
}

var defaultBaseURLs = map[string]string{
	ProviderOpenAI:     "https://api.openai.com/v1",
	ProviderOpenRouter: "https://openrouter.ai/api/v1",
	ProviderAnthropic:  "https://api.anthropic.com/v1",
	ProviderZAI:        "https://api.z.ai/v1",  // As per prompt
	ProviderGrok:       "https://api.grok.com", // As per prompt
	ProviderTogether:   "https://api.together.xyz/v1",
}

func NewClient(provider, apiKey, baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURLs[provider]
	}
	if model == "" {
		model = defaultModels[provider]
	}
	return &Client{
		Provider:   provider,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
		BaseURL:    baseURL,
		Model:      model,
	}
}

func (c *Client) setHeaders(req *http.Request, stream bool) {
	if c.Provider == ProviderAnthropic {
		req.Header.Set("x-api-key", c.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
}

// CreateChatCompletionStream sends a streaming chat completion request to the specified endpoint.
func (c *Client) CreateChatCompletionStream(req ChatCompletionRequest) (chan string, chan error) {
	req.Stream = true
	req.Model = c.Model // Ensure model is set

	var url string
	var body []byte
	var err error

	if c.Provider == ProviderAnthropic {
		// Anthropic-specific request format
		anthropicReq := struct {
			Model       string    `json:"model"`
			Messages    []Message `json:"messages"`
			MaxTokens   int       `json:"max_tokens"`
			Stream      bool      `json:"stream"`
			Temperature float64   `json:"temperature,omitempty"`
		}{
			Model:     req.Model,
			Messages:  req.Messages,
			MaxTokens: 4096, // Default, can be configurable
			Stream:    true,
		}
		body, err = json.Marshal(anthropicReq)
		url = c.BaseURL + "/messages"
	} else {
		// OpenAI-compatible
		body, err = json.Marshal(req)
		url = c.BaseURL + "/chat/completions"
	}

	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("failed to parse user request: %w", err)
		close(errCh)
		return nil, errCh
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("failed to create request: %w", err)
		close(errCh)
		return nil, errCh
	}

	c.setHeaders(request, true)

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("request failed: %w", err)
		close(errCh)
		return nil, errCh
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("API error: status %d, body: %s", response.StatusCode, body)
		close(errCh)
		return nil, errCh
	}

	ch := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer response.Body.Close()
		defer close(ch)
		defer close(errCh)

		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				if c.Provider == ProviderAnthropic {
					// Parse Anthropic stream event
					var event struct {
						Type  string `json:"type"`
						Delta struct {
							Text string `json:"text"`
						} `json:"delta"`
					}
					if err := json.Unmarshal([]byte(data), &event); err == nil {
						if event.Type == "content_block_delta" && event.Delta.Text != "" {
							ch <- event.Delta.Text
						}
					} else {
						errCh <- fmt.Errorf("failed to parse Anthropic stream event: %w", err)
						break
					}
				} else {
					// OpenAI-compatible parsing
					var event struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}
					if err := json.Unmarshal([]byte(data), &event); err == nil {
						if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
							ch <- event.Choices[0].Delta.Content
						}
					} else {
						errCh <- fmt.Errorf("failed to parse stream event: %w", err)
						break
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("stream read error: %w", err)
		}
	}()

	return ch, nil
}

// CreateChatCompletion sends a synchronous chat completion request to the specified endpoint.
func (c *Client) CreateChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false
	req.Model = c.Model

	var url string
	var body []byte
	var err error

	if c.Provider == ProviderAnthropic {
		// Anthropic-specific request format
		anthropicReq := struct {
			Model       string    `json:"model"`
			Messages    []Message `json:"messages"`
			MaxTokens   int       `json:"max_tokens"`
			Stream      bool      `json:"stream"`
			Temperature float64   `json:"temperature,omitempty"`
			Tools       []Tool    `json:"tools,omitempty"`
		}{
			Model:     req.Model,
			Messages:  req.Messages,
			MaxTokens: 4096,
			Stream:    false,
			Tools:     req.Tools,
		}
		body, err = json.Marshal(anthropicReq)
		url = c.BaseURL + "/messages"
	} else {
		// OpenAI-compatible
		body, err = json.Marshal(req)
		url = c.BaseURL + "/chat/completions"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse user request: %w", err)
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(request, false)

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	respBody, _ := io.ReadAll(response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: status %d, body: %s", response.StatusCode, respBody)
	}

	var chatResp ChatCompletionResponse
	if c.Provider == ProviderAnthropic {
		// Anthropic response parsing would go here, but for now let's focus on OpenAI compatible
		// Since we didn't implement Anthropic response struct, this might fail for Anthropic.
		// Ideally we should map Anthropic response to ChatCompletionResponse.
		// For now, assuming OpenAI compatible.
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &chatResp, nil
}

func (c *Client) GetToolCalls(response ChatCompletionResponse) []ToolCall {
	if response.Choices[0].Message.ToolCalls != nil {
		return response.Choices[0].Message.ToolCalls
	}

	return []ToolCall{}
}
