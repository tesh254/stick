package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/diff"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "show various types of objects",
	Example: `stick show --repo-path <directory> --commit <commit_hash> \n--repo-path value is optional`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		commit, _ := cmd.Flags().GetString("commit")
		if commit == "" {
			log.Error("please provide the commit hash")
			return
		}

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		jsonOutput, err := diff.ShowCommitDiff(repoPath, commit)
		if err != nil {
			log.Error(fmt.Sprintf("%v\n", err), "repo-path", repoPath, "commit", commit)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}
