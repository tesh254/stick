package vbranch

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tesh254/stick/internal/constants"
)

func Init() {
	EnsureStateInitialized()
	if !isGitRepo() {
		fmt.Println("not a git repository")
		os.Exit(1)
	}

	if err := os.MkdirAll(constants.STICK_DIR, 0755); err != nil {
		fmt.Printf("error creating stick directory: %v\n", err)
		os.Exit(1)
	}

	// create default virtual branch
	defaultBranch := &VirtualBranch{
		Name:      "main-changes",
		ID:        generateID(),
		Files:     make(map[string]string),
		Hunks:     []Hunk{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}

	state.Branches[defaultBranch.ID] = defaultBranch
	state.CurrentBranch = defaultBranch.ID

	saveState()
	fmt.Print(constants.ASCII)
	fmt.Println("stick initialized successfully!")
	fmt.Printf("created default virtual branch: %s\n", defaultBranch.Name)
}

func ListBranches() {
	fmt.Println("virtual branches: ")

	for _, branch := range state.Branches {
		status := " (active)"
		if branch.ID == state.CurrentBranch {
			status += " *"
		}

		fmt.Printf("  %s%s - %d hunks\n", branch.Name, status, len(branch.Hunks))
	}
}

func CreateBranch(name string) {
	branch := &VirtualBranch{
		Name:      name,
		ID:        generateID(),
		Files:     make(map[string]string),
		Hunks:     []Hunk{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}

	state.Branches[branch.ID] = branch
	saveState()
	fmt.Printf("created virtual branch: %s\n", name)
}

func SwitchBranch(name string, args []string) {
	for id, branch := range state.Branches {
		if branch.Name == name {
			state.CurrentBranch = id
			saveState()
			fmt.Printf("switched to virtual branch: %s\n", name)
			return
		}
	}
	fmt.Printf("branch '%s' not found\n", name)
}

func Status() {
	fmt.Println("stick Status:")
	fmt.Printf("git Root: %s\n", state.GitRoot)
	fmt.Printf("current branch: %s\n", getCurrentBranchName())
	fmt.Println()

	// Show Git status
	gitStatus := getGitStatus()
	if len(gitStatus) > 0 {
		fmt.Println("uncommitted changes:")
		for _, file := range gitStatus {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
	}

	// Show virtual branches
	fmt.Println("virtual branches:")
	for _, branch := range state.Branches {
		status := ""
		if branch.Active {
			status = " (active)"
		}
		if branch.ID == state.CurrentBranch {
			status += " *"
		}
		fmt.Printf("  %s%s:\n", branch.Name, status)
		fmt.Printf("    files: %d\n", len(branch.Files))
		fmt.Printf("    hunks: %d\n", len(branch.Hunks))
		fmt.Printf("    updated: %s\n", branch.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}

func AddFile(cmd *cobra.Command, args []string) {
	if state.CurrentBranch == "" {
		fmt.Println("no current virtual branch. Use 'stick branch create' first.")
		return
	}
	all, _ := cmd.Flags().GetBool("all")
	if all || len(args) == 0 || (len(args) == 1 && args[0] == ".") {
		AddAll()
	} else {
		branch := state.Branches[state.CurrentBranch]
		for _, file := range args {
			if err := addFileToVirtualBranch(branch, file); err != nil {
				fmt.Printf("error adding %s: %v\n", file, err)
			} else {
				fmt.Printf("added %s to virtual branch %s\n", file, branch.Name)
			}
		}
		branch.UpdatedAt = time.Now()
		saveState()
	}
}

func MoveHunkToTargetBranch(hunkID string, targetBranchName string) {
	var targetBranch *VirtualBranch
	for _, branch := range state.Branches {
		if branch.Name == targetBranchName {
			targetBranch = branch
			break
		}
	}

	if targetBranch == nil {
		fmt.Printf("target branch '%s' not found\n", targetBranchName)
		return
	}

	// Find and move the hunk
	for _, sourceBranch := range state.Branches {
		for i, hunk := range sourceBranch.Hunks {
			if hunk.ID == hunkID {
				// Remove from source
				sourceBranch.Hunks = append(sourceBranch.Hunks[:i], sourceBranch.Hunks[i+1:]...)
				// Add to target
				targetBranch.Hunks = append(targetBranch.Hunks, hunk)

				sourceBranch.UpdatedAt = time.Now()
				targetBranch.UpdatedAt = time.Now()
				saveState()

				fmt.Printf("moved hunk %s to branch %s\n", hunkID, targetBranchName)
				return
			}
		}
	}

	fmt.Printf("hunk '%s' not found\n", hunkID)
}

func PushBranchToRemoteAsGitBranch(targetBranchName *string) {
	branchName := ""
	if targetBranchName != nil {
		branchName = *targetBranchName
	} else {
		branchName = getCurrentBranchName()
	}

	var targetBranch *VirtualBranch
	for _, branch := range state.Branches {
		if branch.Name == branchName {
			targetBranch = branch
			break
		}
	}

	if targetBranch == nil {
		fmt.Printf("branch '%s' not found\n", branchName)
		return
	}

	if err := pushVirtualBranch(targetBranch); err != nil {
		fmt.Printf("error pushing branch: %v\n", err)
	} else {
		fmt.Printf("successfully pushed virtual branch '%s' to remote\n", branchName)
	}
}

func ApplyVBranchChangesToWorkingDir(targetBranchName *string) {
	branchName := ""
	if targetBranchName != nil {
		branchName = *targetBranchName
	} else {
		branchName = getCurrentBranchName()
	}

	var targetBranch *VirtualBranch
	for _, branch := range state.Branches {
		if branch.Name == branchName {
			targetBranch = branch
			break
		}
	}

	if targetBranch == nil {
		fmt.Printf("branch '%s' not found\n", branchName)
		return
	}

	if err := applyVirtualBranch(targetBranch); err != nil {
		fmt.Printf("error applying branch: %v\n", err)
	} else {
		fmt.Printf("applied virtual branch '%s' to working directory\n", branchName)
	}
}

func UnapplyVBranchChangesToWorkingDir(targetBranchName *string) {
	branchName := ""
	if targetBranchName != nil {
		branchName = *targetBranchName
	} else {
		branchName = getCurrentBranchName()
	}

	var targetBranch *VirtualBranch
	for _, branch := range state.Branches {
		if branch.Name == branchName {
			targetBranch = branch
			break
		}
	}

	if targetBranch == nil {
		fmt.Printf("branch '%s' not found\n", branchName)
		return
	}

	if err := unapplyVirtualBranch(targetBranch); err != nil {
		fmt.Printf("error unapplying branch: %v\n", err)
	} else {
		fmt.Printf("unapplied virtual branch '%s' from working directory\n", branchName)
	}
}

func SyncBranchesWithGitRepoState() {
	fmt.Println("syncing with Git repository...")

	// Update state from current Git status
	if err := syncWithGit(); err != nil {
		fmt.Printf("error syncing: %v\n", err)
		return
	}

	state.LastSync = time.Now()
	saveState()
	fmt.Println("sync completed successfully!")
}
