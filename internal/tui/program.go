package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(deps *Dependencies) error {
	if deps == nil {
		return fmt.Errorf("dependencies are required")
	}

	p := tea.NewProgram(newModel(deps), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}
