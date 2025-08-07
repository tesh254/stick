package agent

import (
	"fmt"

	"github.com/tesh254/stick/internal/shell"
)

func ExecuteTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "run_tool":
		command, ok := args["command"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for run_tool")
		}
		cmd := shell.ExecuteCommand(command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return string(output), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
