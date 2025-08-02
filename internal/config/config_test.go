package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetters(t *testing.T) {
	// Set up a temporary config file for testing
	file, err := os.CreateTemp("", "stick-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Use the global viper instance for testing
	v := viper.GetViper()
	v.SetConfigFile(file.Name())
	v.Set("name", "test-name")
	v.Set("selected_model", "test-model")
	v.Set("api_key", "test-api-key")
	err = v.WriteConfig()
	assert.NoError(t, err)

	// Read the config back into the global viper instance
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	assert.Equal(t, "test-name", GetName())
	assert.Equal(t, "test-model", GetSelectedModel())
	assert.Equal(t, "test-api-key", GetAPIKey())
}