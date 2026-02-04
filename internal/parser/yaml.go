package parser

import (
	"fmt"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

// placeholderRegex matches placeholders like {name} or {count:int}.
var placeholderRegex = regexp.MustCompile(`\{(\w+)(?::(\w+))?\}`)

// YAMLParser parses YAML locale files.
type YAMLParser struct{}

// Parse parses YAML data and returns a flat map of messages.
// Used by the runtime Bundle.
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
// Used by the code generator.
func (p *YAMLParser) ParseMessages(data []byte) ([]Message, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	flat := make(map[string]string)
	flatten("", raw, flat)

	var messages []Message
	for key, template := range flat {
		messages = append(messages, Message{
			Key:          key,
			Template:     template,
			Placeholders: extractPlaceholders(template),
		})
	}

	// Sort by key for deterministic output
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Key < messages[j].Key
	})

	return messages, nil
}

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
