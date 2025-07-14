// helpers.go
package branch

import (
	"fmt"
	"os/exec"
	"strings"
)

// executeGitCommand runs a git command and returns the output
func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git command failed: %s", string(exitError.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkGitRepository verifies we're in a git repository
func checkGitRepository() error {
	_, err := executeGitCommand("rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not in a git repository: %v", err)
	}
	return nil
}

// checkRemoteExists verifies the remote exists
func checkRemoteExists(remoteName string) error {
	_, err := executeGitCommand("remote", "get-url", remoteName)
	if err != nil {
		return fmt.Errorf("remote '%s' not found: %v", remoteName, err)
	}
	return nil
}

// fetchRemote fetches from the specified remote
func fetchRemote(remoteName string) error {
	_, err := executeGitCommand("fetch", remoteName)
	if err != nil {
		return fmt.Errorf("failed to fetch from remote '%s': %v", remoteName, err)
	}
	return nil
}

// getDefaultBranch gets the default branch for the remote
func getDefaultBranch(remoteName string) (string, error) {
	// Try to get the default branch from remote HEAD
	output, err := executeGitCommand("symbolic-ref", fmt.Sprintf("refs/remotes/%s/HEAD", remoteName))
	if err != nil {
		// If that fails, try to set it first
		_, setErr := executeGitCommand("remote", "set-head", remoteName, "--auto")
		if setErr != nil {
			return "", fmt.Errorf("failed to determine default branch: %v", err)
		}
		// Try again
		output, err = executeGitCommand("symbolic-ref", fmt.Sprintf("refs/remotes/%s/HEAD", remoteName))
		if err != nil {
			return "", fmt.Errorf("failed to get default branch: %v", err)
		}
	}

	// Extract branch name from refs/remotes/origin/main
	parts := strings.Split(output, "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("unexpected format for default branch: %s", output)
	}
	return parts[len(parts)-1], nil
}

// getRemoteBranches gets all remote branches for the specified remote
func getRemoteBranches(remoteName string) ([]string, error) {
	output, err := executeGitCommand("branch", "-r", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %v", err)
	}

	var branches []string
	prefix := remoteName + "/"

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip HEAD reference
		if strings.HasSuffix(line, "/HEAD") {
			continue
		}

		// Only include branches from the specified remote
		if strings.HasPrefix(line, prefix) {
			branchName := strings.TrimPrefix(line, prefix)
			branches = append(branches, branchName)
		}
	}

	return branches, nil
}

// branchExists checks if a branch exists (locally or remotely)
func branchExists(branchName string) bool {
	_, err := executeGitCommand("rev-parse", "--verify", branchName)
	return err == nil
}

// remoteBranchExists checks if a remote branch exists
func remoteBranchExists(remoteName, branchName string) bool {
	refName := fmt.Sprintf("refs/remotes/%s/%s", remoteName, branchName)
	return branchExists(refName)
}

// createLocalBranch creates a local branch from a remote branch
func createLocalBranch(localBranchName, remoteName, remoteBranchName string) error {
	remoteRef := fmt.Sprintf("%s/%s", remoteName, remoteBranchName)
	_, err := executeGitCommand("branch", localBranchName, remoteRef)
	if err != nil {
		return fmt.Errorf("failed to create local branch '%s' from '%s': %v", localBranchName, remoteRef, err)
	}
	return nil
}

// getCurrentBranch gets the current branch name
func getCurrentBranch() (string, error) {
	return executeGitCommand("branch", "--show-current")
}

// checkoutBranch checks out a specific branch
func checkoutBranch(branchName string) error {
	_, err := executeGitCommand("checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to checkout branch '%s': %v", branchName, err)
	}
	return nil
}
