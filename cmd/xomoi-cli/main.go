package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	serverAddr := flag.String("server", "localhost:8080", "Address of the Xomoi-Core API")
	flag.Parse()

	// Initialize the Bubble Tea program with the Alternate Screen buffer.
	// This ensures when the user quits, their terminal history is perfectly preserved.
	p := tea.NewProgram(InitialModel(*serverAddr), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal TUI Error: %v", err)
		os.Exit(1)
	}
}
