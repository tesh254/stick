// create.go
package branch

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/tesh254/stick/internal/metadata"
)

func CreateVirtualBranch(branchName *string, remoteName *string) {
	var virtualBranchName string = "portal"
	var baseBranch string

	// check if we're in a git repository
	if err := checkGitRepository(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// determine remote name
	var rm string = "origin"
	if remoteName != nil {
		rm = *remoteName
	}

	// check if remote exists
	if err := checkRemoteExists(rm); err != nil {
		var remoteMessage string = ""
		if remoteName == nil {
			remoteMessage = "\nplease provide a remote name if it's not origin"
		}
		fmt.Println("Error:", err, remoteMessage)
		return
	}

	// fetch from remote
	fmt.Printf("Fetching from remote '%s'...\n", rm)
	if err := fetchRemote(rm); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// get default branch
	defaultBranch, err := getDefaultBranch(rm)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// get all remote branches
	branchList, err := getRemoteBranches(rm)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(branchList) == 0 {
		fmt.Println("no remote branches found")
		return
	}

	// determine base branch
	if branchName != nil {
		baseBranch = *branchName
	} else {
		// prompt user to select branch
		prompt := promptui.Select{
			Label: fmt.Sprintf("select base branch for virtual branch (default: %s)", defaultBranch),
			Items: branchList,
		}
		_, selectedBranch, promptError := prompt.Run()
		if promptError != nil {
			fmt.Println("prompt cancelled or failed:", promptError)
			return
		}
		baseBranch = selectedBranch
		fmt.Printf("selected base branch: %s\n", baseBranch)
	}

	// validate chosen branch exists remotely
	if !remoteBranchExists(rm, baseBranch) {
		fmt.Printf("base branch '%s' not found in remote. available branches: %v\n", baseBranch, branchList)
		return
	}

	// load metadata
	md := metadata.LoadMetadata()
	if _, exists := md.VirtualBranches[virtualBranchName]; exists {
		fmt.Println("virtual branch already exists.")
		return
	}

	// create a new Git branch (e.g., "stick/<virtual_branch_name>")
	gitBranchName := "stick/" + virtualBranchName

	// check if the branch already exists
	if branchExists(gitBranchName) {
		fmt.Printf("Git branch '%s' already exists\n", gitBranchName)
		return
	}

	// create the local branch from the remote branch
	if err := createLocalBranch(gitBranchName, rm, baseBranch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// update metadata
	md.VirtualBranches[virtualBranchName] = metadata.VirtualBranch{
		GitBranch: gitBranchName,
		Files:     []string{},
	}
	metadata.SaveMetadata(md)
	fmt.Printf("virtual branch '%s' created successfully (git branch: %s)\n", virtualBranchName, gitBranchName)
}
