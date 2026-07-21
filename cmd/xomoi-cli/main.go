// Xomoi-Core: Sovereign Edge Node
// Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
