package changes

import (
	"fmt"

	"github.com/fatih/color"
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

func printChange(change ChangeRecord) {
	var statusColor *color.Color
	statusText := ""

	switch change.Type {
	case git.Untracked:
		statusColor = color.New(color.FgGreen)
		statusText = "U"
	case git.Added:
		statusColor = color.New(color.FgGreen)
		statusText = "A"
	case git.Modified:
		statusColor = color.New(color.FgYellow)
		statusText = "M"
	case git.Deleted:
		statusColor = color.New(color.FgRed)
		statusText = "D"
	case git.Renamed:
		statusColor = color.New(color.FgMagenta)
		statusText = "R"
	case git.Copied:
		statusColor = color.New(color.FgCyan)
		statusText = "C"
	case git.UpdatedButUnmerged:
		statusColor = color.New(color.FgRed)
		statusText = "??"
	default:
		statusColor = color.New(color.FgWhite)
		statusText = string(change.Type)
	}

	if change.Type == git.Renamed {
		// Display rename as "oldname -> newname"
		statusColor.Printf("  %s %s -> %s\n", statusText, change.Extra, change.File)
	} else {
		// Use original formatting for non-renamed files
		statusColor.Printf("  %s %s\n", statusText, change.File)
	}
}

func retrieveRepoChangeStatus(repoPath string) ([]ChangeRecord, []ChangeRecord, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get repo changes status: %w", err)
	}

	var stagedChanges []ChangeRecord
	var worktreeChanges []ChangeRecord

	for file, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			extra := ""
			if fileStatus.Staging == git.Renamed {
				extra = fileStatus.Extra
			}
			stagedChanges = append(stagedChanges, ChangeRecord{
				Area:  "Staging",
				Type:  fileStatus.Staging,
				File:  file,
				Extra: extra,
			})
		}
		if fileStatus.Worktree != git.Unmodified {
			extra := ""
			if fileStatus.Worktree == git.Renamed {
				extra = fileStatus.Extra
			}
			worktreeChanges = append(worktreeChanges, ChangeRecord{
				Area:  "Worktree",
				Type:  fileStatus.Worktree,
				File:  file,
				Extra: extra,
			})
		}
	}

	return stagedChanges, worktreeChanges, nil
}
