package tools

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchURL fetches the content of a URL
func FetchURL(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Basic cleanup: remove scripts and styles (very naive)
	content := string(body)
	// TODO: Implement a proper HTML to text converter if needed.
	// For now, we return raw text but maybe truncate if too large?
	// Or just return it and let the agent deal with it.

	// Let's do a very basic "strip tags" by just removing <script> and <style> blocks content?
	// That's complex with regex. Let's just return the body for now.
	// The agent is an LLM, it can handle HTML often.
	// But let's limit size to avoid token explosion.

	if len(content) > 20000 {
		return content[:20000] + "\n... (truncated)", nil
	}

	return content, nil
}
