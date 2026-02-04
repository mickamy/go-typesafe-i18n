package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"

	"github.com/mickamy/go-typesafe-i18n/internal/naming"
	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

// Config holds the configuration for code generation.
type Config struct {
	InputPath   string
	OutputPath  string
	PackageName string
}

// Generate generates Go code from a locale file.
func Generate(cfg Config) error {
	data, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	messages, err := parser.ParseMessagesFile(cfg.InputPath, data)
	if err != nil {
		return fmt.Errorf("failed to parse input file: %w", err)
	}

	if err := naming.DetectCollisions(messages); err != nil {
		return err
	}

	code, err := generateCode(cfg.PackageName, messages)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	if err := os.WriteFile(cfg.OutputPath, code, 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// templateData holds data for the code template.
type templateData struct {
	PackageName string
	Messages    []messageData
}

// messageData holds data for a single message in the template.
type messageData struct {
	Key          string
	Template     string
	FuncName     string
	ParamList    string
	HasArgs      bool
	Placeholders []placeholderData
}

// placeholderData holds data for a placeholder in the template.
type placeholderData struct {
	Name      string
	ParamName string
}

func generateCode(packageName string, messages []parser.Message) ([]byte, error) {
	data := templateData{
		PackageName: packageName,
		Messages:    make([]messageData, 0, len(messages)),
	}

	for _, msg := range messages {
		md := messageData{
			Key:      msg.Key,
			Template: msg.Template,
			FuncName: naming.KeyToFunctionName(msg.Key),
			HasArgs:  len(msg.Placeholders) > 0,
		}

		if len(msg.Placeholders) > 0 {
			md.ParamList = buildParamList(msg.Placeholders)
			md.Placeholders = make([]placeholderData, 0, len(msg.Placeholders))
			for _, p := range msg.Placeholders {
				md.Placeholders = append(md.Placeholders, placeholderData{
					Name:      p.Name,
					ParamName: p.Name,
				})
			}
		}

		data.Messages = append(data.Messages, md)
	}

	var buf bytes.Buffer
	if err := codeTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format generated code: %w", err)
	}

	return formatted, nil
}

// buildParamList builds the parameter list for a function.
// e.g., "name string, count int, price float64"
func buildParamList(placeholders []parser.Placeholder) string {
	params := make([]string, 0, len(placeholders))
	for _, p := range placeholders {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type.String()))
	}
	return strings.Join(params, ", ")
}
