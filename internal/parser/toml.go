package parser

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// TOMLParser parses TOML locale files.
type TOMLParser struct{}

// Parse parses TOML data and returns messages and plurals.
func (p *TOMLParser) Parse(data []byte) (*Result, error) {
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	messages, plurals := flatten("", raw)
	return &Result{Messages: messages, Plurals: plurals}, nil
}

// ParseMessages parses TOML data and returns messages with placeholder information.
func (p *TOMLParser) ParseMessages(data []byte) ([]Message, error) {
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return buildMessages("", raw)
}
