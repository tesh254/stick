package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// ProviderConfig represents the configuration for a single AI provider.
type ProviderConfig struct {
	APIKey string `mapstructure:"apiKey"`
	Model  string `mapstructure:"model"`
}

// maskAPIKey is a helper function to mask the API key for printing.
func MaskAPIKey(apiKey string) string {
	if len(apiKey) > 4 {
		return "********" + apiKey[len(apiKey)-4:]
	}
	return "********"
}

// String returns a string representation of ProviderConfig with APIKey masked.
func (pc ProviderConfig) String() string {
	return fmt.Sprintf("  Model: %s\n  APIKey: %s", pc.Model, MaskAPIKey(pc.APIKey))
}

// StickConfig holds configuration for the application.
type StickConfig struct {
	Name            string                    `mapstructure:"name"`
	DefaultProvider string                    `mapstructure:"defaultProvider"`
	Providers       map[string]ProviderConfig `mapstructure:"providers"`
}

// String returns a string representation of StickConfig.
func (sc StickConfig) String() string {
	s := fmt.Sprintf("Name: %s\nDefault Provider: %s\nProviders:\n", sc.Name, sc.DefaultProvider)
	for name, provider := range sc.Providers {
		s += fmt.Sprintf("  %s:\n%s\n", name, provider.String())
	}
	return s
}

// GetConfig returns the entire StickConfig object.
func GetConfig() *StickConfig {
	return NewStickConfig()
}

// NewStickConfig is a constructor for StickConfig that reads from Viper.
func NewStickConfig() *StickConfig {
	var config StickConfig
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshalling config: %s", err)
	}
	return &config
}

func writeViperConfig() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile = fmt.Sprintf("%s/.stick.yaml", home)
	}
	return viper.WriteConfigAs(configFile)
}

// Set saves a key-value pair to the config file.
func Set(key string, value interface{}) error {
	viper.Set(key, value)
	return writeViperConfig()
}

// GetString returns a string value from the config file.
func GetString(key string) string {
	return viper.GetString(key)
}

// GetProviderConfig returns the configuration for a specific provider.
func GetProviderConfig(provider ...string) (*ProviderConfig, string, error) {
	var providerName string
	if len(provider) > 0 && provider[0] != "" {
		providerName = provider[0]
	} else {
		providerName = viper.GetString("defaultProvider")
	}

	if providerName == "" {
		return nil, "", fmt.Errorf("no AI provider specified and no default provider is set")
	}

	var providers map[string]ProviderConfig
	if err := viper.UnmarshalKey("providers", &providers); err != nil {
		return nil, "", fmt.Errorf("error unmarshalling providers config: %w", err)
	}

	config, ok := providers[providerName]
	if !ok {
		return nil, "", fmt.Errorf("configuration for provider '%s' not found", providerName)
	}

	return &config, providerName, nil
}
