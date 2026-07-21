package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const agplText = `Xomoi-Core: Sovereign Edge Node
Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

func getGoHeader() []byte {
	lines := strings.Split(agplText, "\n")
	var b bytes.Buffer
	for _, l := range lines {
		if l == "" {
			b.WriteString("//\n")
		} else {
			b.WriteString("// " + l + "\n")
		}
	}
	b.WriteString("\n")
	return b.Bytes()
}

func getSvelteHeader() []byte {
	return []byte("<!--\n" + agplText + "\n-->\n\n")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run add_agpl.go <directory_or_file>")
		os.Exit(1)
	}
	target := os.Args[1]

	goHeader := getGoHeader()
	svelteHeader := getSvelteHeader()
	searchToken := []byte("GNU Affero General Public License")

	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip node_modules and vendor directories
			if d.Name() == "node_modules" || d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".svelte" {
			return nil
		}

		// Read the entire file
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", path, err)
			return nil
		}

		// Check if it already has the license
		if bytes.Contains(content, searchToken) {
			// Skip, already licensed
			return nil
		}

		// Inject Header
		var newContent []byte
		if ext == ".go" {
			newContent = append(goHeader, content...)
		} else if ext == ".svelte" {
			newContent = append(svelteHeader, content...)
		}

		// Atomic write back
		err = os.WriteFile(path, newContent, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", path, err)
		} else {
			fmt.Printf("Injected AGPLv3 into %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Walk failed: %v\n", err)
	}
}
