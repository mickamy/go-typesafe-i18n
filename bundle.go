package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

// Bundle manages messages for multiple languages.
type Bundle struct {
	defaultLang language.Tag
	messages    map[language.Tag]map[string]string            // lang -> id -> template
	plurals     map[language.Tag]map[string]map[string]string // lang -> id -> form -> template
	matcher     language.Matcher
}

// NewBundle creates a new Bundle with the specified default language.
func NewBundle(defaultLang language.Tag) *Bundle {
	return &Bundle{
		defaultLang: defaultLang,
		messages:    make(map[language.Tag]map[string]string),
		plurals:     make(map[language.Tag]map[string]map[string]string),
	}
}

// MustLoadFile loads a locale file and panics if an error occurs.
// The language is inferred from the filename (e.g., "ja.yaml" -> "ja").
func (b *Bundle) MustLoadFile(path string) {
	if err := b.LoadFile(path); err != nil {
		panic(fmt.Sprintf("go-typesafe-i18n: failed to load file %s: %v", path, err))
	}
}

// LoadFile loads a locale file into the bundle.
// The language is inferred from the filename (e.g., "ja.yaml" -> "ja").
// Supported formats: YAML (.yaml, .yml)
func (b *Bundle) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return b.LoadBytes(filepath.Base(path), data)
}

// MustLoadBytes loads locale data from bytes and panics if an error occurs.
// The filename is used to infer the language and format (e.g., "ja.yaml" -> "ja").
func (b *Bundle) MustLoadBytes(filename string, data []byte) {
	if err := b.LoadBytes(filename, data); err != nil {
		panic(fmt.Sprintf("go-typesafe-i18n: failed to load bytes for %s: %v", filename, err))
	}
}

// LoadBytes loads locale data from bytes into the bundle.
// The filename is used to infer the language and format (e.g., "ja.yaml" -> "ja").
func (b *Bundle) LoadBytes(filename string, data []byte) error {
	lang, err := inferLanguage(filename)
	if err != nil {
		return fmt.Errorf("failed to infer language from filename: %w", err)
	}

	result, err := parser.ParseFile(filename, data)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	b.messages[lang] = result.Messages
	b.plurals[lang] = result.Plurals
	b.rebuildMatcher()
	return nil
}

// Localizer returns a Localizer for the specified language.
// It uses language matching to find the best available language.
func (b *Bundle) Localizer(lang language.Tag) *Localizer {
	return &Localizer{bundle: b, lang: lang}
}

// rebuildMatcher rebuilds the language matcher with all loaded languages.
func (b *Bundle) rebuildMatcher() {
	tags := make([]language.Tag, 0, len(b.messages))
	for tag := range b.messages {
		tags = append(tags, tag)
	}
	b.matcher = language.NewMatcher(tags)
}

// inferLanguage extracts the language tag from a file path.
// e.g., "resources/ja.yaml" -> language.Japanese
func inferLanguage(path string) (language.Tag, error) {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	langStr := strings.TrimSuffix(base, ext)
	return language.Parse(langStr)
}
