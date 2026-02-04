package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		packageName string
		wantErr     bool
		contains    []string
	}{
		{
			name:        "simple message",
			input:       `greeting: "Hello"`,
			packageName: "messages",
			contains: []string{
				"package messages",
				"func Greeting() i18n.Message",
				`ID: "greeting"`,
			},
		},
		{
			name:        "message with string placeholder",
			input:       `hello: "Hello, {name}!"`,
			packageName: "messages",
			contains: []string{
				"func Hello(name string) i18n.Message",
				`"name": name`,
			},
		},
		{
			name:        "message with int placeholder",
			input:       `count: "{count:int} items"`,
			packageName: "messages",
			contains: []string{
				"func Count(count int) i18n.Message",
				`"count": count`,
			},
		},
		{
			name:        "message with float placeholder",
			input:       `price: "Total: ${price:float}"`,
			packageName: "messages",
			contains: []string{
				"func Price(price float64) i18n.Message",
				`"price": price`,
			},
		},
		{
			name:        "message with multiple placeholders",
			input:       `transfer: "Transfer {amount:int} from {from} to {to}"`,
			packageName: "messages",
			contains: []string{
				"func Transfer(amount int, from string, to string) i18n.Message",
				`"amount"`,
				`"from"`,
				`"to"`,
			},
		},
		{
			name: "nested keys",
			input: `
user:
  not_found: "User not found"
  deleted: "User deleted"
`,
			packageName: "messages",
			contains: []string{
				"func UserNotFound() i18n.Message",
				"func UserDeleted() i18n.Message",
				`ID: "user.not_found"`,
				`ID: "user.deleted"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			inputPath := filepath.Join(tmpDir, "en.yaml")
			outputPath := filepath.Join(tmpDir, "messages.go")

			if err := os.WriteFile(inputPath, []byte(tt.input), 0o644); err != nil {
				t.Fatalf("failed to write input file: %v", err)
			}

			cfg := codegen.Config{
				InputPath:   inputPath,
				OutputPath:  outputPath,
				PackageName: tt.packageName,
			}

			err := codegen.Generate(cfg)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			outputStr := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output should contain %q\noutput:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestGenerate_Collision(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "en.yaml")
	outputPath := filepath.Join(tmpDir, "messages.go")

	input := `
user.name: "User name"
user_name: "Username"
`
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cfg := codegen.Config{
		InputPath:   inputPath,
		OutputPath:  outputPath,
		PackageName: "messages",
	}

	err := codegen.Generate(cfg)
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}

	if !strings.Contains(err.Error(), "collision") {
		t.Errorf("error should mention collision, got: %v", err)
	}
}

func TestGenerate_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "en.toml")
	outputPath := filepath.Join(tmpDir, "messages.go")

	if err := os.WriteFile(inputPath, []byte(`greeting = "Hello"`), 0o644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cfg := codegen.Config{
		InputPath:   inputPath,
		OutputPath:  outputPath,
		PackageName: "messages",
	}

	err := codegen.Generate(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported extension, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported extension, got: %v", err)
	}
}
