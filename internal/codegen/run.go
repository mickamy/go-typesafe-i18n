package codegen

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/language"
)

// Run executes the go-typesafe-i18n CLI: it generates the typed message
// package from a locale directory, or only validates the locales with
// -check. Warnings are written to stderr; they do not fail the run.
func Run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("go-typesafe-i18n", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		dir         = fs.String("dir", "locales", "directory containing locale files")
		defaultLang = fs.String("default", "en", "default language defining the generated signatures")
		out         = fs.String("out", "messages/messages.gen.go", "output file path")
		pkg         = fs.String("pkg", "", "package name of the generated file (default: base name of the output directory)")
		check       = fs.Bool("check", false, "validate the locales without writing the output file")
	)
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	tag, err := language.Parse(*defaultLang)
	if err != nil {
		return fmt.Errorf("invalid default language %q: %w", *defaultLang, err)
	}

	model, warnings, err := Analyze(*dir, tag)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintln(stderr, "warning:", w)
	}
	name := *pkg
	if name == "" {
		absOut, err := filepath.Abs(*out)
		if err != nil {
			return fmt.Errorf("resolve output path: %w", err)
		}
		name = filepath.Base(filepath.Dir(absOut))
	}
	src, err := Render(model, name)
	if err != nil {
		return err
	}
	if *check {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(*out, src, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Fprintf(stdout, "generated %s (%d messages)\n", *out, len(model.Messages))
	return nil
}
