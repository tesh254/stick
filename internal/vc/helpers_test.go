package vc

import (
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommit(t *testing.T) {
	repo, path := setupRepo(t)
	commit := createCommit(t, repo, path, "initial commit")

	t.Run("valid commit", func(t *testing.T) {
		c, err := getCommit(path, commit.Hash.String())
		require.NoError(t, err)
		assert.Equal(t, commit.Hash, c.Hash)
	})

	t.Run("invalid commit hash", func(t *testing.T) {
		_, err := getCommit(path, "invalidhash")
		assert.Error(t, err)
	})

	t.Run("invalid repo path", func(t *testing.T) {
		_, err := getCommit("/invalid/path", commit.Hash.String())
		assert.Error(t, err)
	})
}

func TestGetChanges(t *testing.T) {
	repo, path := setupRepo(t)
	commit := createCommit(t, repo, path, "initial commit")

	changes, err := getChanges(commit)
	require.NoError(t, err)
	assert.Len(t, changes, 1) // test.txt was added
}

func TestBuildFileChanges(t *testing.T) {
	repo, path := setupRepo(t)
	commit := createCommit(t, repo, path, "initial commit")
	changes, err := getChanges(commit)
	require.NoError(t, err)

	fileChanges, err := buildFileChanges(changes)
	require.NoError(t, err)
	assert.Len(t, fileChanges, 1)
	assert.Equal(t, "test.txt", fileChanges[0].Path)
	assert.Equal(t, "added", fileChanges[0].Status)
}

func TestRetrieveRepoChangeStatus(t *testing.T) {
	repo, path := setupRepo(t)
	createCommit(t, repo, path, "initial commit")

	// No changes initially in worktree after commit
	staged, worktree, err := retrieveRepoChangeStatus(path)
	require.NoError(t, err)
	assert.Empty(t, staged)
	assert.Empty(t, worktree)

	// Add a file to worktree
	filePath := filepath.Join(path, "newfile.txt")
	err = os.WriteFile(filePath, []byte("new file"), 0644)
	require.NoError(t, err)

	staged, worktree, err = retrieveRepoChangeStatus(path)
	require.NoError(t, err)
	assert.Empty(t, staged)
	assert.Len(t, worktree, 1)
	assert.Equal(t, "newfile.txt", worktree[0].File)
	assert.Equal(t, git.Untracked, worktree[0].Type)

	// Stage the file
	w, err := repo.Worktree()
	require.NoError(t, err)
	_, err = w.Add("newfile.txt")
	require.NoError(t, err)

	staged, worktree, err = retrieveRepoChangeStatus(path)
	require.NoError(t, err)
	assert.Len(t, staged, 1)
	assert.Empty(t, worktree)
	assert.Equal(t, "newfile.txt", staged[0].File)
	assert.Equal(t, git.Added, staged[0].Type)
}