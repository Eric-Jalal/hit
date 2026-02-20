package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elisa-content-delivery/hit/internal/app"
	"github.com/elisa-content-delivery/hit/internal/git"
)

func main() {
	repo, err := git.OpenCwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	owner, repoName, err := repo.OwnerRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s (CI features disabled)\n", err)
		owner = ""
		repoName = ""
	}

	model := app.NewModel(repo, owner, repoName)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
