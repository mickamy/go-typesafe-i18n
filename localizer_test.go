package i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n"
)

func TestLocalizer_Localize(t *testing.T) {
	t.Parallel()

	bundle := setupTestBundle(t)
	loc := bundle.Localizer(language.English)

	tests := []struct {
		name     string
		msg      i18n.Message
		expected string
	}{
		{
			name:     "simple message",
			msg:      i18n.Message{ID: "greeting"},
			expected: "Hello",
		},
		{
			name: "message with string arg",
			msg: i18n.Message{
				ID:   "hello",
				Args: map[string]any{"name": "John"},
			},
			expected: "Hello, John!",
		},
		{
			name: "message with int arg",
			msg: i18n.Message{
				ID:   "items_count",
				Args: map[string]any{"count": 5},
			},
			expected: "5 items",
		},
		{
			name: "message with float arg",
			msg: i18n.Message{
				ID:   "total_price",
				Args: map[string]any{"price": 1234.56},
			},
			expected: "Total: $1234.56",
		},
		{
			name: "message with multiple args",
			msg: i18n.Message{
				ID:   "transfer",
				Args: map[string]any{"from": "Alice", "to": "Bob", "amount": 1000},
			},
			expected: "Transfer 1000 from Alice to Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := loc.Localize(tt.msg)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLocalizer_Localize_Escape(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "en.yaml")
	content := `
escaped: "Use \\{name\\} for placeholders"
mixed: "Hello \\{literal\\} and {name}"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile(path)
	loc := bundle.Localizer(language.English)

	tests := []struct {
		name     string
		msg      i18n.Message
		expected string
	}{
		{
			name:     "escaped braces",
			msg:      i18n.Message{ID: "escaped"},
			expected: "Use {name} for placeholders",
		},
		{
			name: "mixed escaped and placeholder",
			msg: i18n.Message{
				ID:   "mixed",
				Args: map[string]any{"name": "World"},
			},
			expected: "Hello {literal} and World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := loc.Localize(tt.msg)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLocalizer_Localize_Fallback(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// English has both keys
	enPath := filepath.Join(tmpDir, "en.yaml")
	enContent := `
common: "Common message"
english_only: "English only"
`
	if err := os.WriteFile(enPath, []byte(enContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Japanese has only common key
	jaPath := filepath.Join(tmpDir, "ja.yaml")
	jaContent := `common: "Common in Japanese"`
	if err := os.WriteFile(jaPath, []byte(jaContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile(enPath)
	bundle.MustLoadFile(jaPath)

	loc := bundle.Localizer(language.Japanese)

	// Should get Japanese message
	got := loc.Localize(i18n.Message{ID: "common"})
	if got != "Common in Japanese" {
		t.Errorf("expected Japanese message, got %q", got)
	}

	// Should fallback to English
	got = loc.Localize(i18n.Message{ID: "english_only"})
	if got != "English only" {
		t.Errorf("expected English fallback, got %q", got)
	}
}

func TestLocalizer_Localize_Panics(t *testing.T) {
	t.Parallel()

	bundle := setupTestBundle(t)
	loc := bundle.Localizer(language.English)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	loc.Localize(i18n.Message{ID: "nonexistent"})
}

func setupTestBundle(t *testing.T) *i18n.Bundle {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "en.yaml")
	content := `
greeting: "Hello"
hello: "Hello, {name}!"
items_count: "{count:int} items"
total_price: "Total: ${price:float}"
transfer: "Transfer {amount:int} from {from} to {to}"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile(path)
	return bundle
}
