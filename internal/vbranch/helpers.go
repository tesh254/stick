package vbranch

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func getGitRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

func getGitStatus() []string {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 3 {
			files = append(files, line)
		}
	}
	return files
}

func getCurrentBranchName() string {
	if branch, exists := state.Branches[state.CurrentBranch]; exists {
		return branch.Name
	}
	return ""
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func pushVirtualBranch(branch *VirtualBranch) error {
	cmd := exec.Command("git", "checkout", "-b", branch.Name)
	if err := cmd.Run(); err != nil {
		return err
	}

	for filename, content := range branch.Files {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return err
		}
	}

	cmd = exec.Command("git", "add", "-A") // Stage all changes, including deletions
	if err := cmd.Run(); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("Virtual branch: %s", branch.Name)
	if branch.Description != "" {
		commitMsg = branch.Description
	}
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "push", "-u", "origin", branch.Name)
	return cmd.Run()
}

func applyVirtualBranch(branch *VirtualBranch) error {
	for filename, content := range branch.Files {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return err
		}
	}
	for _, filename := range branch.DeletedFiles {
		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func unapplyVirtualBranch(branch *VirtualBranch) error {
	files := append([]string{}, branch.DeletedFiles...)
	for file := range branch.Files {
		files = append(files, file)
	}
	for _, file := range files {
		cmd := exec.Command("git", "checkout", "HEAD", "--", file)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func syncWithGit() error {
	// Get current Git status and update virtual branches accordingly
	gitFiles := getGitStatus()

	if len(gitFiles) == 0 {
		return nil
	}

	// If there's a current branch, add uncommitted changes to it
	if state.CurrentBranch != "" {
		branch := state.Branches[state.CurrentBranch]
		for _, file := range gitFiles {
			if len(file) > 3 {
				filename := strings.TrimSpace(file[3:])
				if err := addFileToVirtualBranch(branch, filename); err != nil {
					fmt.Printf("Warning: Could not add %s: %v\n", filename, err)
				}
			}
		}
		branch.UpdatedAt = time.Now()
	}

	return nil
}

func getFileStatus(filename string) string {
	cmd := exec.Command("git", "status", "--porcelain", filename)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	status := strings.TrimSpace(string(output))
	if status != "" {
		return status[:2] // First two characters (e.g., " M", "??", " D")
	}
	return ""
}

func addFileToVirtualBranch(branch *VirtualBranch, filename string) error {
	status := getFileStatus(filename)
	if status == "" {
		return fmt.Errorf("file %s is not tracked or has no changes", filename)
	}

	// Handle deleted files
	if status[1] == 'D' { // Deleted in working tree
		branch.DeletedFiles = append(branch.DeletedFiles, filename)
		hunk := Hunk{
			ID:        generateID(),
			File:      filename,
			Type:      "remove",
			CreatedAt: time.Now(),
		}
		branch.Hunks = append(branch.Hunks, hunk)
		return nil
	}

	// Handle new or modified files
	if status[0] == '?' || status[1] == 'M' || status[0] == 'A' { // New (??), modified ( M), or added (A )
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		branch.Files[filename] = string(content)
		hunkType := "modify"
		if status[0] == '?' { // New file
			hunkType = "add"
		}
		hunk := Hunk{
			ID:        generateID(),
			File:      filename,
			StartLine: 1,
			EndLine:   len(strings.Split(string(content), "\n")),
			Content:   string(content),
			Type:      hunkType,
			CreatedAt: time.Now(),
		}
		branch.Hunks = append(branch.Hunks, hunk)
		return nil
	}

	return fmt.Errorf("file %s has no changes to add", filename)
}

func getUniqueBranchName(desiredName string) string {
	name := desiredName
	counter := 1
	for {
		found := false
		for _, branch := range state.Branches {
			if branch.Name == name {
				found = true
				break
			}
		}
		if !found {
			return name
		}
		name = fmt.Sprintf("%s-%d", desiredName, counter)
		counter++
	}
}
