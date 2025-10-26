package rag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateEmbeddingsSuccess(t *testing.T) {
	// Create a test server that returns embedding data
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Failed to unmarshal payload", http.StatusBadRequest)
			return
		}
		
		if payload["text"] != "test content" {
			http.Error(w, "Unexpected text in payload", http.StatusBadRequest)
			return
		}
		
		// Return embedding data
		response := embeddingResponse{
			Data:    [][]float32{expectedEmbedding},
			Shape:   []int{1, 5},
			Pooling: "mean",
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     5,
				CompletionTokens: 0,
				TotalTokens:      5,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a RAG instance with the test server
	rag := &RAG{
		Embeddings: Embeddings{
			client: &http.Client{},
			url:    server.URL,
		},
	}

	content := "test content"
	embeddings, err := rag.GenerateEmbeddings(content)
	if err != nil {
		t.Fatalf("GenerateEmbeddings returned error: %v", err)
	}

	if len(embeddings) != len(expectedEmbedding) {
		t.Errorf("Expected embedding length %d, got %d", len(expectedEmbedding), len(embeddings))
	}

	for i, v := range expectedEmbedding {
		if embeddings[i] != v {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, v, embeddings[i])
		}
	}
}

func TestGenerateEmbeddingsEmptyResponse(t *testing.T) {
	// Test server that returns empty embedding data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := embeddingResponse{
			Data:    [][]float32{{}}, // Empty embedding
			Shape:   []int{1, 0},
			Pooling: "mean",
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	rag := &RAG{
		Embeddings: Embeddings{
			client: &http.Client{},
			url:    server.URL,
		},
	}

	_, err := rag.GenerateEmbeddings("test content")
	if err == nil {
		t.Fatal("Expected error for empty embedding, got nil")
	}

	if !strings.Contains(err.Error(), "empty embedding returned") {
		t.Errorf("Expected error to contain 'empty embedding returned', got: %v", err)
	}
}

func TestGenerateEmbeddingsHTTPError(t *testing.T) {
	// Test server that returns a 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	rag := &RAG{
		Embeddings: Embeddings{
			client: &http.Client{},
			url:    server.URL,
		},
	}

	_, err := rag.GenerateEmbeddings("test content")
	if err == nil {
		t.Fatal("Expected error for 500 status code, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected error to contain 'unexpected status code: 500', got: %v", err)
	}
}

func TestGenerateEmbeddingsInvalidJSONResponse(t *testing.T) {
	// Test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"invalid": json`)
	}))
	defer server.Close()

	rag := &RAG{
		Embeddings: Embeddings{
			client: &http.Client{},
			url:    server.URL,
		},
	}

	_, err := rag.GenerateEmbeddings("test content")
	if err == nil {
		t.Fatal("Expected error for invalid JSON response, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected error to contain 'failed to decode response', got: %v", err)
	}
}

func TestGenerateEmbeddingsRequestError(t *testing.T) {
	rag := &RAG{
		Embeddings: Embeddings{
			client: &http.Client{},
			url:    "invalid-url", // Invalid URL to trigger error
		},
	}

	_, err := rag.GenerateEmbeddings("")
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}

	// The error could vary depending on the Go version, so we just check that we get an error
	if !strings.Contains(err.Error(), "invalid-url") && !strings.Contains(err.Error(), "unsupported protocol scheme") {
		t.Logf("Got expected error for invalid URL: %v", err)
	}
}

func TestMarshalEmbeddings(t *testing.T) {
	rag := &RAG{}
	
	embeddings := []float32{0.1, 0.2, 0.3}
	
	result, err := rag.Marshal(embeddings)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	
	// Verify the result is valid JSON
	var parsed []float32
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}
	
	if len(parsed) != len(embeddings) {
		t.Errorf("Expected length %d, got %d", len(embeddings), len(parsed))
	}
	
	for i, v := range embeddings {
		if parsed[i] != v {
			t.Errorf("Expected [%d] = %f, got %f", i, v, parsed[i])
		}
	}
}

func TestMarshalEmbeddingsError(t *testing.T) {
	rag := &RAG{}
	
	// Actually, JSON marshaling of valid float32 slices should not fail in normal cases
	// So this test is more about documenting that Marshal can potentially return an error
	embeddings := []float32{0.1, 0.2, 0.3}
	_, err := rag.Marshal(embeddings)
	if err != nil {
		// Standard []float32 should not cause marshal error
		t.Errorf("Marshal should not return error for valid embeddings: %v", err)
	}
}

func TestUnmarshalEmbeddings(t *testing.T) {
	rag := &RAG{}
	
	validJSON := `[0.1, 0.2, 0.3]`
	
	result, err := rag.Unmarshal(validJSON)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	
	expected := []float32{0.1, 0.2, 0.3}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected [%d] = %f, got %f", i, v, result[i])
		}
	}
}

func TestUnmarshalEmbeddingsError(t *testing.T) {
	rag := &RAG{}
	
	invalidJSON := `{"invalid": json`
	
	_, err := rag.Unmarshal(invalidJSON)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to unmarshal embeddings") {
		t.Errorf("Expected error to contain 'failed to unmarshal embeddings', got: %v", err)
	}
}

func TestUnmarshalEmbeddingsNonArray(t *testing.T) {
	rag := &RAG{}
	
	nonArrayJSON := `{"not": "an array"}`
	
	_, err := rag.Unmarshal(nonArrayJSON)
	if err == nil {
		t.Fatal("Expected error for non-array JSON, got nil")
	}
}