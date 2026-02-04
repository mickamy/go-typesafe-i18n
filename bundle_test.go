package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"
)

func TestNewBundle(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(language.English)

	if bundle.defaultLang != language.English {
		t.Errorf("expected default lang %v, got %v", language.English, bundle.defaultLang)
	}
	if bundle.messages == nil {
		t.Error("expected messages map to be initialized")
	}
}

func TestBundle_LoadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
		wantLang language.Tag
		wantKeys []string
		wantErr  bool
	}{
		{
			name:     "simple yaml",
			filename: "en.yaml",
			content:  "greeting: Hello",
			wantLang: language.English,
			wantKeys: []string{"greeting"},
		},
		{
			name:     "nested yaml",
			filename: "ja.yaml",
			content: `
user:
  not_found: "ユーザーが見つかりません"
  deleted: "削除しました"
`,
			wantLang: language.Japanese,
			wantKeys: []string{"user.not_found", "user.deleted"},
		},
		{
			name:     "yml extension",
			filename: "fr.yml",
			content:  "bonjour: Bonjour",
			wantLang: language.French,
			wantKeys: []string{"bonjour"},
		},
		{
			name:     "invalid language",
			filename: "invalid-lang-tag.yaml",
			content:  "key: value",
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

			bundle := NewBundle(language.English)
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

			msgs, ok := bundle.messages[tt.wantLang]
			if !ok {
				t.Fatalf("messages not found for language %v", tt.wantLang)
			}

			for _, key := range tt.wantKeys {
				if _, ok := msgs[key]; !ok {
					t.Errorf("missing key: %s", key)
				}
			}
		})
	}
}

func TestBundle_MustLoadFile_Panics(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(language.English)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	bundle.MustLoadFile("/nonexistent/path/en.yaml")
}

func TestBundle_Localizer(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(language.English)
	loc := bundle.Localizer(language.Japanese)

	if loc.bundle != bundle {
		t.Error("localizer should reference the bundle")
	}
	if loc.lang != language.Japanese {
		t.Errorf("expected lang %v, got %v", language.Japanese, loc.lang)
	}
}

func TestInferLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		wantLang language.Tag
		wantErr  bool
	}{
		{"en.yaml", language.English, false},
		{"ja.yaml", language.Japanese, false},
		{"zh-Hans.yaml", language.SimplifiedChinese, false},
		{"/path/to/locales/fr.yml", language.French, false},
		{"invalid-lang-tag.yaml", language.Tag{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			got, err := inferLanguage(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.wantLang {
				t.Errorf("expected %v, got %v", tt.wantLang, got)
			}
		})
	}
}
