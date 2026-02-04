package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Result contains the parsed messages as a flat map.
type Result struct {
	Messages map[string]string // key -> template (flattened)
}

// Parser parses a locale file and returns a flat map of messages.
type Parser interface {
	Parse(data []byte) (*Result, error)
}

// ParseFile parses a file using the appropriate parser based on file extension.
func ParseFile(path string, data []byte) (*Result, error) {
	parser, err := parserForExtension(filepath.Ext(path))
	if err != nil {
		return nil, err
	}
	return parser.Parse(data)
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
