package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads the content of a file
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes content to a file
func WriteFile(path, content string) (string, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return "File written successfully.", nil
}

// ListFiles lists files in a directory (non-recursive for safety/speed, or shallow recursive)
func ListFiles(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		isDir := " "
		if entry.IsDir() {
			isDir = "/"
		}
		sb.WriteString(fmt.Sprintf("%s%s (%d bytes)\n", entry.Name(), isDir, info.Size()))
	}
	return sb.String(), nil
}

// SearchFiles searches for a pattern in files using grep or similar
// We use a simple walk here to avoid shelling out if possible, but grep is faster.
// Let's use the existing ExecuteCommand but wrapper for safety?
// Actually, Go implementation is safer than shell injection.
func SearchFiles(path, pattern string) (string, error) {
	// Using grep via exec is often most efficient for search
	// strict security: ensure path is safe?
	// For now, let's use a pure Go walk to be safe and portable
	var matches []string
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), pattern) {
			matches = append(matches, filePath)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "No matches found.", nil
	}
	return strings.Join(matches, "\n"), nil
}
