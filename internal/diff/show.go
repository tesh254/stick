// git show <commit_hash>
package diff

import (
	"encoding/json"
	"fmt"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/format/diff"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type CommitOutput struct {
	Commit struct {
		Hash    string `json:"hash"`
		Author  string `json:"author"`
		Email   string `json:"email"`
		Date    string `json:"date"`
		Message string `json:"message"`
	} `json:"commit"`
	Files []FileChange `json:"files"`
}

type FileChange struct {
	Path     string       `json:"path"`
	Status   string       `json:"status"`
	Changes  []LineChange `json:"changes"`
	isBinary bool         `json:"is_binary"`
}

type LineChange struct {
	Line    int    `json:"line"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func ShowCommitDiff(repoPath string, commitHash string) ([]byte, error) {
	repo, err := git.PlainOpen(repoPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commitHash))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve commit hash: %w", err)
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	output := CommitOutput{}
	output.Commit.Hash = commit.Hash.String()
	output.Commit.Author = commit.Author.Name
	output.Commit.Email = commit.Author.Email
	output.Commit.Date = commit.Author.When.Format(time.RFC3339)
	output.Commit.Message = commit.Message

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit tree: %w", err)
	}

	var parentTree *object.Tree
	if commit.NumParents() > 0 {
		parent, err := commit.Parents().Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get parent commit: %w", err)
		}
		parentTree, err = parent.Tree()
		if err != nil {
			return nil, fmt.Errorf("failed to get parent tree: %w", err)
		}
	}

	var changes object.Changes
	if parentTree != nil {
		changes, err = object.DiffTree(parentTree, commitTree)
		if err != nil {
			return nil, fmt.Errorf("failed to compute diff: %w", err)
		}
	} else {
		changes, err = object.DiffTree(&object.Tree{}, commitTree)
		if err != nil {
			return nil, fmt.Errorf("failed to compute diff for initial commit: %w", err)
		}
	}

	for _, change := range changes {
		fileChange := FileChange{
			Path:   change.To.Name,
			Status: getChangeStatus(change),
		}

		patch, err := change.Patch()
		if err != nil {
			return nil, fmt.Errorf("failed to get patch for file %s: %w", change.To.Name, err)
		}

		for _, filePatch := range patch.FilePatches() {
			fileChange.isBinary = filePatch.IsBinary()
			if fileChange.isBinary {
				continue
			}
			for i, chunk := range filePatch.Chunks() {
				lineType := ""
				switch chunk.Type() {
				case diff.Add:
					lineType = "added"
				case diff.Delete:
					lineType = "deleted"
				default:
					continue // Skip unchanged lines
				}
				fileChange.Changes = append(fileChange.Changes, LineChange{
					Line:    i + 1, // Approximate line number
					Type:    lineType,
					Content: chunk.Content(),
				})
			}
		}
		output.Files = append(output.Files, fileChange)
	}

	return json.MarshalIndent(output, "", "  ")
}

func getChangeStatus(change *object.Change) string {
	action, _ := change.Action()
	switch action.String() {
	case "Addition":
		return "added"
	case "Deletion":
		return "deleted"
	case "Modification":
		return "modified"
	default:
		return "unknown"
	}
}
