package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tesh254/ffs/core"
	"github.com/tesh254/stick/internal/config"
)

func Run(provider string) {
	// Validate provider selection via config
	if _, _, err := config.GetProviderConfig(provider); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dirTree, err := core.WorkingDirectoryTree(nil, []string{".git"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m := NewModel(provider, cwd, &dirTree)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
