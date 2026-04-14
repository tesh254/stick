package functions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/tesh254/stick/internal/crawl"
)

// SetProvider: args ["platform", "model", "api_key"] -> sets the provider configuration
func SetProvider(args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("requires 3 arguments: platform, model, api_key")
	}

	platform := strings.TrimSpace(unquoteIfQuoted(args[0]))
	model := strings.TrimSpace(unquoteIfQuoted(args[1]))
	apiKey := strings.TrimSpace(unquoteIfQuoted(args[2]))

	// Update viper config
	viper.Set(fmt.Sprintf("providers.%s.apikey", platform), apiKey)
	viper.Set(fmt.Sprintf("providers.%s.model", platform), model)
	
	// Also set as current provider if desired, or just set the specific provider settings.
	// The user asked to "set a new custom provider", implying they want to use it.
	// So we should probably set `ai.platform` or `provider` to this platform.
	viper.Set("ai.platform", platform)
	// We might also want to set global model/apikey as fallback, but provider-specific is better.
	
	// Save config
	err := viper.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfig()
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return fmt.Sprintf("Provider configured: %s (model: %s). Configuration saved.", platform, model), nil
}

// Add: args ["1","2"] -> sum numerically; missing args default to 0
func Add(args []string) (string, error) {
	var a, b float64
	var err error

	if len(args) >= 1 {
		a, err = parseNumber(args[0])
		if err != nil {
			return "", fmt.Errorf("add arg1: %w", err)
		}
	}
	if len(args) >= 2 {
		b, err = parseNumber(args[1])
		if err != nil {
			return "", fmt.Errorf("add arg2: %w", err)
		}
	}
	return fmt.Sprintf("%g", a+b), nil
}

func parseNumber(s string) (float64, error) {
	s = strings.TrimSpace(unquoteIfQuoted(s))
	if s == "" {
		return 0, nil // Return 0 for empty strings as per the function's spec
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return f, nil
}

func unquoteIfQuoted(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		// Check for double quotes
		if s[0] == '"' && s[len(s)-1] == '"' {
			u, err := strconv.Unquote(s)
			if err == nil {
				return u
			}
		} else if s[0] == '\'' && s[len(s)-1] == '\'' { // Check for single quotes - simple removal
			return s[1 : len(s)-1] // Remove first and last characters
		}
	}
	return s
}

// Echo: args ["text1", "text2", ...] -> returns concatenated text with quotes removed from arguments
func Echo(args []string) (string, error) {
	// Remove quotes from each argument
	unquotedArgs := make([]string, len(args))
	for i, arg := range args {
		unquotedArgs[i] = unquoteIfQuoted(arg)
	}
	return strings.Join(unquotedArgs, " "), nil
}

// GetLLMText: fetch llm text content to and then we can look into it
func GetLLMText(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no arguments provided")
	}

	// Remove quotes from each argument
	config := crawl.DefaultConfig()
	textCrawler := crawl.NewLlmText(args[0], config)

	return textCrawler.GetContent(args[0])
}

// GetPageHTMLContentToMarkdown: fetch page html content and then convert it to markdown
func GetPageHTMLContentToMarkdown(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no arguments provided")
	}

	// Remove quotes from each argument
	config := crawl.DefaultConfig()
	htmlCrawler := crawl.NewPageHTMLToMarkdown(args[0], config)

	return htmlCrawler.GetContent(args[0])
}
