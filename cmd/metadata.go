package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/metadata"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Stick in the current repository",
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure it’s a Git repository
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			fmt.Println("not a git repository. initializing...")
			_, err := git.PlainInit(".", false)
			if err != nil {
				fmt.Println("failed to initialize git repository:", err)
				return
			}
		}

		// Create .stick directory and metadata file
		err := os.MkdirAll(".stick", 0755)
		if err != nil {
			fmt.Println("Failed to create .stick directory:", err)
			return
		}

		md := metadata.Metadata{VirtualBranches: make(map[string]metadata.VirtualBranch)}
		metadata.SaveMetadata(md)
		fmt.Println("Stick initialized.")
	},
}
