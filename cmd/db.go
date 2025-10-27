/*
Copyright © 2025 Erick Wachira (tesh254)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dbtui "github.com/tesh254/stick/internal/tui/db"
)

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Interact with the local database",
	Long:  `Provides access to database records including conversations, messages, and usage information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use subcommands to interact with the database: list, view, etc.")
	},
}

// listCmd represents the list subcommand
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List database records",
	Long:  `List different types of records in the database.`,
}

// listConversationsCmd represents the list conversations subcommand
var listConversationsCmd = &cobra.Command{
	Use:   "conversations",
	Short: "List all conversations",
	Long:  `Displays all conversations stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		listConversations()
	},
}

// listMessagesCmd represents the list messages subcommand
var listMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "List all messages",
	Long:  `Displays all messages stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		listMessages()
	},
}

// listUsageCmd represents the list usage subcommand
var listUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "List all usage records",
	Long:  `Displays all usage records stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		listUsage()
	},
}

// viewCmd represents the view subcommand
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View specific database records",
	Long:  `View details of specific records in the database.`,
}

// viewConversationCmd represents the view conversation subcommand
var viewConversationCmd = &cobra.Command{
	Use:   "conversation [id]",
	Short: "View a specific conversation and its messages",
	Long:  `Displays details of a specific conversation and all its messages.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewConversation(args[0])
	},
}

// init initializes the db command and its subcommands
func init() {
	// Add db command to root
	rootCmd.AddCommand(dbCmd)

	// Add subcommands to db
	dbCmd.AddCommand(listCmd)
	dbCmd.AddCommand(viewCmd)

	// Add subcommands to list
	listCmd.AddCommand(listConversationsCmd)
	listCmd.AddCommand(listMessagesCmd)
	listCmd.AddCommand(listUsageCmd)

	// Add subcommands to view
	viewCmd.AddCommand(viewConversationCmd)
}

// listConversations lists all conversations using TUI
func listConversations() {
	// Create and run the conversation list TUI
	model, err := dbtui.NewConversationListModel()
	if err != nil {
		fmt.Printf("Error creating conversation list TUI: %v\n", err)
		return
	}
	if err := model.Run(); err != nil {
		fmt.Printf("Error running conversation list TUI: %v\n", err)
	}
}

// listMessages lists all messages using TUI
func listMessages() {
	fmt.Println("Messages listing will be implemented with TUI")
}

// listUsage lists all usage records using TUI
func listUsage() {
	fmt.Println("Usage listing will be implemented with TUI")
}

// viewConversation displays a specific conversation and its messages using TUI
func viewConversation(id string) {
	fmt.Printf("Viewing conversation with ID: %s\n", id)
}