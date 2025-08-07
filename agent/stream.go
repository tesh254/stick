package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// MessageStream sends a streaming chat completion request and returns a channel of response chunks
func (c *Client) MessageStream(req ChatCompletionRequest) (chan string, error) {
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
