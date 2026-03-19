package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/esteban-herrera/rad.io/internal/player"
	"github.com/esteban-herrera/rad.io/internal/store"
	"github.com/esteban-herrera/rad.io/internal/ui"
)

func main() {
	stations, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load stations: %v\n", err)
		os.Exit(1)
	}

	p := player.New()
	m := ui.New(stations, p)

	prog := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
