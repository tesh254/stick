package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMaskAPIKey(t *testing.T) {
	assert.Equal(t, "********ikey", MaskAPIKey("someapikey"))
	assert.Equal(t, "********", MaskAPIKey("key"))
	assert.Equal(t, "********", MaskAPIKey(""))
}

func TestProviderConfig_String(t *testing.T) {
	pc := ProviderConfig{
		APIKey: "test-api-key",
		Model:  "test-model",
	}
	expected := "  Model: test-model\n  APIKey: ********-key"
	assert.Equal(t, expected, pc.String())
}

func TestStickConfig_String(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("name", "test-name")
	v.Set("defaultProvider", "together")
	v.Set("providers.together.apiKey", "together-api-key")
	v.Set("providers.together.model", "together-model")
	v.Set("providers.anthropic.apiKey", "anthropic-api-key")
	v.Set("providers.anthropic.model", "anthropic-model")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	config := NewStickConfig()

	output := config.String()
	assert.Contains(t, output, "Name: test-name")
	assert.Contains(t, output, "Default Provider: together")
	assert.Contains(t, output, `  anthropic:
  Model: anthropic-model
  APIKey: ********-key`)
	assert.Contains(t, output, "  together:\n  Model: together-model\n  APIKey: ********-key")
}

func TestGetConfig(t *testing.T) {

	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("name", "test-name")
	v.Set("defaultProvider", "together")
	v.Set("providers.together.apiKey", "together-api-key")
	v.Set("providers.together.model", "together-model")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	config := GetConfig()

	assert.Equal(t, "test-name", config.Name)
	assert.Equal(t, "together", config.DefaultProvider)
	assert.Equal(t, "together-api-key", config.Providers["together"].APIKey)
	assert.Equal(t, "together-model", config.Providers["together"].Model)
}

func TestNewStickConfig(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("name", "test-name")
	v.Set("defaultProvider", "together")
	v.Set("providers.together.apiKey", "together-api-key")
	v.Set("providers.together.model", "together-model")
	v.Set("providers.anthropic.apiKey", "anthropic-api-key")
	v.Set("providers.anthropic.model", "anthropic-model")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	config := NewStickConfig()

	assert.Equal(t, "test-name", config.Name)
	assert.Equal(t, "together", config.DefaultProvider)
	assert.Equal(t, "together-api-key", config.Providers["together"].APIKey)
	assert.Equal(t, "together-model", config.Providers["together"].Model)
	assert.Equal(t, "anthropic-api-key", config.Providers["anthropic"].APIKey)
	assert.Equal(t, "anthropic-model", config.Providers["anthropic"].Model)
}

func TestGetProviderConfig(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("defaultProvider", "together")
	v.Set("providers.together.apiKey", "together-api-key")
	v.Set("providers.together.model", "together-model")
	v.Set("providers.anthropic.apiKey", "anthropic-api-key")
	v.Set("providers.anthropic.model", "anthropic-model")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	// Test with default provider
	cfg, name, err := GetProviderConfig()
	assert.NoError(t, err)
	assert.Equal(t, "together", name)
	assert.Equal(t, "together-api-key", cfg.APIKey)
	assert.Equal(t, "together-model", cfg.Model)

	// Test with specific provider
	cfg, name, err = GetProviderConfig("anthropic")
	assert.NoError(t, err)
	assert.Equal(t, "anthropic", name)
	assert.Equal(t, "anthropic-api-key", cfg.APIKey)
	assert.Equal(t, "anthropic-model", cfg.Model)

	// Test with non-existent provider
	cfg, name, err = GetProviderConfig("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "", name)
	assert.Contains(t, err.Error(), "configuration for provider 'nonexistent' not found")

	// Test with no default and no provider specified
	v.Set("defaultProvider", "")
	err = v.WriteConfig()
	assert.NoError(t, err)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	cfg, name, err = GetProviderConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "", name)
	assert.Contains(t, err.Error(), "no AI provider specified and no default provider is set")
}

func TestSet(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	err = v.WriteConfig() // Ensure config file exists
	assert.NoError(t, err)

	// Set a value
	err = Set("test.key", "test-value")
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	assert.Equal(t, "test-value", viper.GetString("test.key"))

	// Set a provider specific value
	err = Set("providers.newprovider.apiKey", "new-api-key")
	assert.NoError(t, err)

	err = viper.ReadInConfig()
	assert.NoError(t, err)

	cfg, name, err := GetProviderConfig("newprovider")
	assert.NoError(t, err)
	assert.Equal(t, "newprovider", name)
	assert.Equal(t, "new-api-key", cfg.APIKey)
}

func TestGetString(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("some.string.key", "expected-string")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	assert.Equal(t, "expected-string", GetString("some.string.key"))
	assert.Equal(t, "", GetString("non.existent.key"))
}
