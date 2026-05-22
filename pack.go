package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dinakars777/moody/voice"
)

func runPackCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: moody pack validate <path>")
		os.Exit(2)
	}

	switch args[0] {
	case "validate":
		runPackValidate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown pack command: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Usage: moody pack validate <path>")
		os.Exit(2)
	}
}

func runPackValidate(args []string) {
	validateFlags := flag.NewFlagSet("pack validate", flag.ExitOnError)
	jsonOutput := validateFlags.Bool("json", false, "Print machine-readable JSON")
	if err := validateFlags.Parse(args); err != nil {
		os.Exit(2)
	}
	if validateFlags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: moody pack validate <path>")
		os.Exit(2)
	}

	manifest, packName, err := voice.ValidatePack(validateFlags.Arg(0))
	if err != nil {
		log.Fatalf("pack validation failed: %v", err)
	}

	result := struct {
		Valid       bool   `json:"valid"`
		Pack        string `json:"pack"`
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}{
		Valid:       true,
		Pack:        packName,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			log.Fatal(err)
		}
		return
	}

	fmt.Printf("✓ valid pack: %s (%s)\n", result.Pack, result.Name)
}
