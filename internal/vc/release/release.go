package release

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
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

	// Using SSH agent for authentication
	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		tableData[0][3] = red("failed")
		printTable(tableData)
		return fmt.Errorf("failed to get auth for origin: %w", err)
	}

	// Push the tag
	err = remote.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagName, tagName)),
		},
		Auth: auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		tableData[0][3] = red("failed")
		printTable(tableData)
		return fmt.Errorf("failed to push tag: %w", err)
	}

	// Update status to success
	tableData[0][3] = green("success")

	// Print the table
	printTable(tableData)

	return nil
}

func printTable(data [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	header := []string{"Tag", "Commit", "Message", "Status"}
	fmt.Fprintln(w, strings.Join(header, "	"))
	for _, row := range data {
		fmt.Fprintln(w, strings.Join(row, "	"))
	}
}
