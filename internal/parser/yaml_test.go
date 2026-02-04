package parser

import (
	"reflect"
	"testing"
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

			p := &YAMLParser{}
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
		expected []Message
		wantErr  bool
	}{
		{
			name:  "no placeholders",
			input: `greeting: "Hello"`,
			expected: []Message{
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
			expected: []Message{
				{
					Key:      "hello",
					Template: "Hello, {name}!",
					Placeholders: []Placeholder{
						{Name: "name", Type: TypeString},
					},
				},
			},
		},
		{
			name:  "int placeholder",
			input: `count: "{count:int} items"`,
			expected: []Message{
				{
					Key:      "count",
					Template: "{count:int} items",
					Placeholders: []Placeholder{
						{Name: "count", Type: TypeInt},
					},
				},
			},
		},
		{
			name:  "float placeholder",
			input: `price: "Total: ${price:float}"`,
			expected: []Message{
				{
					Key:      "price",
					Template: "Total: ${price:float}",
					Placeholders: []Placeholder{
						{Name: "price", Type: TypeFloat64},
					},
				},
			},
		},
		{
			name:  "multiple placeholders",
			input: `transfer: "Transfer {amount:int} from {from} to {to}"`,
			expected: []Message{
				{
					Key:      "transfer",
					Template: "Transfer {amount:int} from {from} to {to}",
					Placeholders: []Placeholder{
						{Name: "amount", Type: TypeInt},
						{Name: "from", Type: TypeString},
						{Name: "to", Type: TypeString},
					},
				},
			},
		},
		{
			name:  "duplicate placeholder uses first occurrence",
			input: `repeat: "{name} and {name} again"`,
			expected: []Message{
				{
					Key:      "repeat",
					Template: "{name} and {name} again",
					Placeholders: []Placeholder{
						{Name: "name", Type: TypeString},
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
			expected: []Message{
				{
					Key:      "user.greeting",
					Template: "Hello, {name}!",
					Placeholders: []Placeholder{
						{Name: "name", Type: TypeString},
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

			p := &YAMLParser{}
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
