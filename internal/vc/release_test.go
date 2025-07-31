package vc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRepo creates a new git repository in a temporary directory and returns the repo object and the path.
func setupRepo(t *testing.T) (*git.Repository, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "test-repo")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	return repo, dir
}

// createCommit creates a new commit in the repository and returns the commit object.
func createCommit(t *testing.T, repo *git.Repository, dir, msg string) *object.Commit {
	t.Helper()
	w, err := repo.Worktree()
	require.NoError(t, err)

	// Create a file to commit
	filename := filepath.Join(dir, "test.txt")
	err = os.WriteFile(filename, []byte("hello world"), 0644)
	require.NoError(t, err)

	_, err = w.Add("test.txt")
	require.NoError(t, err)

	commitHash, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	commit, err := repo.CommitObject(commitHash)
	require.NoError(t, err)

	return commit
}

func TestOpenRepositoryAndGetCommit(t *testing.T) {
	repo, path := setupRepo(t)
	commit := createCommit(t, repo, path, "initial commit")

	t.Run("valid commit hash", func(t *testing.T) {
		r, c, err := openRepositoryAndGetCommit(path, commit.Hash.String())
		require.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, c)
		assert.Equal(t, commit.Hash, c.Hash)
	})

	t.Run("invalid commit hash", func(t *testing.T) {
		_, _, err := openRepositoryAndGetCommit(path, "invalidhash")
		assert.Error(t, err)
	})

	t.Run("invalid repo path", func(t *testing.T) {
		_, _, err := openRepositoryAndGetCommit("/invalid/path", commit.Hash.String())
		assert.Error(t, err)
	})
}

func TestCreateTag(t *testing.T) {
	repo, path := setupRepo(t)
	commit := createCommit(t, repo, path, "initial commit")

	t.Run("create tag without message (lightweight)", func(t *testing.T) {
		tagName := "v1.0.0"
		err := createTag(repo, tagName, "", commit)
		require.NoError(t, err)

		// For lightweight tags, we just check that the reference exists and points to the correct commit.
		ref, err := repo.Reference(plumbing.NewTagReferenceName(tagName), true)
		require.NoError(t, err)
		assert.Equal(t, commit.Hash, ref.Hash())
	})

	t.Run("create tag with message (annotated)", func(t *testing.T) {
		tagName := "v1.1.0"
		message := "release message"
		err := createTag(repo, tagName, message, commit)
		require.NoError(t, err)

		// To verify an annotated tag, we first get the reference...
		tagRef, err := repo.Reference(plumbing.NewTagReferenceName(tagName), true)
		require.NoError(t, err)

		// ...then get the tag object from the reference's hash.
		tagObject, err := repo.TagObject(tagRef.Hash())
		require.NoError(t, err)

		assert.Equal(t, tagName, tagObject.Name)
		assert.Equal(t, message, strings.TrimSpace(tagObject.Message))
		assert.Equal(t, commit.Hash, tagObject.Target)
	})

	t.Run("create existing tag", func(t *testing.T) {
		tagName := "v1.2.0"
		err := createTag(repo, tagName, "", commit)
		require.NoError(t, err)
		// Creating it again should fail
		err = createTag(repo, tagName, "", commit)
		assert.Error(t, err)
	})
}
