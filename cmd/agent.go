package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/agent"
	"github.com/tesh254/stick/agent/message"
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

		responseChan := make(chan any)
		userInputChan := make(chan string)

		agentSession, err := agent.NewAgentSession(provider, responseChan, userInputChan)
		if err != nil {
			fmt.Println("Error creating agent session:", err)
			return
		}

		go agentSession.Run()
		agentSession.ProcessPrompt(prompt)

		for {
			response, ok := <-responseChan
			if !ok {
				break
			}

			switch msg := response.(type) {
			case string:
				if strings.HasPrefix(msg, "USER_INPUT_REQUEST:") {
					prompt := strings.TrimPrefix(msg, "USER_INPUT_REQUEST:")
					fmt.Printf("Agent: %s\n", prompt)
					fmt.Print("You: ")

					reader := bufio.NewReader(os.Stdin)
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("Error reading input:", err)
						break
					}
					userInputChan <- strings.TrimSpace(input)
				} else if msg == "AGENT_DONE" {
					return
				} else {
					fmt.Printf("Agent: %s\n", msg)
				}
			case message.AgentToolCallMsg:
				fmt.Printf("Tool Call: %s(%s)\n", msg.Name, msg.Args)
			case message.AgentToolResultMsg:
				if msg.IsError {
					fmt.Printf("Tool Error: %s\n", msg.Result)
				} else {
					fmt.Printf("Tool Result: %s\n", msg.Result)
				}
			}
		}
	},
}

func init() {
	agentCmd.AddCommand(agentCliCmd)
	agentCmd.AddCommand(agentInitCmd)
}
