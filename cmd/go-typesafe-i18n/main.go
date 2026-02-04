package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
	"github.com/mickamy/go-typesafe-i18n/internal/lint"
)

var (
	version             = "dev"
	supportedExtensions = map[string]bool{
		".yaml": true,
		".yml":  true,
		".json": true,
		".toml": true,
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "go-typesafe-i18n: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("no command specified")
	}

	switch os.Args[1] {
	case "generate":
		return runGenerate(os.Args[2:])
	case "lint":
		return runLint(os.Args[2:])
	case "version", "-version", "--version":
		fmt.Printf("go-typesafe-i18n version %s\n", version)
		return nil
	case "help", "-h", "-help", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: go-typesafe-i18n <command> [options]

Commands:
  generate    Generate type-safe message functions from a locale file
  lint        Check key consistency across locale files

Run 'go-typesafe-i18n <command> -help' for more information on a command.
`)
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	pkg := fs.String("pkg", "messages", "package name for generated code")
	out := fs.String("out", "messages_gen.go", "output file path")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go-typesafe-i18n generate [options] <locale-file>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return fmt.Errorf("expected exactly one input file")
	}

	cfg := codegen.Config{
		InputPath:   fs.Arg(0),
		OutputPath:  *out,
		PackageName: *pkg,
	}

	if err := codegen.Generate(cfg); err != nil {
		return err
	}

	fmt.Printf("Generated %s\n", *out)
	return nil
}

func runLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	base := fs.String("base", "", "base locale name or file path (required)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: go-typesafe-i18n lint -base=<locale> <directory>
       go-typesafe-i18n lint -base=<file> <target-files...>

Examples:
  go-typesafe-i18n lint -base=en locales/
  go-typesafe-i18n lint -base=locales/en.yaml locales/ja.yaml

Options:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *base == "" {
		fs.Usage()
		return fmt.Errorf("-base flag is required")
	}

	if fs.NArg() == 0 {
		fs.Usage()
		return fmt.Errorf("at least one target path is required")
	}

	basePath, targetPaths, err := resolveLocalePaths(*base, fs.Args())
	if err != nil {
		return err
	}

	results, err := lint.Check(basePath, targetPaths)
	if err != nil {
		return err
	}

	hasIssues := false
	for _, r := range results {
		if !r.HasIssues() {
			continue
		}
		hasIssues = true
		fmt.Printf("%s:\n", r.Path)
		if len(r.Missing) > 0 {
			fmt.Printf("  missing: %s\n", strings.Join(r.Missing, ", "))
		}
		if len(r.Extra) > 0 {
			fmt.Printf("  extra: %s\n", strings.Join(r.Extra, ", "))
		}
	}

	if hasIssues {
		return fmt.Errorf("key consistency check failed")
	}

	fmt.Println("All locale files are consistent.")
	return nil
}

func resolveLocalePaths(base string, args []string) (basePath string, targetPaths []string, err error) {
	if len(args) == 1 {
		info, statErr := os.Stat(args[0])
		if statErr == nil && info.IsDir() {
			return scanLocaleDir(base, args[0])
		}
	}
	return base, args, nil
}

func scanLocaleDir(base, dir string) (basePath string, targetPaths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !supportedExtensions[ext] {
			continue
		}

		path := filepath.Join(dir, name)
		nameWithoutExt := strings.TrimSuffix(name, ext)

		if nameWithoutExt == base {
			basePath = path
		} else {
			targetPaths = append(targetPaths, path)
		}
	}

	if basePath == "" {
		return "", nil, fmt.Errorf("base locale %q not found in %s", base, dir)
	}

	return basePath, targetPaths, nil
}
