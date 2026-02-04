package parser

import (
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
