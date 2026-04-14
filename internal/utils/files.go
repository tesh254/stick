package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ListFiles recursively lists files in the current directory, ignoring hidden files and directories.
func ListFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		rel, err := filepath.Rel(root, path)
		if err == nil {
			files = append(files, rel)
		} else {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
