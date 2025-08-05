// Package diff provides functionality to show differences between commits.
package vc

import (
	"encoding/json"
	"time"
)

// ShowCommitDiff generates a diff for a given commit hash in a repository.
// It returns the diff as a JSON byte array.
func ShowCommitDiff(repoPath string, commitHash string) ([]byte, error) {
	commit, err := getCommit(repoPath, commitHash)
	if err != nil {
		return nil, err
	}

	changes, err := getChanges(commit)
	if err != nil {
		return nil, err
	}

	fileChanges, err := buildFileChanges(changes)
	if err != nil {
		return nil, err
	}

	output := CommitOutput{}
	output.Commit.Hash = commit.Hash.String()
	output.Commit.Author = commit.Author.Name
	output.Commit.Email = commit.Author.Email
	output.Commit.Date = commit.Author.When.Format(time.RFC3339)
	output.Commit.Message = commit.Message
	output.Files = fileChanges

	return json.MarshalIndent(output, "", "  ")
}
