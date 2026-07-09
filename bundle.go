package i18n

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/mickamy/go-typesafe-i18n/internal/locale"
)

// Bundle holds the message catalogs of all loaded languages.
//
// Load every locale first, then hand out Localizers; a Bundle must not be
// mutated once Localizers are in use. Localizers are safe for concurrent
// use.
type Bundle struct {
	defaultTag language.Tag
	catalogs   map[language.Tag]locale.Catalog
	printers   map[language.Tag]*message.Printer
	tags       []language.Tag // default first, rest sorted; rebuilt on load
	matcher    language.Matcher
}

// NewBundle creates a Bundle that falls back to defaultLang for messages
// missing from the requested language.
func NewBundle(defaultLang language.Tag) *Bundle {
	return &Bundle{
		defaultTag: defaultLang,
		catalogs:   make(map[language.Tag]locale.Catalog),
		printers:   make(map[language.Tag]*message.Printer),
	}
}

// LoadFile loads a locale file. The language is derived from the filename
// stem (e.g., "en.yaml" is English) and the format from the extension
// (.yaml, .yml, or .toml).
func (b *Bundle) LoadFile(path string) error {
	c, err := locale.ParseFile(path)
	if err != nil {
		return fmt.Errorf("i18n: %w", err)
	}
	return b.add(c)
}

// MustLoadFile is like LoadFile but panics on error.
func (b *Bundle) MustLoadFile(path string) {
	if err := b.LoadFile(path); err != nil {
		panic(err)
	}
}

// LoadYAML loads YAML locale data for the given language. It is the
// primitive for sources other than files, such as embed.FS.
func (b *Bundle) LoadYAML(lang language.Tag, data []byte) error {
	c, err := locale.ParseYAML(lang, data)
	if err != nil {
		return fmt.Errorf("i18n: %w", err)
	}
	return b.add(c)
}

// LoadTOML loads TOML locale data for the given language.
func (b *Bundle) LoadTOML(lang language.Tag, data []byte) error {
	c, err := locale.ParseTOML(lang, data)
	if err != nil {
		return fmt.Errorf("i18n: %w", err)
	}
	return b.add(c)
}

func (b *Bundle) add(c locale.Catalog) error {
	if _, ok := b.catalogs[c.Tag]; ok {
		return fmt.Errorf("i18n: locale %s already loaded", c.Tag)
	}
	b.catalogs[c.Tag] = c
	b.printers[c.Tag] = message.NewPrinter(c.Tag)
	b.rebuildMatcher()
	return nil
}

// rebuildMatcher caches the matcher and its tag list so Localizer does not
// pay for matcher construction on every call. The default language comes
// first so that unmatched requests resolve to it; the rest are sorted for
// determinism.
func (b *Bundle) rebuildMatcher() {
	tags := make([]language.Tag, 0, len(b.catalogs))
	if _, ok := b.catalogs[b.defaultTag]; ok {
		tags = append(tags, b.defaultTag)
	}
	rest := make([]language.Tag, 0, len(b.catalogs))
	for t := range b.catalogs {
		if t != b.defaultTag {
			rest = append(rest, t)
		}
	}
	slices.SortFunc(rest, func(a, b language.Tag) int { return strings.Compare(a.String(), b.String()) })
	tags = append(tags, rest...)
	b.tags = tags
	b.matcher = language.NewMatcher(tags)
}
