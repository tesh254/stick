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

const DEFAULT_MODEL = "moonshotai/kimi-k2-0905"

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
		BaseURL:    "https://openrouter.ai/api/v1",
	}
}

func (c *Client) setHeaders(req *http.Request, stream bool) {
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
}

// CreateChatCompletionStream sends a streaming chat completion request to the specified endpoint.
func (c *Client) CreateChatCompletionStream(req ChatCompletionRequest) (chan string, chan error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("failed to parse user request: %w", err)
		close(errCh)
		return nil, errCh
	}

	base := c.BaseURL
	if strings.TrimSpace(base) == "" {
		base = "https://openrouter.ai/api/v1"
	}
	url := strings.TrimRight(base, "/") + "/chat/completions"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("failed to create request: %w", err)
		close(errCh)
		return nil, errCh
	}

	// Set headers for authentication, content type, and streaming
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
		defer close(ch)    // close channel when done
		defer close(errCh) // close error channel when done

		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				var event struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err = json.Unmarshal([]byte(data), &event); err == nil {
					if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
						ch <- event.Choices[0].Delta.Content
					}
				} else {
					errCh <- fmt.Errorf("failed to parse stream event: %w", err)
					break
				}
			}
		}
		if scanner.Err(); err != nil {
			errCh <- fmt.Errorf("stream read error: %w", err)
		}
	}()

	return ch, nil
}

func (c *Client) GetToolCalls(response ChatCompletionResponse) []ToolCall {
	if response.Choices[0].Message.ToolCalls != nil {
		return response.Choices[0].Message.ToolCalls
	}

	return []ToolCall{}
}
