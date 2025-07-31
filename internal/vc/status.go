package vc

import (
	"fmt"

	git "github.com/go-git/go-git/v6"
)

type ChangeRecord struct {
	Area  string
	Type  git.StatusCode
	File  string
	Extra string
}

func PrintRepoStatus(repoPath string) error {
	staged, worktree, err := retrieveRepoChangeStatus(repoPath)
	if err != nil {
		return err
	}

	if len(staged) == 0 && len(worktree) == 0 {
		fmt.Println("no changes")
		return nil
	}

	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		for _, change := range staged {
			printChange(change)
		}
		fmt.Println()
	}

	if len(worktree) > 0 {
		fmt.Println("\nChanges not staged for commit:")
		for _, change := range worktree {
			printChange(change)
		}
		fmt.Println()
	}

	return nil
}
