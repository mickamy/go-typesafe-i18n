package i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"

	i18n "github.com/mickamy/go-typesafe-i18n"
)

func TestBundle_LoadFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")
	writeFile(t, filepath.Join(dir, "ja.toml"), "greeting = \"こんにちは！\"\n")

	b := i18n.NewBundle(language.English)
	if err := b.LoadFile(filepath.Join(dir, "en.yaml")); err != nil {
		t.Fatalf("LoadFile(en.yaml) returned error: %v", err)
	}
	if err := b.LoadFile(filepath.Join(dir, "ja.toml")); err != nil {
		t.Fatalf("LoadFile(ja.toml) returned error: %v", err)
	}
	for _, tag := range []language.Tag{language.English, language.Japanese} {
		if !b.Loaded(tag) {
			t.Errorf("Loaded(%v) = false, want true", tag)
		}
	}
}

func TestBundle_LoadFile_error(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: 123\n")

	b := i18n.NewBundle(language.English)
	if err := b.LoadFile(filepath.Join(dir, "fr.yaml")); err == nil {
		t.Error("LoadFile() returned nil error for a missing file")
	}
	if err := b.LoadFile(filepath.Join(dir, "en.yaml")); err == nil {
		t.Error("LoadFile() returned nil error for invalid content")
	}
}

func TestBundle_LoadFile_duplicateLanguage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en.yaml"), "greeting: \"Hello!\"\n")

	b := i18n.NewBundle(language.English)
	if err := b.LoadFile(filepath.Join(dir, "en.yaml")); err != nil {
		t.Fatalf("LoadFile() returned error: %v", err)
	}
	if err := b.LoadYAML(language.English, []byte("greeting: \"Hi!\"\n")); err == nil {
		t.Error("loading the same language twice returned nil error")
	}
}

func TestBundle_MustLoadFile_panics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Error("MustLoadFile() did not panic")
		}
	}()
	i18n.NewBundle(language.English).MustLoadFile("does-not-exist.yaml")
}

func TestBundle_LoadYAML(t *testing.T) {
	t.Parallel()

	b := i18n.NewBundle(language.English)
	if err := b.LoadYAML(language.English, []byte("greeting: \"Hello!\"\n")); err != nil {
		t.Fatalf("LoadYAML() returned error: %v", err)
	}
	if !b.Loaded(language.English) {
		t.Error("Loaded(en) = false, want true")
	}
	if err := b.LoadYAML(language.Japanese, []byte("greeting: 123\n")); err == nil {
		t.Error("LoadYAML() returned nil error for invalid data")
	}
}

func TestBundle_LoadTOML(t *testing.T) {
	t.Parallel()

	b := i18n.NewBundle(language.English)
	if err := b.LoadTOML(language.Japanese, []byte("greeting = \"こんにちは！\"\n")); err != nil {
		t.Fatalf("LoadTOML() returned error: %v", err)
	}
	if !b.Loaded(language.Japanese) {
		t.Error("Loaded(ja) = false, want true")
	}
	if err := b.LoadTOML(language.English, []byte("greeting = 123\n")); err == nil {
		t.Error("LoadTOML() returned nil error for invalid data")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
