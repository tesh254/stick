/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup initial stick configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
