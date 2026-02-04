package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "go-typesafe-i18n: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		pkg         = flag.String("pkg", "messages", "package name for generated code")
		out         = flag.String("out", "messages_gen.go", "output file path")
		showVersion = flag.Bool("version", false, "show version")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go-typesafe-i18n [options] <yaml-file>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("go-typesafe-i18n version %s\n", version)
		return nil
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return fmt.Errorf("expected exactly one input file")
	}

	cfg := codegen.Config{
		InputPath:   args[0],
		OutputPath:  *out,
		PackageName: *pkg,
	}

	if err := codegen.Generate(cfg); err != nil {
		return err
	}

	fmt.Printf("Generated %s\n", *out)
	return nil
}
