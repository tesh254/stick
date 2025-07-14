package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/branch"
)

var createBranchCmd = &cobra.Command{
	Use:   "create [branch_name] [remote_name]",
	Short: "create a new virtual branch",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var remoteName *string
		var branchName *string
		if len(args) == 2 {
			branchName = &args[0]
			remoteName = &args[1]
		}
		branch.CreateVirtualBranch(branchName, remoteName)
	},
}
