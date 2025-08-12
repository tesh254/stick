package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/agent"
	"github.com/tesh254/stick/tui"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "interact with ai features",
	Example: `stick agent`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")
		tui.Run(provider)
	},
}

var agentInitCmd = &cobra.Command{
	Use:     "init",
	Short:   "setup agent settings",
	Example: `stick agent init`,
	Run: func(cmd *cobra.Command, args []string) {
		agent.SelectModel()
	},
}

var agentCliCmd = &cobra.Command{
	Use:     "cli",
	Short:   "interact with ai features in cli mode",
	Example: `stick agent cli "what is the meaning of life?"`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]
		provider, _ := cmd.Flags().GetString("provider")

		responseChan := make(chan string)
		userInputChan := make(chan string)
		defer close(userInputChan)

		go agent.RunAgent(prompt, provider, responseChan, userInputChan)

		for {
			response, ok := <-responseChan
			if !ok {
				break
			}

			if strings.HasPrefix(response, "USER_INPUT_REQUEST:") {
				prompt := strings.TrimPrefix(response, "USER_INPUT_REQUEST:")
				fmt.Printf("Agent: %s\n", prompt)
				fmt.Print("You: ")

				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input:", err)
					break
				}
				userInputChan <- strings.TrimSpace(input)
			} else {
				fmt.Printf("Agent: %s\n", response)
			}
		}
	},
}

func init() {
	agentCmd.AddCommand(agentCliCmd)
}
