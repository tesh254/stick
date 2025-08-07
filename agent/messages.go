package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (c *Client) Create(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
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
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, body)
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}
