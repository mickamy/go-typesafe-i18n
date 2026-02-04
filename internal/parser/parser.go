package parser

import (
	"fmt"
	"path/filepath"
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
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
