package rag

import "net/http"

type RAG struct {
	Embeddings Embeddings
}

type Embeddings struct {
	client *http.Client
	url    string
}

// embeddingResponse matches the Cloudflare Worker’s JSON response structure.
type embeddingResponse struct {
	Data    [][]float32 `json:"data"`
	Shape   []int       `json:"shape"`
	Pooling string      `json:"pooling"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
