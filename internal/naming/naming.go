package naming

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/mickamy/go-typesafe-i18n/internal/parser"
)

// KeyToFunctionName converts a message key to a PascalCase function name.
// e.g., "user.not_found" -> "UserNotFound"
func KeyToFunctionName(key string) string {
	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '.' || r == '_' || r == '-'
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}

	return result.String()
}

// DetectCollisions checks for function name collisions among messages.
// Returns an error describing the collisions if any are found.
func DetectCollisions(messages []parser.Message) error {
	seen := make(map[string][]string)

	for _, msg := range messages {
		funcName := KeyToFunctionName(msg.Key)
		seen[funcName] = append(seen[funcName], msg.Key)
	}

	var collisions []string
	for funcName, keys := range seen {
		if len(keys) > 1 {
			collisions = append(collisions, fmt.Sprintf("%s: %v", funcName, keys))
		}
	}

	if len(collisions) == 0 {
		return nil
	}

	return fmt.Errorf("function name collisions detected:\n  %s", strings.Join(collisions, "\n  "))
}
