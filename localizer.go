package i18n

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

// Localizer retrieves messages for a specific language.
type Localizer struct {
	bundle *Bundle
	lang   language.Tag
}

// Escape markers for placeholder processing.
const (
	escapedOpenBrace  = "\x00OPEN\x00"
	escapedCloseBrace = "\x00CLOSE\x00"
)

// Localize returns the localized message for the given Message.
// It uses language matching to find the best available translation.
// If the message is not found in any language, it panics.
// Escaped braces (\{ and \}) are converted to literal { and }.
func (l *Localizer) Localize(msg Message) string {
	tmpl := l.getTemplate(msg.ID)

	// Temporarily replace escaped braces
	result := strings.ReplaceAll(tmpl, `\{`, escapedOpenBrace)
	result = strings.ReplaceAll(result, `\}`, escapedCloseBrace)

	// Replace placeholders
	if msg.Args != nil {
		for name, value := range msg.Args {
			// Replace typed placeholders like {name:int} as well as {name}
			for _, suffix := range []string{":int", ":float", ":string", ""} {
				placeholder := "{" + name + suffix + "}"
				result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
			}
		}
	}

	// Restore escaped braces as literal braces
	result = strings.ReplaceAll(result, escapedOpenBrace, "{")
	result = strings.ReplaceAll(result, escapedCloseBrace, "}")

	return result
}

// getTemplate retrieves the template for the given message ID.
// It uses the language matcher to find the best available language,
// then falls back to the default language.
func (l *Localizer) getTemplate(id string) string {
	// Use matcher to find best matching language
	if l.bundle.matcher != nil {
		matched, _, _ := l.bundle.matcher.Match(l.lang)
		if msgs, ok := l.bundle.messages[matched]; ok {
			if tmpl, ok := msgs[id]; ok {
				return tmpl
			}
		}
	}

	// Fallback to default language
	if msgs, ok := l.bundle.messages[l.bundle.defaultLang]; ok {
		if tmpl, ok := msgs[id]; ok {
			return tmpl
		}
	}

	panic(fmt.Sprintf("go-typesafe-i18n: message not found: %s", id))
}
