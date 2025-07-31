package vc

import (
	"context"

	"github.com/fatih/color"
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

// CreateAndPushRelease creates a new git tag and pushes it to the origin remote.
func CreateAndPushRelease(ctx context.Context, repoPath, tagName, commitHash, message string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	tableData := [][]string{
		{"tag", "commit", "message", "status"},
		{tagName, commitHash, message, "pending"},
	}

	repo, commit, err := openRepositoryAndGetCommit(repoPath, commitHash)
	if err != nil {
		tableData[1][3] = red("failed")
		printTable(tableData)
		return err
	}
	// Update commit hash to be the full hash
	tableData[1][1] = commit.Hash.String()

	if err := createTag(repo, tagName, message, commit); err != nil {
		tableData[1][3] = red("failed")
		printTable(tableData)
		return err
	}

	if err := pushReleaseTag(repo, tagName); err != nil {
		tableData[1][3] = red("failed")
		printTable(tableData)
		return err
	}

	tableData[1][3] = green("success")
	printTable(tableData)

	return nil
}
