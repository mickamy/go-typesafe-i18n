package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickamy/go-typesafe-i18n/internal/lint"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		base        string
		targets     map[string]string
		wantResults []struct {
			path    string
			missing []string
			extra   []string
		}
	}{
		{
			name: "all keys match",
			base: `
greeting: Hello
goodbye: Goodbye
`,
			targets: map[string]string{
				"ja.yaml": `
greeting: こんにちは
goodbye: さようなら
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "ja.yaml", missing: nil, extra: nil},
			},
		},
		{
			name: "missing keys",
			base: `
greeting: Hello
goodbye: Goodbye
welcome: Welcome
`,
			targets: map[string]string{
				"ja.yaml": `
greeting: こんにちは
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "ja.yaml", missing: []string{"goodbye", "welcome"}, extra: nil},
			},
		},
		{
			name: "extra keys",
			base: `
greeting: Hello
`,
			targets: map[string]string{
				"ja.yaml": `
greeting: こんにちは
extra_key: 余分なキー
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "ja.yaml", missing: nil, extra: []string{"extra_key"}},
			},
		},
		{
			name: "missing and extra keys",
			base: `
greeting: Hello
goodbye: Goodbye
`,
			targets: map[string]string{
				"ja.yaml": `
greeting: こんにちは
extra_key: 余分なキー
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "ja.yaml", missing: []string{"goodbye"}, extra: []string{"extra_key"}},
			},
		},
		{
			name: "multiple targets",
			base: `
greeting: Hello
goodbye: Goodbye
`,
			targets: map[string]string{
				"ja.yaml": `
greeting: こんにちは
`,
				"fr.yaml": `
greeting: Bonjour
goodbye: Au revoir
extra: Supplémentaire
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "fr.yaml", missing: nil, extra: []string{"extra"}},
				{path: "ja.yaml", missing: []string{"goodbye"}, extra: nil},
			},
		},
		{
			name: "nested keys",
			base: `
user:
  greeting: Hello
  goodbye: Goodbye
`,
			targets: map[string]string{
				"ja.yaml": `
user:
  greeting: こんにちは
`,
			},
			wantResults: []struct {
				path    string
				missing []string
				extra   []string
			}{
				{path: "ja.yaml", missing: []string{"user.goodbye"}, extra: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			basePath := filepath.Join(tmpDir, "en.yaml")
			if err := os.WriteFile(basePath, []byte(tt.base), 0o644); err != nil {
				t.Fatalf("failed to write base file: %v", err)
			}

			var targetPaths []string
			for name, content := range tt.targets {
				path := filepath.Join(tmpDir, name)
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("failed to write target file: %v", err)
				}
				targetPaths = append(targetPaths, path)
			}

			results, err := lint.Check(basePath, targetPaths)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != len(tt.wantResults) {
				t.Fatalf("expected %d results, got %d", len(tt.wantResults), len(results))
			}

			// Build a map for easier comparison (order may vary)
			resultMap := make(map[string]lint.Result)
			for _, r := range results {
				resultMap[filepath.Base(r.Path)] = r
			}

			for _, want := range tt.wantResults {
				got, ok := resultMap[want.path]
				if !ok {
					t.Errorf("missing result for %s", want.path)
					continue
				}

				if !stringSliceEqual(got.Missing, want.missing) {
					t.Errorf("%s missing: got %v, want %v", want.path, got.Missing, want.missing)
				}

				if !stringSliceEqual(got.Extra, want.extra) {
					t.Errorf("%s extra: got %v, want %v", want.path, got.Extra, want.extra)
				}

				t.Logf("%s passed: missing %v, extra %v", want.path, got.Missing, got.Extra)
			}
		})
	}
}

func TestResult_HasIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result lint.Result
		want   bool
	}{
		{
			name:   "no issues",
			result: lint.Result{Path: "test.yaml", Missing: nil, Extra: nil},
			want:   false,
		},
		{
			name:   "has missing",
			result: lint.Result{Path: "test.yaml", Missing: []string{"key"}, Extra: nil},
			want:   true,
		},
		{
			name:   "has extra",
			result: lint.Result{Path: "test.yaml", Missing: nil, Extra: []string{"key"}},
			want:   true,
		},
		{
			name:   "has both",
			result: lint.Result{Path: "test.yaml", Missing: []string{"a"}, Extra: []string{"b"}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.result.HasIssues(); got != tt.want {
				t.Errorf("HasIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
