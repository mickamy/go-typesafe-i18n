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
	def, err := defaultCatalog(catalogs, defaultLang, dir)
	if err != nil {
		return Model{}, nil, err
	}
	model, err := buildModel(def)
	if err != nil {
		return Model{}, nil, err
	}
	var warnings []Warning
	for _, c := range catalogs {
		if c.Tag == def.Tag {
			continue
		}
		w, err := crossCheck(model, c)
		if err != nil {
			return Model{}, nil, err
		}
		warnings = append(warnings, w...)
	}
	return model, warnings, nil
}

// defaultCatalog resolves the default locale among the loaded catalogs the
// same way the runtime matches languages, so en-US.yaml satisfies -default
// en. Low-confidence matches are rejected to avoid picking an unrelated
// language.
func defaultCatalog(catalogs []locale.Catalog, defaultLang language.Tag, dir string) (locale.Catalog, error) {
	tags := make([]language.Tag, len(catalogs))
	for i, c := range catalogs {
		tags[i] = c.Tag
	}
	_, idx, conf := language.NewMatcher(tags).Match(defaultLang)
	if conf < language.High {
		return locale.Catalog{}, fmt.Errorf("default locale %s not found in %s (available: %v)", defaultLang, dir, tags)
	}
	return catalogs[idx], nil
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
		if goName == importAlias {
			return Message{}, fmt.Errorf(
				"key %q: parameter %q conflicts with the generated import alias %q",
				entry.Key, p.Name, importAlias,
			)
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

// crossCheck validates a translation against the generation model: unknown
// keys, shape mismatches, and unknown parameters are errors, while keys
// missing from the translation are warnings because the runtime falls back
// to the default language. Reusing the model avoids rebuilding the default
// entries' parameter lists per locale.
func crossCheck(model Model, other locale.Catalog) ([]Warning, error) {
	defMessages := make(map[string]Message, len(model.Messages))
	for _, msg := range model.Messages {
		defMessages[msg.Key] = msg
	}
	for _, key := range slices.Sorted(maps.Keys(other.Entries)) {
		entry := other.Entries[key]
		defMsg, ok := defMessages[key]
		if !ok {
			return nil, fmt.Errorf("locale %s: key %q does not exist in default locale %s", other.Tag, key, model.DefaultTag)
		}
		if (entry.Plural != nil) != defMsg.Plural {
			return nil, fmt.Errorf("locale %s: key %q: plural shape differs from default locale", other.Tag, key)
		}
		known := make(map[string]bool, len(defMsg.Params))
		for _, p := range defMsg.Params {
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
	for _, msg := range model.Messages {
		if _, ok := other.Entries[msg.Key]; !ok {
			missing = append(missing, msg.Key)
		}
	}
	if len(missing) > 0 {
		return []Warning{Warning(fmt.Sprintf("locale %s: missing keys: %s", other.Tag, strings.Join(missing, ", ")))}, nil
	}
	return nil, nil
}
