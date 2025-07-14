package vbranch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tesh254/stick/internal/constants"
)

var state *StickState

func InitializeState() {
	if state == nil {
		state = &StickState{
			Branches:   make(map[string]*VirtualBranch),
			WorkingDir: getCurrentDir(),
			GitRoot:    getGitRoot(),
		}
	}

	// Try to load existing state
	if _, err := os.Stat(getStateFilePath()); err == nil {
		loadState()
	}
}

// EnsureStateInitialized ensures state is initialized before use
func EnsureStateInitialized() {
	if state == nil {
		InitializeState()
	}
}

// GetState returns the current state (initializing if necessary)
func GetState() *StickState {
	EnsureStateInitialized()
	return state
}

func getStateFilePath() string {
	return filepath.Join(constants.STICK_DIR, "state.json")
}

func loadState() error {
	stateFile := getStateFilePath()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// Ensure maps are initialized
	if state.Branches == nil {
		state.Branches = make(map[string]*VirtualBranch)
	}

	return nil
}

func saveState() error {
	EnsureStateInitialized()

	stateFile := getStateFilePath()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

func AddAll() {
	branch := state.Branches[state.CurrentBranch]
	gitStatus := getGitStatus()
	for _, statusLine := range gitStatus {
		if len(statusLine) > 3 {
			status := statusLine[:2]
			filename := strings.TrimSpace(statusLine[3:])
			switch status[1] {
			case 'M', 'A', '?': // Modified, added, or new
				content, err := os.ReadFile(filename)
				if err == nil {
					branch.Files[filename] = string(content)
					hunkType := "modify"
					if status[0] == '?' || status[0] == 'A' {
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
				}
			case 'D': // Deleted
				branch.DeletedFiles = append(branch.DeletedFiles, filename)
				hunk := Hunk{
					ID:        generateID(),
					File:      filename,
					Type:      "remove",
					CreatedAt: time.Now(),
				}
				branch.Hunks = append(branch.Hunks, hunk)
			}
		}
	}
	branch.UpdatedAt = time.Now()
	saveState()
	fmt.Println("added all changes to virtual branch", branch.Name)
}
