package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/config"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Configure AI settings",
	Long:  `Configure AI settings such as the provider, model, and API key.`,
	Run: func(cmd *cobra.Command, args []string) {
		userConfig := config.GetConfig()
		fmt.Println(userConfig.Providers)
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set AI configuration",
	Long:  `Set the AI provider, model, and API key.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")
		apiKey, _ := cmd.Flags().GetString("apiKey")
		model, _ := cmd.Flags().GetString("model")

		if provider == "" {
			fmt.Println("Provider not specified. Use the --provider flag to set the provider.")
			return
		}

		if apiKey != "" {
			if err := config.Set(fmt.Sprintf("providers.%s.apiKey", provider), apiKey); err != nil {
				fmt.Println("Error setting apiKey:", err)
			}
			fmt.Println("API key set for provider:", provider)
		}

		if model != "" {
			if err := config.Set(fmt.Sprintf("providers.%s.model", provider), model); err != nil {
				fmt.Println("Error setting model:", err)
			}
			fmt.Println("Model set for provider:", provider)
		}
	},
}
