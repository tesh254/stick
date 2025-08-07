package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/agent"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "interact with ai features",
	Example: `stick agent`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")
		agent.RunAgent("print the current directory", provider)
	},
}

var agentInitCmd = &cobra.Command{
	Use:     "init",
	Short:   "setup agent settings",
	Example: `stick agent init`,
	Run: func(cmd *cobra.Command, args []string) {
		agent.SelectModel()
	},
}
