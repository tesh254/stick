package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TogetherClient is a client for the Together AI API.
type TogetherClient struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

// NewTogetherClient creates a new TogetherClient.
func NewTogetherClient(apiKey string) (*TogetherClient, error) {
	return &TogetherClient{
		Endpoint:   "https://api.together.xyz/v1/chat/completions",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Create sends a chat completion request to the Together AI API.
func (c *TogetherClient) Create(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s: error: %w", resp.StatusCode, body, err)
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}

// MessageStream sends a streaming chat completion request to the Together AI API.
func (c *TogetherClient) MessageStream(req ChatCompletionRequest) (chan string, error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for authentication, content type, and streaming
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	ch := make(chan string)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
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

				if err := json.Unmarshal([]byte(data), &event); err != nil {
					if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
						ch <- event.Choices[0].Delta.Content
					}
				}
			}
		}
	}()

	return ch, nil
}
