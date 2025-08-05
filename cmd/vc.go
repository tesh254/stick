package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/vc"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "show various types of objects",
	Example: `stick show --repo-path <directory> --commit <commit_hash>`,
	Aliases: []string{"sh"},
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

		jsonOutput, err := vc.ShowCommitDiff(repoPath, commit)
		if err != nil {
			log.Error(fmt.Sprintf("%v\n", err), "repo-path", repoPath, "commit", commit)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

var releaseCmd = &cobra.Command{
	Use:     "release",
	Short:   "commit and push relase",
	Example: `stick release --repo-path <directory> --commit <commit_hash> --tag <tag_name> --message <commit_message>`,
	Aliases: []string{"rl"},
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		commit, _ := cmd.Flags().GetString("commit")
		message, _ := cmd.Flags().GetString("message")
		tag, _ := cmd.Flags().GetString("tag")

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		err := vc.CreateAndPushRelease(cmd.Context(), repoPath, tag, commit, message)
		if err != nil {
			log.Error(fmt.Errorf("%w", err))
		}
	},
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "check repository for changes",
	Aliases: []string{"st"},
	Example: `stick status`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		err := vc.PrintRepoStatus(repoPath)
		if err != nil {
			log.Error(fmt.Errorf("%w", err))
		}
	},
}

var diffCmd = &cobra.Command{
	Use:     "diff",
	Short:   "show diff for a commit",
	Aliases: []string{"df"},
	Example: `stick diff --repo-path <directory> --commit <commit_hash>`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if repoPath == "current" {
			repoPath, _ = os.Getwd()
		}

		if jsonOutput {
			jsonBytes, err := vc.DiffJSON(repoPath)
			if err != nil {
				log.Error(fmt.Errorf("%w", err))
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
		} else {
			err := vc.Diff(repoPath)
			if err != nil {
				log.Error(fmt.Errorf("%w", err))
			}
		}
	},
}
