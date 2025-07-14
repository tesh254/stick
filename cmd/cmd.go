package cmd

import "github.com/spf13/cobra"

var stickCmd = &cobra.Command{
	Use:   "stick",
	Short: "Stick is a lightweight CLI tool for managing multiple virtual branches in Git, enabling seamless work on different features without branch switching.",
}
