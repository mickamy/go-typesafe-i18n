package naming_test

import (
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/naming"
	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

func TestKeyToFunctionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		expected string
	}{
		{"greeting", "Greeting"},
		{"user.not_found", "UserNotFound"},
		{"user.deleted", "UserDeleted"},
		{"items_count", "ItemsCount"},
		{"total-price", "TotalPrice"},
		{"a.b.c", "ABC"},
		{"user_name", "UserName"},
		{"already_PascalCase", "AlreadyPascalCase"},
		{"UPPER_CASE", "UPPERCASE"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			got := naming.KeyToFunctionName(tt.key)
			if got != tt.expected {
				t.Errorf("KeyToFunctionName(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestDetectCollisions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		messages []parser.Message
		wantErr  bool
	}{
		{
			name: "no collisions",
			messages: []parser.Message{
				{Key: "user.not_found"},
				{Key: "user.deleted"},
				{Key: "greeting"},
			},
			wantErr: false,
		},
		{
			name: "collision between dot and underscore",
			messages: []parser.Message{
				{Key: "user.name"},
				{Key: "user_name"},
			},
			wantErr: true,
		},
		{
			name: "collision between dot and hyphen",
			messages: []parser.Message{
				{Key: "user.name"},
				{Key: "user-name"},
			},
			wantErr: true,
		},
		{
			name: "multiple collisions",
			messages: []parser.Message{
				{Key: "user.name"},
				{Key: "user_name"},
				{Key: "item.count"},
				{Key: "item_count"},
			},
			wantErr: true,
		},
		{
			name:     "empty messages",
			messages: []parser.Message{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := naming.DetectCollisions(tt.messages)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDetectCollisions_ErrorMessage(t *testing.T) {
	t.Parallel()

	messages := []parser.Message{
		{Key: "user.name"},
		{Key: "user_name"},
	}

	err := naming.DetectCollisions(messages)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "UserName") {
		t.Errorf("error should contain function name 'UserName', got: %s", errStr)
	}
	if !contains(errStr, "user.name") {
		t.Errorf("error should contain key 'user.name', got: %s", errStr)
	}
	if !contains(errStr, "user_name") {
		t.Errorf("error should contain key 'user_name', got: %s", errStr)
	}

	t.Logf("collision error message:\n%s", errStr)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
