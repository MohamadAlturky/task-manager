package main

import (
	"fmt"
	"os"

	"devtui/src/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model, err := ui.NewModel()
	if err != nil {
		fmt.Printf("Error initializing model: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
