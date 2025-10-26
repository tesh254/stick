package crawl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetContentSuccess(t *testing.T) {
	// Create a test server that returns plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "This is sample text content for testing.")
	}))
	defer server.Close()

	// Create a new LlmTxtContent instance
	config := DefaultConfig()
	l := NewLlmText(server.URL, config)

	// Test with the test server URL
	content, err := l.GetContent(server.URL)
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}

	expected := "This is sample text content for testing."
	if content != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, content)
	}

	if len(content) == 0 {
		t.Error("GetContent returned empty content")
	}
}

func TestGetContentWithConfig(t *testing.T) {
	// Test with custom configuration
	config := &CrawlConfig{
		UserAgent: "TestUserAgent/1.0",
		Timeout:   10 * time.Second,
	}

	l := NewLlmText("https://example.com", config)

	if l.Config.UserAgent != "TestUserAgent/1.0" {
		t.Errorf("Expected UserAgent 'TestUserAgent/1.0', got '%s'", l.Config.UserAgent)
	}

	if l.Config.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout 10s, got %v", l.Config.Timeout)
	}
}

func TestGetContentWithDefaultConfig(t *testing.T) {
	// Test with nil config (should use default)
	l := NewLlmText("https://example.com", nil)

	defaultConfig := DefaultConfig()
	if l.Config.UserAgent != defaultConfig.UserAgent {
		t.Errorf("Expected UserAgent '%s', got '%s'", defaultConfig.UserAgent, l.Config.UserAgent)
	}
}

func TestGetContentHTTPError(t *testing.T) {
	// Test with a URL that returns a 404 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	config := DefaultConfig()
	l := NewLlmText(server.URL, config)

	_, err := l.GetContent(server.URL)
	if err == nil {
		t.Fatal("Expected error for 404 status code, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 404") {
		t.Errorf("Expected error to contain 'unexpected status code: 404', got: %v", err)
	}
}

func TestGetContentInvalidURLError(t *testing.T) {
	config := DefaultConfig()
	l := NewLlmText("", config)

	// Test with an invalid URL
	_, err := l.GetContent("://invalid-url")
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("Expected error to contain 'failed to create request', got: %v", err)
	}
}

func TestGetContentTimeout(t *testing.T) {
	// Test with a server that takes too long to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		<-ctx.Done() // Wait for context to be cancelled
	}))
	defer server.Close()

	// Create a config with a short timeout
	config := &CrawlConfig{
		UserAgent: "TestUserAgent/1.0",
		Timeout:   100 * time.Millisecond, // Very short timeout
	}

	l := NewLlmText(server.URL, config)

	_, err := l.GetContent(server.URL)
	if err == nil {
		t.Fatal("Expected error for timeout, got nil")
	}

	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("Expected error to contain 'failed to fetch', got: %v", err)
	}
}

func TestGetContentReadBodyError(t *testing.T) {
	// Create a mock client that returns a response that fails when reading body
	// We'll use httptest to simulate a server that closes connection mid-stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		// Intentionally not writing anything or closing connection properly
	}))
	defer server.Close()

	config := DefaultConfig()
	l := NewLlmText(server.URL, config)

	_, err := l.GetContent(server.URL)
	if err != nil {
		// The error could be related to connection issues, which is expected
		t.Logf("Got expected error when reading body: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.UserAgent != "Mozilla/5.0 (compatible; StickScraper/1.0)" {
		t.Errorf("Expected default UserAgent 'Mozilla/5.0 (compatible; StickScraper/1.0)', got '%s'", config.UserAgent)
	}
	
	if config.Timeout != 60*time.Second {
		t.Errorf("Expected default Timeout 60s, got %v", config.Timeout)
	}
	
	if config.RequestDelay != 3*time.Second {
		t.Errorf("Expected default RequestDelay 3s, got %v", config.RequestDelay)
	}
}

func TestNewPageHTMLToMarkdown(t *testing.T) {
	config := &CrawlConfig{
		UserAgent: "TestUserAgent/1.0",
		Timeout:   10 * time.Second,
	}

	h := NewPageHTMLToMarkdown("https://example.com", config)

	if h.Config.UserAgent != "TestUserAgent/1.0" {
		t.Errorf("Expected UserAgent 'TestUserAgent/1.0', got '%s'", h.Config.UserAgent)
	}

	if h.Config.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout 10s, got %v", h.Config.Timeout)
	}
}

func TestPageHTMLToMarkdownGetContentHTML(t *testing.T) {
	// Create a test server that returns HTML content
	htmlContent := `<html><body><h1>Test Title</h1><p>This is a test paragraph.</p></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlContent)
	}))
	defer server.Close()

	config := DefaultConfig()
	h := NewPageHTMLToMarkdown(server.URL, config)

	content, err := h.GetContent(server.URL)
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}

	// Check that the HTML was converted to markdown
	expectedMarkdown := "# Test Title\n\nThis is a test paragraph."
	if !strings.Contains(content, expectedMarkdown) {
		t.Errorf("Expected content to contain markdown '%s', got '%s'", expectedMarkdown, content)
	}
}

func TestPageHTMLToMarkdownGetContentPlainText(t *testing.T) {
	// Create a test server that returns plain text content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "This is plain text content.")
	}))
	defer server.Close()

	config := DefaultConfig()
	h := NewPageHTMLToMarkdown(server.URL, config)

	content, err := h.GetContent(server.URL)
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}

	// Plain text should be returned as is (not converted)
	expected := "This is plain text content."
	if content != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, content)
	}
}

func TestPageHTMLToMarkdownGetContentHTMLFile(t *testing.T) {
	// Create a test server that returns HTML content without explicit content-type
	htmlContent := `<html><body><h1>HTML File Test</h1><p>Testing .html file conversion.</p></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No explicit content-type header
		fmt.Fprint(w, htmlContent)
		// Simulate URL that ends with .html by checking the request path
		if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}
	}))
	defer server.Close()

	// Use a fake URL ending with .html for the client
	fakeURL := server.URL + "/test.html"
	config := DefaultConfig()
	h := NewPageHTMLToMarkdown(fakeURL, config)

	content, err := h.GetContent(fakeURL)
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}

	// Check that the HTML was converted to markdown (based on URL ending with .html)
	expectedMarkdown := "# HTML File Test"
	if !strings.Contains(content, expectedMarkdown) {
		t.Errorf("Expected content to contain markdown '%s', got '%s'", expectedMarkdown, content)
	}
}

func TestPageHTMLToMarkdownGetContentError(t *testing.T) {
	// Test with a URL that returns a 404 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	config := DefaultConfig()
	h := NewPageHTMLToMarkdown(server.URL, config)

	_, err := h.GetContent(server.URL)
	if err == nil {
		t.Fatal("Expected error for 404 status code, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 404") {
		t.Errorf("Expected error to contain 'unexpected status code: 404', got: %v", err)
	}
}

func TestToMarkdown(t *testing.T) {
	htmlContent := `<h1>Test Heading</h1><p>This is a paragraph with <strong>bold text</strong>.</p>`
	
	markdown, err := ToMarkdown(htmlContent)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	
	if !strings.Contains(markdown, "# Test Heading") {
		t.Errorf("Expected markdown to contain '# Test Heading', got '%s'", markdown)
	}
	
	if !strings.Contains(markdown, "**bold text**") {
		t.Errorf("Expected markdown to contain '**bold text**', got '%s'", markdown)
	}
}
