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
