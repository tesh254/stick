package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/vc/changes"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "check repository for changes",
	Example: `stick status`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		err := changes.PrintRepoStatus(repoPath)
		if err != nil {
			log.Error(fmt.Errorf("%w", err))
		}
	},
}
