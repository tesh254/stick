package release

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"golang.org/x/term"
)

type ReleaseOutput struct {
	Tag struct {
		Name      string `json:"name"`
		Commit    string `json:"commit"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"tag"`
	PushStatus string `json:"push_status"`
}

func CreateAndPushRelease(ctx context.Context, repoPath, tagName, commitHash, message string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commitHash))
	if err != nil {
		return fmt.Errorf("failed to resolve commit hash: %w", err)
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return fmt.Errorf("failed to get commit %s:  %w", commitHash, err)
	}

	tableData := [][]string{
		{
			tagName,
			commit.Hash.String(),
			message,
			"pending",
		},
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// create tag
	if message == "" {
		_, err = repo.CreateTag(tagName, *hash, nil)
	} else {
		_, err = repo.CreateTag(tagName, *hash, &git.CreateTagOptions{
			Tagger: &object.Signature{
				Name:  commit.Author.Name,
				Email: commit.Author.Email,
				When:  time.Now(),
			},
			Message: message,
		})
	}
	if err != nil {
		tableData[0][3] = red("failed")
		printTable(tableData)
		return fmt.Errorf("failed to create tag: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		tableData[0][3] = red("failed")
		printTable(tableData)
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	// Using public key for authentication
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	sshPath := filepath.Join(home, ".ssh")
	files, err := os.ReadDir(sshPath)
	if err != nil {
		tableData[0][3] = red("failed")
		printTable(tableData)
		return fmt.Errorf("failed to read ssh keys from %s: %w", sshPath, err)
	}

	var lastPushErr error
	pushSuccess := false

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
			pushSuccess = true
			break
		}

		lastPushErr = pushErr
	}

	if !pushSuccess {
		tableData[0][3] = red("failed")
		printTable(tableData)
		if lastPushErr != nil {
			return fmt.Errorf("failed to push tag, last error: %w", lastPushErr)
		}
		return fmt.Errorf("failed to push tag: no valid ssh key found")
	}

	tableData[0][3] = green("success")

	printTable(tableData)

	return nil
}

func printTable(data [][]string) {
	if len(data) == 0 || len(data[0]) < 4 {
		return
	}
	row := data[0]
	tag := row[0]
	commit := row[1]
	message := row[2]
	status := row[3]

	bold := color.New(color.Bold).SprintFunc()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "|\t\t|")
	fmt.Fprintln(w, "|----------\t|------------------------------------------|")
	fmt.Fprintf(w, "| %s\t| %s\t|\n", bold("commit"), commit)
	fmt.Fprintf(w, "| %s\t| %s\t|\n", bold("tag"), tag)
	fmt.Fprintf(w, "| %s\t| %s\t|\n", bold("message"), message)
	fmt.Fprintf(w, "| %s\t| %s\t|\n", bold("status"), status)
}
