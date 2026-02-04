package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Result contains the parsed messages as a flat map.
// Used by the runtime Bundle for message lookup.
type Result struct {
	Messages map[string]string // key -> template (flattened)
}

// Message represents a parsed message with placeholder information.
// Used by the code generator to produce typed functions.
type Message struct {
	Key          string        // dot-separated key (e.g., "user.not_found")
	Template     string        // original template string
	Placeholders []Placeholder // placeholders in order of appearance
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
	return p.ParseMessages(data)
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
	seen := make(map[string]bool)
	var placeholders []Placeholder

	for _, match := range matches {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true

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

// buildMessages converts a flat map to a sorted slice of Messages.
func buildMessages(flat map[string]string) []Message {
	var messages []Message
	for key, template := range flat {
		messages = append(messages, Message{
			Key:          key,
			Template:     template,
			Placeholders: extractPlaceholders(template),
		})
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Key < messages[j].Key
	})

	return messages
}
