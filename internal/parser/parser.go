package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Result contains the parsed messages.
// Used by the runtime Bundle for message lookup.
type Result struct {
	Messages map[string]string            // key -> template (non-plural)
	Plurals  map[string]map[string]string // key -> form -> template (plural)
}

// Message represents a parsed message with placeholder information.
// Used by the code generator to produce typed functions.
type Message struct {
	Key          string            // dot-separated key (e.g., "user.not_found")
	Template     string            // original template string (empty if plural)
	Placeholders []Placeholder     // placeholders in order of appearance
	Plural       map[string]string // plural forms: "one" -> template, "other" -> template
}

// IsPlural returns true if the message has plural forms.
func (m Message) IsPlural() bool {
	return len(m.Plural) > 0
}

// pluralForms are the valid CLDR plural category keys.
var pluralForms = map[string]struct{}{
	"zero":  {},
	"one":   {},
	"two":   {},
	"few":   {},
	"many":  {},
	"other": {},
}

// Placeholder represents a placeholder in a message template.
type Placeholder struct {
	Name string          // placeholder name (e.g., "name", "count")
	Type PlaceholderType // Go type
}

// PlaceholderType represents the type of placeholder.
type PlaceholderType int

const (
	TypeString PlaceholderType = iota
	TypeInt
	TypeFloat64
)

// String returns the Go type name for the placeholder type.
func (t PlaceholderType) String() string {
	switch t {
	case TypeInt:
		return "int"
	case TypeFloat64:
		return "float64"
	default:
		return "string"
	}
}

// Parser parses a locale file.
type Parser interface {
	// Parse returns a flat map of messages (for runtime).
	Parse(data []byte) (*Result, error)
	// ParseMessages returns messages with placeholder info (for codegen).
	ParseMessages(data []byte) ([]Message, error)
}

// ParseFile parses a file using the appropriate parser based on file extension.
// Used by the runtime Bundle.
func ParseFile(path string, data []byte) (*Result, error) {
	p, err := parserForExtension(filepath.Ext(path))
	if err != nil {
		return nil, err
	}
	return p.Parse(data)
}

// ParseMessagesFile parses a file and returns messages with placeholder info.
// Used by the code generator.
func ParseMessagesFile(path string, data []byte) ([]Message, error) {
	p, err := parserForExtension(filepath.Ext(path))
	if err != nil {
		return nil, err
	}
	messages, err := p.ParseMessages(data)
	if err != nil {
		return nil, err
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Key < messages[j].Key
	})
	return messages, nil
}

// parserForExtension returns the appropriate parser for the given file extension.
func parserForExtension(ext string) (Parser, error) {
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		return &YAMLParser{}, nil
	case ".json":
		return &JSONParser{}, nil
	case ".toml":
		return &TOMLParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// placeholderRegex matches placeholders like {name} or {count:int}.
var placeholderRegex = regexp.MustCompile(`\{(\w+)(?::(\w+))?\}`)

// extractPlaceholders extracts placeholders from a template string.
// Placeholders are returned in order of appearance.
func extractPlaceholders(template string) []Placeholder {
	matches := placeholderRegex.FindAllStringSubmatch(template, -1)
	seen := make(map[string]struct{})
	var placeholders []Placeholder

	for _, match := range matches {
		name := match[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		typ := TypeString
		if len(match) > 2 && match[2] != "" {
			typ = parseType(match[2])
		}

		placeholders = append(placeholders, Placeholder{
			Name: name,
			Type: typ,
		})
	}

	return placeholders
}

// parseType converts a type hint string to PlaceholderType.
func parseType(hint string) PlaceholderType {
	switch hint {
	case "int":
		return TypeInt
	case "float":
		return TypeFloat64
	default:
		return TypeString
	}
}

// buildMessages processes raw parsed data and detects plural forms.
func buildMessages(prefix string, raw map[string]any) ([]Message, error) {
	var messages []Message

	for k, v := range raw {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case string:
			messages = append(messages, Message{
				Key:          key,
				Template:     val,
				Placeholders: extractPlaceholders(val),
			})
		case map[string]any:
			if isPluralMap(val) {
				plural := make(map[string]string)
				var placeholders []Placeholder
				seen := make(map[string]struct{})

				for form, tmpl := range val {
					s, ok := tmpl.(string)
					if !ok {
						return nil, fmt.Errorf("invalid plural form value for key %q form %q: expected string", key, form)
					}
					plural[form] = s
					for _, p := range extractPlaceholders(s) {
						if _, ok := seen[p.Name]; !ok {
							seen[p.Name] = struct{}{}
							placeholders = append(placeholders, p)
						}
					}
				}

				messages = append(messages, Message{
					Key:          key,
					Plural:       plural,
					Placeholders: placeholders,
				})
			} else {
				nested, err := buildMessages(key, val)
				if err != nil {
					return nil, err
				}
				messages = append(messages, nested...)
			}
		}
	}
	return messages, nil
}

// isPluralMap checks if all keys in the map are valid plural forms.
func isPluralMap(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}
	for k, v := range m {
		if _, ok := pluralForms[k]; !ok {
			return false
		}
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}
