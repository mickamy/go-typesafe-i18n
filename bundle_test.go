package i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n"
)

func TestNewBundle(t *testing.T) {
	t.Parallel()

	// Just verify it doesn't panic
	bundle := i18n.NewBundle(language.English)
	if bundle == nil {
		t.Error("expected non-nil bundle")
	}
}

func TestBundle_LoadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "simple yaml",
			filename: "en.yaml",
			content:  "greeting: Hello",
			wantErr:  false,
		},
		{
			name:     "nested yaml",
			filename: "ja.yaml",
			content: `
user:
  not_found: "User not found"
  deleted: "User deleted"
`,
			wantErr: false,
		},
		{
			name:     "yml extension",
			filename: "fr.yml",
			content:  "bonjour: Bonjour",
			wantErr:  false,
		},
		{
			name:     "invalid language tag",
			filename: "invalid-lang-tag.yaml",
			content:  "key: value",
			wantErr:  true,
		},
		{
			name:     "invalid yaml",
			filename: "en.yaml",
			content:  "invalid: [",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)

			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			bundle := i18n.NewBundle(language.English)
			err := bundle.LoadFile(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBundle_LoadFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	bundle := i18n.NewBundle(language.English)
	err := bundle.LoadFile("/nonexistent/path/en.yaml")

	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestBundle_MustLoadFile_Panics(t *testing.T) {
	t.Parallel()

	bundle := i18n.NewBundle(language.English)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	bundle.MustLoadFile("/nonexistent/path/en.yaml")
}

func TestBundle_Localizer(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "en.yaml")
	content := `greeting: "Hello"`

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile(path)

	loc := bundle.Localizer(language.English)
	if loc == nil {
		t.Error("expected non-nil localizer")
	}

	// Verify localizer works
	msg := loc.Localize(i18n.Message{ID: "greeting"})
	if msg != "Hello" {
		t.Errorf("expected %q, got %q", "Hello", msg)
	}
}
