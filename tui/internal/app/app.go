package app

import (
	"fmt"
	"os"

	"github.com/MohamedElashri/snipo/tui/internal/config"
	"github.com/MohamedElashri/snipo/tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsConfigured() {
		fmt.Fprintf(os.Stderr, "Error: snippy is not configured. Please run 'snippy config' first\n")
		os.Exit(1)
	}

	m := ui.NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
