package parser_test

import (
	"reflect"
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

func TestYAMLParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:  "simple key-value",
			input: `greeting: "Hello"`,
			expected: map[string]string{
				"greeting": "Hello",
			},
		},
		{
			name: "nested keys",
			input: `
user:
  not_found: "User not found"
  deleted: "User deleted"
`,
			expected: map[string]string{
				"user.not_found": "User not found",
				"user.deleted":   "User deleted",
			},
		},
		{
			name: "deeply nested keys",
			input: `
errors:
  validation:
    required: "This field is required"
    invalid: "Invalid value"
`,
			expected: map[string]string{
				"errors.validation.required": "This field is required",
				"errors.validation.invalid":  "Invalid value",
			},
		},
		{
			name: "with placeholders",
			input: `
greeting: "Hello, {name}!"
items_count: "{count:int} items"
total_price: "Total: ${price:float}"
`,
			expected: map[string]string{
				"greeting":    "Hello, {name}!",
				"items_count": "{count:int} items",
				"total_price": "Total: ${price:float}",
			},
		},
		{
			name:    "invalid yaml",
			input:   `invalid: [`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.YAMLParser{}
			result, err := p.Parse([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Messages) != len(tt.expected) {
				t.Errorf("expected %d messages, got %d", len(tt.expected), len(result.Messages))
			}

			for key, want := range tt.expected {
				got, ok := result.Messages[key]
				if !ok {
					t.Errorf("missing key: %s", key)
					continue
				}
				if got != want {
					t.Errorf("key %s: expected %q, got %q", key, want, got)
				}
			}
		})
	}
}

func TestYAMLParser_ParseMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []parser.Message
		wantErr  bool
	}{
		{
			name:  "no placeholders",
			input: `greeting: "Hello"`,
			expected: []parser.Message{
				{
					Key:          "greeting",
					Template:     "Hello",
					Placeholders: nil,
				},
			},
		},
		{
			name:  "string placeholder",
			input: `hello: "Hello, {name}!"`,
			expected: []parser.Message{
				{
					Key:      "hello",
					Template: "Hello, {name}!",
					Placeholders: []parser.Placeholder{
						{Name: "name", Type: parser.TypeString},
					},
				},
			},
		},
		{
			name:  "int placeholder",
			input: `count: "{count:int} items"`,
			expected: []parser.Message{
				{
					Key:      "count",
					Template: "{count:int} items",
					Placeholders: []parser.Placeholder{
						{Name: "count", Type: parser.TypeInt},
					},
				},
			},
		},
		{
			name:  "float placeholder",
			input: `price: "Total: ${price:float}"`,
			expected: []parser.Message{
				{
					Key:      "price",
					Template: "Total: ${price:float}",
					Placeholders: []parser.Placeholder{
						{Name: "price", Type: parser.TypeFloat64},
					},
				},
			},
		},
		{
			name:  "multiple placeholders",
			input: `transfer: "Transfer {amount:int} from {from} to {to}"`,
			expected: []parser.Message{
				{
					Key:      "transfer",
					Template: "Transfer {amount:int} from {from} to {to}",
					Placeholders: []parser.Placeholder{
						{Name: "amount", Type: parser.TypeInt},
						{Name: "from", Type: parser.TypeString},
						{Name: "to", Type: parser.TypeString},
					},
				},
			},
		},
		{
			name:  "duplicate placeholder uses first occurrence",
			input: `repeat: "{name} and {name} again"`,
			expected: []parser.Message{
				{
					Key:      "repeat",
					Template: "{name} and {name} again",
					Placeholders: []parser.Placeholder{
						{Name: "name", Type: parser.TypeString},
					},
				},
			},
		},
		{
			name: "nested keys",
			input: `
user:
  greeting: "Hello, {name}!"
`,
			expected: []parser.Message{
				{
					Key:      "user.greeting",
					Template: "Hello, {name}!",
					Placeholders: []parser.Placeholder{
						{Name: "name", Type: parser.TypeString},
					},
				},
			},
		},
		{
			name:    "invalid yaml",
			input:   `invalid: [`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.YAMLParser{}
			messages, err := p.ParseMessages([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(messages) != len(tt.expected) {
				t.Fatalf("expected %d messages, got %d", len(tt.expected), len(messages))
			}

			for i, want := range tt.expected {
				got := messages[i]
				if got.Key != want.Key {
					t.Errorf("message %d: expected key %q, got %q", i, want.Key, got.Key)
				}
				if got.Template != want.Template {
					t.Errorf("message %d: expected template %q, got %q", i, want.Template, got.Template)
				}
				if !reflect.DeepEqual(got.Placeholders, want.Placeholders) {
					t.Errorf("message %d: expected placeholders %+v, got %+v", i, want.Placeholders, got.Placeholders)
				}
			}
		})
	}
}

func TestYAMLParser_ParseMessages_Plural(t *testing.T) {
	t.Parallel()

	input := `
items_count:
  one: "You have 1 item"
  other: "You have {count:int} items"
greeting: "Hello"
`

	p := &parser.YAMLParser{}
	messages, err := p.ParseMessages([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	// Build map for order-independent lookup
	byKey := make(map[string]parser.Message)
	for _, m := range messages {
		byKey[m.Key] = m
	}

	greeting, ok := byKey["greeting"]
	if !ok {
		t.Fatal("missing key 'greeting'")
	}
	if greeting.IsPlural() {
		t.Error("greeting should not be plural")
	}

	items, ok := byKey["items_count"]
	if !ok {
		t.Fatal("missing key 'items_count'")
	}
	if !items.IsPlural() {
		t.Error("items_count should be plural")
	}
	if items.Plural["one"] != "You have 1 item" {
		t.Errorf("expected one form, got %q", items.Plural["one"])
	}
	if items.Plural["other"] != "You have {count:int} items" {
		t.Errorf("expected other form, got %q", items.Plural["other"])
	}
	if len(items.Placeholders) != 1 {
		t.Errorf("expected 1 placeholder, got %d", len(items.Placeholders))
	}
	if items.Placeholders[0].Name != "count" {
		t.Errorf("expected placeholder 'count', got %q", items.Placeholders[0].Name)
	}
	if items.Placeholders[0].Type != parser.TypeInt {
		t.Errorf("expected type int, got %v", items.Placeholders[0].Type)
	}
}
