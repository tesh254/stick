package vc

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintRepoStatus(t *testing.T) {
	repo, path := setupRepo(t)
	createCommit(t, repo, path, "initial commit") // To have a HEAD

	t.Run("no changes", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintRepoStatus(path)
		require.NoError(t, err)

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "no changes")
	})

	t.Run("with changes", func(t *testing.T) {
		// Disable color for predictable output in tests
		color.NoColor = true
		defer func() { color.NoColor = false }()

		// Capture stdout
		oldOut := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w // Also redirect color output
		defer func() {
			os.Stdout = oldOut
			color.Output = oldOut
		}()

		// Create a new file
		newFilePath := filepath.Join(path, "new.txt")
		err := os.WriteFile(newFilePath, []byte("untracked file"), 0644)
		require.NoError(t, err)

		// Modify an existing file that is part of the commit
		existingFilePath := filepath.Join(path, "test.txt")
		err = os.WriteFile(existingFilePath, []byte("modified content"), 0644)
		require.NoError(t, err)

		// Stage the new file
		wt, err := repo.Worktree()
		require.NoError(t, err)
		_, err = wt.Add("new.txt")
		require.NoError(t, err)

		err = PrintRepoStatus(path)
		require.NoError(t, err)
		w.Close()

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Changes to be committed:")
		assert.Contains(t, output, "  A new.txt")
		assert.Contains(t, output, "Changes not staged for commit:")
		assert.Contains(t, output, "  M test.txt")
	})
}
