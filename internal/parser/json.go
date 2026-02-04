package parser

import (
	"encoding/json"
	"fmt"
)

// JSONParser parses JSON locale files.
type JSONParser struct{}

// Parse parses JSON data and returns messages and plurals.
func (p *JSONParser) Parse(data []byte) (*Result, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	messages, plurals := flatten("", raw)
	return &Result{Messages: messages, Plurals: plurals}, nil
}

// ParseMessages parses JSON data and returns messages with placeholder information.
func (p *JSONParser) ParseMessages(data []byte) ([]Message, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return buildMessages("", raw)
}
