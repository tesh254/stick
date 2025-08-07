package agent

import (
	"fmt"

	"github.com/tesh254/stick/internal/config"
)

// NewAIClient creates a new AI client based on the configuration.
func NewAIClient(provider ...string) (AIClient, error) {
	providerConfig, providerName, err := config.GetProviderConfig(provider...)
	if err != nil {
		return nil, err
	}

	switch providerName {
	case "together":
		return NewTogetherClient(providerConfig.APIKey)
	// case "anthropic":
	// 	return NewAnthropicClient(providerConfig.APIKey)
	// case "google":
	// 	return NewGoogleClient(providerConfig.APIKey)
	// case "openai":
	// 	return NewOpenAIClient(providerConfig.APIKey)
	// case "openrouter":
	// 	return NewOpenRouterClient(providerConfig.APIKey)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", providerName)
	}
}
