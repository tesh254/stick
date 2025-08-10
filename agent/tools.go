package agent

import (
	"fmt"
	"os"

	"github.com/tesh254/ffs/core"
	"github.com/tesh254/ffs/ffs"
	"github.com/tesh254/stick/internal/shell"
)

type Edit struct {
	Action  string `json:"action"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

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
	case "create_file":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for create_file")
		}
		content, ok := args["content"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for create_file")
		}
		err := core.WriteFile(path, []byte(content))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("created file %s", path), nil
	case "create_dir":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for create_dir")
		}
		err := core.CreateDir(path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("created dir %s", path), nil
	case "delete_dir":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for delete_dir")
		}
		err := core.DeleteDir(path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deleted dir %s", path), nil
	case "delete_file":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for delete_file")
		}
		err := core.DeleteFile(path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deleted file %s", path), nil
	case "read_file":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for read_file")
		}
		content, err := core.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(content), nil
	case "dir_tree":
		fs := ffs.New()
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current dir: %w", err)
		}
		dir := fs.Dir(currentDir)
		tree, err := dir.Tree(nil, []string{".git"})
		if err != nil {
			return "", fmt.Errorf("failed to get dir tree: %w", err)
		}
		minified, err := core.GetTreeMinifiedJSON(tree)
		if err != nil {
			return "", fmt.Errorf("failed to minify dir tree: %w", err)
		}
		return minified, nil
	case "patch_file":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid arguments for path")
		}
		edits, ok := args["edits"].([]core.EditInstruction)
		if !ok {
			return "", fmt.Errorf("invalid arguments for edits")
		}
		fmt.Println(edits, path)
		fileRequest := core.FileEditRequest{
			FilePath: path,
			Edits:    edits,
		}
		err := core.ApplyPatch(fileRequest, true, true, true)
		if err != nil {
			return "", fmt.Errorf("failed to apply patch to file %s: %w", path, err)
		}
		return fmt.Sprintf("applied patch to file %s any more edits or any other file?", path), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
