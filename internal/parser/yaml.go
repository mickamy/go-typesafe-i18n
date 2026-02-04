package parser

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAMLParser parses YAML locale files.
type YAMLParser struct{}

// Parse parses YAML data and returns a flat map of messages.
func (p *YAMLParser) Parse(data []byte) (*Result, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	messages := make(map[string]string)
	flatten("", raw, messages)

	return &Result{Messages: messages}, nil
}

// ParseMessages parses YAML data and returns messages with placeholder information.
func (p *YAMLParser) ParseMessages(data []byte) ([]Message, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return buildMessages("", raw)
}
