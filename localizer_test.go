package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"
)

func TestLocalizer_Localize(t *testing.T) {
	t.Parallel()

	bundle := setupTestBundle(t)
	loc := bundle.Localizer(language.English)

	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{
			name:     "simple message",
			msg:      Message{ID: "greeting"},
			expected: "Hello",
		},
		{
			name: "message with string arg",
			msg: Message{
				ID:   "hello",
				Args: map[string]any{"name": "John"},
			},
			expected: "Hello, John!",
		},
		{
			name: "message with int arg",
			msg: Message{
				ID:   "items_count",
				Args: map[string]any{"count": 5},
			},
			expected: "5 items",
		},
		{
			name: "message with float arg",
			msg: Message{
				ID:   "total_price",
				Args: map[string]any{"price": 1234.56},
			},
			expected: "Total: $1234.56",
		},
		{
			name: "message with multiple args",
			msg: Message{
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

	bundle := NewBundle(language.English)
	bundle.MustLoadFile(path)
	loc := bundle.Localizer(language.English)

	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{
			name:     "escaped braces",
			msg:      Message{ID: "escaped"},
			expected: "Use {name} for placeholders",
		},
		{
			name: "mixed escaped and placeholder",
			msg: Message{
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
	jaContent := `common: "共通メッセージ"`
	if err := os.WriteFile(jaPath, []byte(jaContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bundle := NewBundle(language.English)
	bundle.MustLoadFile(enPath)
	bundle.MustLoadFile(jaPath)

	loc := bundle.Localizer(language.Japanese)

	// Should get Japanese message
	got := loc.Localize(Message{ID: "common"})
	if got != "共通メッセージ" {
		t.Errorf("expected Japanese message, got %q", got)
	}

	// Should fallback to English
	got = loc.Localize(Message{ID: "english_only"})
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

	loc.Localize(Message{ID: "nonexistent"})
}

func setupTestBundle(t *testing.T) *Bundle {
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

	bundle := NewBundle(language.English)
	bundle.MustLoadFile(path)
	return bundle
}
