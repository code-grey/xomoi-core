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
	"log"
	"os"

	"github.com/code-grey/xomoi-core/internal/transpiler"
)

func main() {
	protoDir := flag.String("proto-dir", "proto/v1", "Path to protobuf directory")
	target := flag.String("target", "cpp", "Target language (cpp)")
	profile := flag.String("profile", "basic_sensor", "Adaptive profile name (e.g., dht22, bme280)")
	
	flag.Parse()

	log.Println("Xomoi Transpiler [xomoi-ctl] Starting...")

	ast, err := transpiler.ParseDirectory(*protoDir)
	if err != nil {
		log.Fatalf("Failed to parse protos: %v", err)
	}

	if *target == "cpp" {
		out, err := transpiler.GenerateLiteSDK(ast, transpiler.AdaptiveProfile{DeviceType: *profile})
		if err != nil {
			log.Fatalf("Failed to generate SDK: %v", err)
		}
		
		// Output the transpiled SDK to stdout (or disk in a real run)
		os.Stdout.Write(out)
		log.Printf("\nSuccessfully generated SDK for %s", *profile)
	}
}
