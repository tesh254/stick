package crawl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func DefaultConfig() *CrawlConfig {
	return &CrawlConfig{
		UserAgent:    "Mozilla/5.0 (compatible; StickScraper/1.0)",
		Timeout:      60 * time.Second,
		RequestDelay: 3 * time.Second,
	}
}

func NewLlmText(url string, config *CrawlConfig) *LlmTxtContent {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	l := &LlmTxtContent{
		URL:             url,
		Config:          config,
		client:          client,
		lastRequestTime: make(map[string]time.Time),
	}

	return l
}

func NewPageHTMLToMarkdown(url string, config *CrawlConfig) *PageHTMLToMarkdown {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	h := &PageHTMLToMarkdown{
		URL:             url,
		Config:          config,
		client:          client,
		lastRequestTime: make(map[string]time.Time),
	}

	return h
}

func (l *LlmTxtContent) GetContent(url string) (string, error) {
	// Create a request with context and user agent
	ctx, cancel := context.WithTimeout(context.Background(), l.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", l.Config.UserAgent)

	// Make HTTP GET request
	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(bodyBytes), nil
}

func (h *PageHTMLToMarkdown) GetContent(url string) (string, error) {
	// Create a request with context and user agent
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", h.Config.UserAgent)

	// Make HTTP GET request
	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	htmlContent := string(bodyBytes)

	// Check content type - if it's an HTML file or HTML content, convert to markdown
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") || 
		strings.Contains(strings.ToLower(url), ".html") || 
		strings.Contains(strings.ToLower(url), ".htm") {
		// Convert HTML to markdown
		markdownContent, err := ToMarkdown(htmlContent)
		if err != nil {
			return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
		}
		return markdownContent, nil
	}

	// If not HTML content, return as is
	return htmlContent, nil
}
