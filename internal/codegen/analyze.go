package codegen

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n/internal/locale"
	"github.com/mickamy/go-typesafe-i18n/internal/template"
)

// Model is the input to code generation, derived from the default locale.
type Model struct {
	DefaultTag language.Tag
	Messages   []Message // sorted by key
}

// Message is one message constructor to generate.
type Message struct {
	Key      string
	FuncName string
	Plural   bool
	Params   []Param
}

// Param is one parameter of a generated constructor.
type Param struct {
	Name   string // placeholder name in the template
	GoName string // camelCase Go parameter name
	GoType string
}

// Warning is a non-fatal finding, such as a locale missing translations.
type Warning string

// Analyze loads every locale file in dir, builds the generation model from
// the default locale, and validates the other locales against it.
func Analyze(dir string, defaultLang language.Tag) (Model, []Warning, error) {
	catalogs, err := loadCatalogs(dir)
	if err != nil {
		return Model{}, nil, err
	}
	at := slices.IndexFunc(catalogs, func(c locale.Catalog) bool { return c.Tag == defaultLang })
	if at < 0 {
		return Model{}, nil, fmt.Errorf("default locale %s not found in %s", defaultLang, dir)
	}
	def := catalogs[at]
	model, err := buildModel(def)
	if err != nil {
		return Model{}, nil, err
	}
	var warnings []Warning
	for _, c := range catalogs {
		if c.Tag == defaultLang {
			continue
		}
		w, err := crossCheck(def, c)
		if err != nil {
			return Model{}, nil, err
		}
		warnings = append(warnings, w...)
	}
	return model, warnings, nil
}

func loadCatalogs(dir string) ([]locale.Catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read locale directory: %w", err)
	}
	seen := make(map[language.Tag]string)
	var catalogs []locale.Catalog
	for _, e := range entries {
		if e.IsDir() || !isLocaleFile(e.Name()) {
			continue
		}
		c, err := locale.ParseFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("load locale file: %w", err)
		}
		if prev, ok := seen[c.Tag]; ok {
			return nil, fmt.Errorf("locale %s defined by both %s and %s", c.Tag, prev, e.Name())
		}
		seen[c.Tag] = e.Name()
		catalogs = append(catalogs, c)
	}
	if len(catalogs) == 0 {
		return nil, fmt.Errorf("no locale files found in %s", dir)
	}
	return catalogs, nil
}

func isLocaleFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".yaml", ".yml", ".toml":
		return true
	default:
		return false
	}
}

func buildModel(def locale.Catalog) (Model, error) {
	keys := slices.Sorted(maps.Keys(def.Entries))
	funcNames := make(map[string]string, len(keys))
	messages := make([]Message, 0, len(keys))
	for _, key := range keys {
		entry := def.Entries[key]
		params, err := entry.Params()
		if err != nil {
			return Model{}, fmt.Errorf("locale %s: %w", def.Tag, err)
		}
		msg, err := buildMessage(entry, params)
		if err != nil {
			return Model{}, fmt.Errorf("locale %s: %w", def.Tag, err)
		}
		if prev, ok := funcNames[msg.FuncName]; ok {
			return Model{}, fmt.Errorf("keys %q and %q both generate func %s", prev, key, msg.FuncName)
		}
		funcNames[msg.FuncName] = key
		messages = append(messages, msg)
	}
	return Model{DefaultTag: def.Tag, Messages: messages}, nil
}

func buildMessage(entry locale.Entry, params []template.Param) (Message, error) {
	msg := Message{
		Key:      entry.Key,
		FuncName: FuncName(entry.Key),
		Plural:   entry.Plural != nil,
		Params:   make([]Param, 0, len(params)),
	}
	goNames := make(map[string]string, len(params))
	for _, p := range params {
		goName, err := ParamName(p.Name)
		if err != nil {
			return Message{}, fmt.Errorf("key %q: %w", entry.Key, err)
		}
		if prev, ok := goNames[goName]; ok {
			return Message{}, fmt.Errorf(
				"key %q: parameters %q and %q both map to Go parameter %q",
				entry.Key, prev, p.Name, goName,
			)
		}
		goNames[goName] = p.Name
		msg.Params = append(msg.Params, Param{Name: p.Name, GoName: goName, GoType: goType(p.Kind)})
	}
	return msg, nil
}

func goType(k template.Kind) string {
	switch k {
	case template.KindInt:
		return "int"
	case template.KindNumber:
		return "float64"
	case template.KindString:
		return "string"
	default:
		return "string"
	}
}

// crossCheck validates a translation against the default catalog: unknown
// keys, shape mismatches, and unknown parameters are errors, while keys
// missing from the translation are warnings because the runtime falls back
// to the default language.
func crossCheck(def, other locale.Catalog) ([]Warning, error) {
	for _, key := range slices.Sorted(maps.Keys(other.Entries)) {
		entry := other.Entries[key]
		defEntry, ok := def.Entries[key]
		if !ok {
			return nil, fmt.Errorf("locale %s: key %q does not exist in default locale %s", other.Tag, key, def.Tag)
		}
		if (entry.Plural != nil) != (defEntry.Plural != nil) {
			return nil, fmt.Errorf("locale %s: key %q: plural shape differs from default locale", other.Tag, key)
		}
		defParams, err := defEntry.Params()
		if err != nil {
			return nil, fmt.Errorf("locale %s: %w", def.Tag, err)
		}
		known := make(map[string]bool, len(defParams))
		for _, p := range defParams {
			known[p.Name] = true
		}
		params, err := entry.Params()
		if err != nil {
			return nil, fmt.Errorf("locale %s: %w", other.Tag, err)
		}
		for _, p := range params {
			if !known[p.Name] {
				return nil, fmt.Errorf(
					"locale %s: key %q: parameter %q does not exist in default locale",
					other.Tag, key, p.Name,
				)
			}
		}
	}
	var missing []string
	for _, key := range slices.Sorted(maps.Keys(def.Entries)) {
		if _, ok := other.Entries[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return []Warning{Warning(fmt.Sprintf("locale %s: missing keys: %s", other.Tag, strings.Join(missing, ", ")))}, nil
	}
	return nil, nil
}
