package vc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommitDiff(t *testing.T) {
	repo, path := setupRepo(t)
	w, err := repo.Worktree()
	require.NoError(t, err)

	// Initial commit
	filePath := filepath.Join(path, "test.txt")
	err = os.WriteFile(filePath, []byte("hello world"), 0644)
	require.NoError(t, err)
	_, err = w.Add("test.txt")
	require.NoError(t, err)
	commit1Hash, err := w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	commit1, err := repo.CommitObject(commit1Hash)
	require.NoError(t, err)

	// Second commit with changes
	// Overwrite the file to create a change
	err = os.WriteFile(filePath, []byte(`hello world
hello again`), 0644)
	require.NoError(t, err)

	_, err = w.Add("test.txt")
	require.NoError(t, err)

	commit2Hash, err := w.Commit("second commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	commit2, err := repo.CommitObject(commit2Hash)
	require.NoError(t, err)

	t.Run("diff for initial commit", func(t *testing.T) {
		jsonData, err := ShowCommitDiff(path, commit1.Hash.String())
		require.NoError(t, err)

		var output CommitOutput
		err = json.Unmarshal(jsonData, &output)
		require.NoError(t, err)

		assert.Equal(t, commit1.Hash.String(), output.Commit.Hash)
		assert.Len(t, output.Files, 1)
		assert.Equal(t, "added", output.Files[0].Status)
		assert.Equal(t, "test.txt", output.Files[0].Path)
	})

	t.Run("diff for second commit", func(t *testing.T) {
		jsonData, err := ShowCommitDiff(path, commit2.Hash.String())
		require.NoError(t, err)

		var output CommitOutput
		err = json.Unmarshal(jsonData, &output)
		require.NoError(t, err)

		assert.Equal(t, commit2.Hash.String(), output.Commit.Hash)
		assert.Len(t, output.Files, 1)
		assert.Equal(t, "modified", output.Files[0].Status)
		assert.Equal(t, "test.txt", output.Files[0].Path)
	})

	t.Run("invalid commit hash", func(t *testing.T) {
		_, err := ShowCommitDiff(path, "invalidhash")
		assert.Error(t, err)
	})
}
