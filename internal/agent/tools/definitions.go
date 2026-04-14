package tools

import "github.com/tesh254/stick/internal/provider"

func GetTools() []provider.Tool {
	return []provider.Tool{
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "execute_command",
				Description: "Execute a shell command on the user's machine. Use this to run builds, tests, etc. PREFER specific file tools (read_file, list_files) over shell commands like 'cat' or 'ls' for better reliability.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"command": {
							Type:        "string",
							Description: "The command to execute (e.g., 'go test ./...').",
						},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "read_file",
				Description: "Read the content of a file.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"path": {
							Type:        "string",
							Description: "The absolute or relative path to the file.",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "write_file",
				Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"path": {
							Type:        "string",
							Description: "The absolute or relative path to the file.",
						},
						"content": {
							Type:        "string",
							Description: "The content to write.",
						},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "list_files",
				Description: "List files and directories in a path.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"path": {
							Type:        "string",
							Description: "The path to list (defaults to current directory if empty).",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "search_files",
				Description: "Search for a string pattern in files within a directory.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"path": {
							Type:        "string",
							Description: "The directory to search in.",
						},
						"pattern": {
							Type:        "string",
							Description: "The string pattern to search for.",
						},
					},
					Required: []string{"path", "pattern"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "fetch_url",
				Description: "Fetch the text content of a URL.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"url": {
							Type:        "string",
							Description: "The URL to fetch.",
						},
					},
					Required: []string{"url"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "create_task_slice",
				Description: "Create a list of tasks to track progress.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"tasks": {
							Type:        "array",
							Description: "List of task descriptions.",
							Items: &provider.Item{
								Type: "string",
							},
						},
					},
					Required: []string{"tasks"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "update_task_status",
				Description: "Update the status of a task.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"index": {
							Type:        "integer",
							Description: "Index of the task to update.",
						},
						"is_done": {
							Type:        "boolean",
							Description: "Whether the task is done.",
						},
					},
					Required: []string{"index", "is_done"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "task_complete",
				Description: "Signal that all tasks are completed.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"message": {
							Type:        "string",
							Description: "Completion message.",
						},
					},
					Required: []string{"message"},
				},
			},
		},
		{
			Type: "function",
			Function: &provider.FunctionDefinition{
				Name:        "switch_mode",
				Description: "Switch the agent's operating mode.",
				Parameters: &provider.Parameters{
					Type: "object",
					Properties: map[string]provider.Property{
						"mode": {
							Type:        "string",
							Description: "The mode to switch to. Allowed values: 'coding', 'planning', 'debugging', 'architect'.",
						},
					},
					Required: []string{"mode"},
				},
			},
		},
	}
}
