package lint

import (
	"fmt"
	"os"
	"sort"

	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

// Result contains the lint result for a single file.
type Result struct {
	Path    string
	Missing []string
	Extra   []string
}

// HasIssues returns true if there are any missing or extra keys.
func (r Result) HasIssues() bool {
	return len(r.Missing) > 0 || len(r.Extra) > 0
}

// Check compares locale files against a base file and returns lint results.
func Check(basePath string, targetPaths []string) ([]Result, error) {
	baseKeys, err := loadKeys(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base file %s: %w", basePath, err)
	}

	var results []Result
	for _, targetPath := range targetPaths {
		if targetPath == basePath {
			continue
		}

		targetKeys, err := loadKeys(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load file %s: %w", targetPath, err)
		}

		result := compare(targetPath, baseKeys, targetKeys)
		results = append(results, result)
	}

	return results, nil
}

// loadKeys loads a locale file and returns a set of keys.
func loadKeys(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := parser.ParseFile(path, data)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(parsed.Messages))
	for key := range parsed.Messages {
		keys[key] = struct{}{}
	}

	return keys, nil
}

// compare compares base and target keys and returns a Result.
func compare(targetPath string, baseKeys, targetKeys map[string]struct{}) Result {
	var missing, extra []string

	for key := range baseKeys {
		if _, ok := targetKeys[key]; !ok {
			missing = append(missing, key)
		}
	}

	for key := range targetKeys {
		if _, ok := baseKeys[key]; !ok {
			extra = append(extra, key)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)

	return Result{
		Path:    targetPath,
		Missing: missing,
		Extra:   extra,
	}
}
