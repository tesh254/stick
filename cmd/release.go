package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/vc/release"
)

var releaseCmd = &cobra.Command{
	Use:     "release",
	Short:   "commit and push relase",
	Example: `stick release --repo-path <directory> --commit <commit_hash> --tag <tag_name> --message <commit_message>`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		commit, _ := cmd.Flags().GetString("commit")
		message, _ := cmd.Flags().GetString("message")
		tag, _ := cmd.Flags().GetString("tag")

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		err := release.CreateAndPushRelease(cmd.Context(), repoPath, tag, commit, message)
		if err != nil {
			log.Error(fmt.Errorf("%w", err))
		}
	},
}
