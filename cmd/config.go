package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manages application configuration",
	Long:  `Use this command to set and get configuration values.`,
}
