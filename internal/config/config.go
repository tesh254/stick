package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// StickConfig holds configuration for the application.
type StickConfig struct {
	Name          string
	SelectedModel string
	APIKey        string
}

// NewStickConfig is a constructor for StickConfig that reads from Viper.
func NewStickConfig() *StickConfig {
	return &StickConfig{
		Name:          viper.GetString("name"),
		SelectedModel: viper.GetString("selected_model"),
		APIKey:        viper.GetString("api_key"),
	}
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

// SetAPIKey saves the API key to the config file.
func SetAPIKey(apiKey string) error {
	viper.Set("api_key", apiKey)
	return writeViperConfig()
}

// SetSelectedModel saves the selected model to the config file.
func SetSelectedModel(model string) error {
	viper.Set("selected_model", model)
	return writeViperConfig()
}

// GetAPIKey returns the API key from the config file.
func GetAPIKey() string {
	return viper.GetString("api_key")
}

// GetSelectedModel returns the selected model from the config file.
func GetSelectedModel() string {
	return viper.GetString("selected_model")
}

// GetName returns the name from the config file.
func GetName() string {
	return viper.GetString("name")
}
