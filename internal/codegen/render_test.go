package codegen_test

import (
	"bytes"
	"flag"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

var update = flag.Bool("update", false, "update golden files")

func TestRender_golden(t *testing.T) {
	t.Parallel()

	got := renderBasic(t)

	golden := filepath.Join("testdata", "basic", "messages.gen.go.golden")
	if *update {
		if err := os.WriteFile(golden, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("Render() differs from %s; run go test ./internal/codegen -update to refresh\ngot:\n%s", golden, got)
	}
}

func TestRender_outputParses(t *testing.T) {
	t.Parallel()

	got := renderBasic(t)
	if _, err := parser.ParseFile(token.NewFileSet(), "messages.gen.go", got, parser.AllErrors); err != nil {
		t.Errorf("generated code does not parse: %v", err)
	}
}

func TestRender_deterministic(t *testing.T) {
	t.Parallel()

	if !bytes.Equal(renderBasic(t), renderBasic(t)) {
		t.Error("Render() is not deterministic")
	}
}

func TestRender_invalidPackageName(t *testing.T) {
	t.Parallel()

	for _, pkg := range []string{"", "1bad", "my-pkg"} {
		if _, err := codegen.Render(codegen.Model{}, pkg); err == nil {
			t.Errorf("Render(%q) returned nil error", pkg)
		}
	}
}

func renderBasic(t *testing.T) []byte {
	t.Helper()
	model, warnings, err := codegen.Analyze(filepath.Join("testdata", "basic", "locales"), language.English)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Analyze() returned warnings: %v", warnings)
	}
	src, err := codegen.Render(model, "messages")
	if err != nil {
		t.Fatal(err)
	}
	return src
}
