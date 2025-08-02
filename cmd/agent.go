package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/agent"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "interact with ai features",
	Example: `stick agent`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("this is where ai stuff will run")
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
