package codegen_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

func TestRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	locales := filepath.Join(dir, "locales")
	if err := os.Mkdir(locales, 0o750); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(locales, "en.yaml"), "greeting: \"Hello!\"\nhello: \"Hello, {name}!\"\n")
	writeFile(t, filepath.Join(locales, "ja.yaml"), "greeting: \"こんにちは！\"\n")
	out := filepath.Join(dir, "messages", "messages.gen.go")

	var stdout, stderr bytes.Buffer
	err := codegen.Run([]string{"-dir", locales, "-default", "en", "-out", out}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	src, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"package messages", "func Greeting()", "func Hello(name string)"} {
		if !strings.Contains(string(src), want) {
			t.Errorf("generated file does not contain %q", want)
		}
	}
	if !strings.Contains(stdout.String(), "generated") {
		t.Errorf("stdout = %q, want a generated report", stdout.String())
	}
	if !strings.Contains(stderr.String(), "missing keys: hello") {
		t.Errorf("stderr = %q, want a missing-keys warning", stderr.String())
	}
}

func TestRun_check(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")
	out := filepath.Join(dir, "messages", "messages.gen.go")

	var stdout, stderr bytes.Buffer
	err := codegen.Run([]string{"-dir", dir, "-out", out, "-check"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("-check wrote the output file: %v", err)
	}
}

func TestRun_checkReportsErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")
	writeFile(t, filepath.Join(dir, "ja.yaml"), "extra: \"余分\"\n")

	var stdout, stderr bytes.Buffer
	if err := codegen.Run([]string{"-dir", dir, "-check"}, &stdout, &stderr); err == nil {
		t.Error("Run() returned nil error for an invalid catalog")
	}
}

func TestRun_pkgFlag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")
	out := filepath.Join(dir, "out", "gen.go")

	var stdout, stderr bytes.Buffer
	err := codegen.Run([]string{"-dir", dir, "-out", out, "-pkg", "i18nmsg"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	src, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "package i18nmsg") {
		t.Errorf("generated file does not use the -pkg name:\n%s", src)
	}
}

// No t.Parallel: t.Chdir is incompatible with parallel tests.
func TestRun_outInCurrentDir(t *testing.T) { //nolint:paralleltest
	dir := filepath.Join(t.TempDir(), "myapp")
	if err := os.Mkdir(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")
	t.Chdir(dir)

	var stdout, stderr bytes.Buffer
	err := codegen.Run([]string{"-dir", ".", "-out", "messages.gen.go"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	src, err := os.ReadFile("messages.gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "package myapp") {
		t.Errorf("generated file does not derive the package from the current directory:\n%s", src)
	}
}

func TestRun_error(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")

	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"-nope"}},
		{name: "unexpected positional args", args: []string{"-dir", dir, "extra"}},
		{name: "invalid default language", args: []string{"-dir", dir, "-default", "not a tag"}},
		{name: "missing locale directory", args: []string{"-dir", filepath.Join(dir, "nope")}},
		{name: "invalid package name", args: []string{"-dir", dir, "-pkg", "my-pkg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			if err := codegen.Run(tt.args, &stdout, &stderr); err == nil {
				t.Errorf("Run(%v) returned nil error", tt.args)
			}
		})
	}
}
