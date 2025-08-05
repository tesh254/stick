package vc

import (
	"encoding/json"
	"fmt"
)

func Diff(repoPath string) error {
	changes, err := getWorktreeChanges(repoPath)
	if err != nil {
		return fmt.Errorf("error getting worktree changes: %w", err)
	}

	printWorktreeChanges(changes)
	return nil
}

func DiffJSON(repoPath string) ([]byte, error) {
	changes, err := getWorktreeChanges(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error getting worktree changes: %w", err)
	}

	return json.MarshalIndent(changes, "", "  ")
}
