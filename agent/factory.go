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

	endpoints := map[string]string{
		"together":   "https://api.together.xyz/v1/chat/completions",
		"openai":     "https://api.openai.com/v1/chat/completions",
		"openrouter": "https://openrouter.ai/api/v1/chat/completions",
		"atlascloud": "https://api.atlascloud.ai/v1/chat/completions",
	}

	if endpoint, ok := endpoints[providerName]; ok {
		return NewPlatformClient(endpoint, providerConfig.APIKey), nil
	} else {
		return nil, fmt.Errorf("unsupported AI provider: %s", providerName)
	}
}
