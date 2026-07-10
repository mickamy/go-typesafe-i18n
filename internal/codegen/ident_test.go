package codegen_test

import (
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

func TestFuncName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want string
	}{
		{key: "greeting", want: "Greeting"},
		{key: "items_count", want: "ItemsCount"},
		{key: "user.not_found", want: "UserNotFound"},
		{key: "a.b_c", want: "ABC"},
		{key: "step_2", want: "Step2"},
	}
	for _, tt := range tests {
		if got := codegen.FuncName(tt.key); got != tt.want {
			t.Errorf("FuncName(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestParamName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "name", want: "name"},
		{name: "user_name", want: "userName"},
		{name: "a_b_c", want: "aBC"},
		{name: "_x", want: "x"},
	}
	for _, tt := range tests {
		got, err := codegen.ParamName(tt.name)
		if err != nil {
			t.Errorf("ParamName(%q) returned error: %v", tt.name, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParamName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}

	for _, name := range []string{"range", "type", "func", "_", "__", "_2x"} {
		if _, err := codegen.ParamName(name); err == nil {
			t.Errorf("ParamName(%q) returned nil error", name)
		}
	}
}
