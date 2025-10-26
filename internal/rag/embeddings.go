package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (r *RAG) GenerateEmbeddings(content string) ([]float32, error) {
	payload := map[string]string{"text": content}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// make http POST request to cloudflare worker
	req, err := http.NewRequest("POST", r.Embeddings.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create requests")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Embeddings.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// fmt.Println(result) // Remove this line, it was for debugging

	if len(result.Data) == 0 || len(result.Data[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Data[0], nil
}

// Marshal serializes embeddings to JSON string.
func (r *RAG) Marshal(embeddings []float32) (string, error) {
	b, err := json.MarshalIndent(embeddings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal embeddings: %v", err)
	}
	return string(b), nil
}

// Unmarshal deserializes JSON string to embeddings.
func (r *RAG) Unmarshal(data string) ([]float32, error) {
	var embeddings []float32
	if err := json.Unmarshal([]byte(data), &embeddings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embeddings: %v", err)
	}
	return embeddings, nil
}
