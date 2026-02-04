package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "go-typesafe-i18n: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	var (
		showVersion = flag.Bool("version", false, "Show version")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("go-typesafe-i18n version %s\n", version)
		return nil
	}

	fmt.Println("go-typesafe-i18n: No operation specified.")

	return nil
}
