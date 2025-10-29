package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tesh254/stick/internal/studio"
)

// studioCmd starts the UI-friendly HTTP server with CORS
var studioCmd = &cobra.Command{
	Use:   "studio",
	Short: "Start the Stick Studio server (UI client interface)",
	Long:  "Runs a HTTP server providing API endpoints for UI integrations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := studio.LoadConfig()

		// Override via flags if set
		if portFlag := cmd.Flag("port"); portFlag != nil && portFlag.Value.String() != "" {
			cfg.Port = portFlag.Value.String()
		}
		if originsFlag := cmd.Flag("origins"); originsFlag != nil && originsFlag.Value.String() != "" {
			cfg.AllowedOrigins = originsFlag.Value.String()
		}
		if envFlag := cmd.Flag("env"); envFlag != nil && envFlag.Value.String() != "" {
			cfg.Env = envFlag.Value.String()
		}

		// Allow viper-based overrides for consistency
		if viper.IsSet("studio.port") {
			cfg.Port = viper.GetString("studio.port")
		}
		if viper.IsSet("studio.allowed_origins") {
			cfg.AllowedOrigins = viper.GetString("studio.allowed_origins")
		}
		if viper.IsSet("studio.env") {
			cfg.Env = viper.GetString("studio.env")
		}

		if err := studio.Start(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "studio server failed: %v\n", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(studioCmd)
	studioCmd.Flags().String("port", "5172", "Port for the studio server")
	studioCmd.Flags().String("origins", "*", "Allowed CORS origins (comma-separated)")
	studioCmd.Flags().String("env", "development", "Environment: development|production|test")
}
