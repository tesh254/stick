package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/vbranch"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialise stick in current directory",
		Run: func(cmd *cobra.Command, args []string) {
			vbranch.Init()
		},
	}
}

func branchCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	var branchCmd = &cobra.Command{
		Use:   "branch",
		Short: "manage virtual branches",
	}

	branchCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list all virtual branches",
		Run: func(cmd *cobra.Command, args []string) {
			vbranch.ListBranches()
		},
	})

	branchCmd.AddCommand(&cobra.Command{
		Use:   "create [name]",
		Short: "create a new virtual branch",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("please provide a branch name")
				return
			}
			name := args[0]
			vbranch.CreateBranch(name)
		},
	})

	branchCmd.AddCommand(&cobra.Command{
		Use:   "switch [name]",
		Short: "switch to a virtual branches",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("please provide a virtual branch name")
				return
			}
			name := args[0]
			vbranch.SwitchBranch(name, args)
		},
	})

	return branchCmd
}

func statusCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "status",
		Short: "show status of virtual branches and changes",
		Run: func(cmd *cobra.Command, args []string) {
			vbranch.Status()
		},
	}
}

func addCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	cmd := &cobra.Command{
		Use:   "add [file...]",
		Short: "add file changes to current virtual branch",
		Run: func(cmd *cobra.Command, args []string) {
			vbranch.AddFile(cmd, args)
		},
	}
	cmd.Flags().BoolP("all", "A", false, "Add all changes")
	return cmd
}

func moveCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "move [hunk-id] [target-branch]",
		Short: "Move a change hunk to another virtual branch",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				fmt.Println("please provide a hunk id and a target branch")
				return
			}
			vbranch.MoveHunkToTargetBranch(args[0], args[1])
		},
	}
}

func pushCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "push [branch-name]",
		Short: "push virtual branch to remote as a Git branch",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var branchName *string
			if len(args) == 0 {
				branchName = nil
			} else {
				branchName = &args[0]
			}
			vbranch.PushBranchToRemoteAsGitBranch(branchName)
		},
	}
}

func applyCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "apply [branch-name]",
		Short: "apply virtual branch changes to working directory",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var branchName *string
			if len(args) == 0 {
				branchName = nil
			} else {
				branchName = &args[0]
			}
			vbranch.ApplyVBranchChangesToWorkingDir(branchName)
		},
	}
}

func unapplyCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "unapply [branch-name]",
		Short: "remove virtual branch changes from working directory",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var branchName *string
			if len(args) == 0 {
				branchName = nil
			} else {
				branchName = &args[0]
			}
			vbranch.UnapplyVBranchChangesToWorkingDir(branchName)
		},
	}
}

func syncCmd() *cobra.Command {
	vbranch.EnsureStateInitialized()
	return &cobra.Command{
		Use:   "sync",
		Short: "sync virtual branches with Git repository state",
		Run: func(cmd *cobra.Command, args []string) {
			vbranch.SyncBranchesWithGitRepoState()
		},
	}
}
