package agent

func getTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: &Function{
				Name:        "run_tool",
				Description: "run a command in the terminal",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"command": {
							Type:        "string",
							Description: "the command to run",
						},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "create_file",
				Description: "creates a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to create the file",
						},
						"content": {
							Type:        "string",
							Description: "the content of the file",
						},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "create_dir",
				Description: "creates a directory in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to create the directory",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "delete_dir",
				Description: "deletes a directory in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to delete the directory",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "delete_file",
				Description: "deletes a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to delete the file",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "read_file",
				Description: "reads a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to read the file",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "dir_tree",
				Description: "returns the directory tree of the current directory in json",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to read the directory tree",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "patch_file",
				Description: "patches a file in a given path",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "the path to patch the file",
						},
						"edits": {
							Type:        "array",
							Description: `list of edit operations to perform on the file.`,
							Items: Item{
								Type: "object",
								Properties: map[string]Prop{
									"action": {
										Type:        "string",
										Enum:        []string{"replace", "insert"},
										Description: `The type of edit operation: 'replace' to overwrite a line, 'insert' to add a new line.`,
									},
									"line_number": {
										Type:        "integer",
										Description: `The 1-based line number where the edit should occur.`,
									},
									"new_content": {
										Type:        "string",
										Description: `The content to insert or replace. Supports multi-line content with escaped newlines.`,
									},
								},
								Required: []string{"action", "line", "content"},
							},
						},
					},
					Required: []string{"path", "edits"},
				},
			},
		},
		{
			Type: "function",
			Function: &Function{
				Name:        "request_user_input",
				Description: "requests user input",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]Property{
						"prompt": {
							Type:        "string",
							Description: "the prompt to display to the user",
						},
					},
					Required: []string{"prompt"},
				},
			},
		},
	}
}
