/*
Copyright © 2025 Erick Wachira (tesh254)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or modify the current configuration",
	Long: `Display or modify the currently loaded configuration settings.
This includes settings from the config file, environment variables, and defaults.
Sensitive information like API keys will be masked when displaying.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			displayConfig()
		} else {
			cmd.Help()
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Example: `  stick config set ai.apikey "your-api-key"
  stick config set provider "together"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		// Ensure we have a config file to write to
		if viper.ConfigFileUsed() == "" {
			// If no config file is loaded, we need to set a default one
			// This usually happens if the user hasn't created a config file yet
			// But since we are in the "set" command, we should probably create one
			// However, initConfig in root.go sets the config name/path but doesn't create it.
			// Let's rely on SafeWriteConfig or WriteConfig
		}

		err := viper.WriteConfig()
		if err != nil {
			// Try to write safely if it doesn't exist
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				err = viper.SafeWriteConfig()
			}
		}

		if err != nil {
			fmt.Printf("Error writing config: %v\n", err)
			return
		}

		displayValue := value
		if isSensitive(key) {
			displayValue = "****"
		}
		fmt.Printf("✅ Configuration updated: %s = %s\n", key, displayValue)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}

func displayConfig() {
	fmt.Println("Current Configuration:")
	fmt.Println("----------------------")

	file := viper.ConfigFileUsed()
	if file != "" {
		fmt.Printf("Config File: %s\n", file)
	} else {
		fmt.Println("Config File: None (using defaults/env)")
	}
	fmt.Println("----------------------")

	settings := viper.AllSettings()
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	printConfig(settings, "")
}

func printConfig(settings map[string]interface{}, prefix string) {
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := settings[k]
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		switch v := val.(type) {
		case map[string]interface{}:
			fmt.Printf("%s:\n", fullKey)
			printConfig(v, fullKey)
		case map[interface{}]interface{}:
			// Viper sometimes returns this for nested maps from YAML
			converted := make(map[string]interface{})
			for mk, mv := range v {
				converted[fmt.Sprintf("%v", mk)] = mv
			}
			fmt.Printf("%s:\n", fullKey)
			printConfig(converted, fullKey)
		default:
			if isSensitive(k) {
				valStr := fmt.Sprintf("%v", v)
				if len(valStr) > 4 {
					v = valStr[:4] + "****"
				} else {
					v = "****"
				}
			}
			fmt.Printf("%s: %v\n", fullKey, v)
		}
	}
}

func isSensitive(key string) bool {
	lowerKey := strings.ToLower(key)
	sensitiveKeywords := []string{"key", "secret", "password", "token", "auth"}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerKey, keyword) {
			return true
		}
	}
	return false
}
