package vc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/format/diff"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/go-git/go-git/v6/utils/merkletrie"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

func getCommit(repoPath, commitHash string) (*object.Commit, error) {
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

	return commit, nil
}

func getChanges(commit *object.Commit) (object.Changes, error) {
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

	return changes, nil
}

func buildFileChanges(changes object.Changes) ([]FileChange, error) {
	var fileChanges []FileChange
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
			if filePatch.IsBinary() {
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
		fileChanges = append(fileChanges, fileChange)
	}
	return fileChanges, nil
}

// getChangeStatus determines the status of a change (added, deleted, modified).
func getChangeStatus(change *object.Change) string {
	action, err := change.Action()
	if err != nil {
		return "unknown"
	}
	switch action {
	case merkletrie.Insert:
		return "added"
	case merkletrie.Delete:
		return "deleted"
	case merkletrie.Modify:
		return "modified"
	default:
		return "unknown"
	}
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
		if fileStatus.Staging == git.Untracked || fileStatus.Worktree == git.Untracked {
			worktreeChanges = append(worktreeChanges, ChangeRecord{
				Area: "Worktree",
				Type: git.Untracked,
				File: file,
			})
			continue
		}

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

func openRepositoryAndGetCommit(repoPath, commitHash string) (*git.Repository, *object.Commit, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open repo: %w", err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commitHash))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve commit hash: %w", err)
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit %s: %w", commitHash, err)
	}

	return repo, commit, nil
}

func createTag(repo *git.Repository, tagName, message string, commit *object.Commit) error {
	var err error
	if message == "" {
		_, err = repo.CreateTag(tagName, commit.Hash, nil)
	} else {
		_, err = repo.CreateTag(tagName, commit.Hash, &git.CreateTagOptions{
			Tagger: &object.Signature{
				Name:  commit.Author.Name,
				Email: commit.Author.Email,
				When:  time.Now(),
			},
			Message: message,
		})
	}
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// pushReleaseTag attempts to push the specified tag to the "origin" remote.
// It iterates through available SSH keys, prompting for a password if a key is encrypted.
func pushReleaseTag(repo *git.Repository, tagName string) error {
	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	sshPath := filepath.Join(home, ".ssh")
	files, err := os.ReadDir(sshPath)
	if err != nil {
		return fmt.Errorf("failed to read ssh keys from %s: %w", sshPath, err)
	}

	var lastPushErr error
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		keyPath := filepath.Join(sshPath, file.Name())

		auth, authErr := ssh.NewPublicKeysFromFile("git", keyPath, "")
		if authErr != nil {
			if strings.Contains(authErr.Error(), "ssh: a password is required") {
				fmt.Printf("Enter password for key %s: ", file.Name())
				password, passErr := term.ReadPassword(int(syscall.Stdin))
				if passErr != nil {
					lastPushErr = fmt.Errorf("failed to read password: %w", passErr)
					continue
				}
				fmt.Println()
				auth, authErr = ssh.NewPublicKeysFromFile("git", keyPath, string(password))
				if authErr != nil {
					lastPushErr = fmt.Errorf("failed to create auth with password for %s: %w", keyPath, authErr)
					continue
				}
			} else {
				// Not a password error, just a key that can't be parsed, so skip it.
				lastPushErr = authErr
				continue
			}
		}

		// Attempt to push with the current key
		pushErr := remote.Push(&git.PushOptions{
			RefSpecs: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagName, tagName)),
			},
			Auth: auth,
		})

		if pushErr == nil || pushErr == git.NoErrAlreadyUpToDate {
			// Push succeeded
			return nil
		}

		lastPushErr = pushErr
	}

	// If we get here, all keys failed.
	if lastPushErr != nil {
		return fmt.Errorf("failed to push tag, last error: %w", lastPushErr)
	}
	return fmt.Errorf("failed to push tag: no valid ssh key found that could authenticate")
}

func printTable(data [][]string) {
	if len(data) == 0 || len(data[0]) < 4 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()
}
