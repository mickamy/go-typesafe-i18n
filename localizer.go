package i18n

import (
	"fmt"
	"math"
	"strconv"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	"github.com/mickamy/go-typesafe-i18n/internal/locale"
	"github.com/mickamy/go-typesafe-i18n/internal/template"
)

// Localizer renders messages in a fixed language. Obtain one from
// Bundle.Localizer. The zero value renders every message as its key.
type Localizer struct {
	layers []layer
}

// Localizer returns a Localizer for the loaded locale best matching tag
// (e.g., en-US matches en), falling back to the default language when
// nothing matches.
func (b *Bundle) Localizer(tag language.Tag) Localizer {
	if b.matcher == nil {
		return Localizer{}
	}
	_, idx, _ := b.matcher.Match(tag)
	matched := b.tags[idx]
	layers := []layer{b.layer(matched)}
	if _, hasDefault := b.catalogs[b.defaultTag]; hasDefault && matched != b.defaultTag {
		layers = append(layers, b.layer(b.defaultTag))
	}
	return Localizer{layers: layers}
}

// Localize renders the message in the Localizer's language. It never fails:
// a message missing from the language falls back to the default language and
// finally to the message key itself, and a missing argument renders as the
// placeholder name in braces.
func (l Localizer) Localize(m Message) string {
	for _, ly := range l.layers {
		entry, ok := ly.entries[m.Key]
		if !ok {
			continue
		}
		return ly.render(entry, m)
	}
	return m.Key
}

// layer is one language in the fallback chain of a Localizer. Plural rules
// and number formatting follow the layer's own language, so a message that
// falls back to the default language is rendered entirely under the default
// language's conventions.
type layer struct {
	tag     language.Tag
	entries map[string]locale.Entry
	printer *message.Printer
}

func (b *Bundle) layer(tag language.Tag) layer {
	return layer{
		tag:     tag,
		entries: b.catalogs[tag].Entries,
		printer: b.printers[tag],
	}
}

func (ly layer) render(entry locale.Entry, m Message) string {
	tmpl := entry.Single
	if entry.Plural != nil {
		tmpl = ly.pluralVariant(entry, m)
	}
	return tmpl.Render(func(p template.Param) string {
		v, ok := lookupArg(m, p.Name)
		if !ok {
			return "{" + p.Name + "}"
		}
		return ly.format(v)
	})
}

// pluralVariant selects the plural form for the message's count argument.
// Without a usable count, and for CLDR forms the catalog does not provide,
// it falls back to "other".
func (ly layer) pluralVariant(entry locale.Entry, m Message) template.Template {
	form := "other"
	if v, ok := lookupArg(m, locale.CountParam); ok {
		if n, ok := asInt(v); ok {
			form = locale.FormName(pluralForm(ly.tag, n))
		}
	}
	tmpl, ok := entry.Plural[form]
	if !ok {
		tmpl = entry.Plural["other"]
	}
	return tmpl
}

// format renders an argument by its value type: strings verbatim, integers
// plain, and float64 with the layer's locale conventions (e.g., 1,234.56).
func (ly layer) format(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case float64:
		return ly.printer.Sprint(number.Decimal(v))
	}
	if n, ok := asInt(v); ok {
		return strconv.Itoa(n)
	}
	return fmt.Sprint(v)
}

// asInt converts any standard integer type to int for plural form matching.
// Generated code always passes int; this keeps hand-built Messages working.
func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		if n < math.MinInt || n > math.MaxInt {
			return 0, false
		}
		return int(n), true
	case uint:
		if uint64(n) > math.MaxInt {
			return 0, false
		}
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		if uint64(n) > math.MaxInt {
			return 0, false
		}
		return int(n), true
	case uint64:
		if n > math.MaxInt {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func lookupArg(m Message, name string) (any, bool) {
	for _, a := range m.Args {
		if a.Name == name {
			return a.Value, true
		}
	}
	return nil, false
}

func pluralForm(tag language.Tag, n int) plural.Form {
	if n == math.MinInt {
		n++ // |math.MinInt| overflows int; the plural form is unaffected at this magnitude
	}
	if n < 0 {
		n = -n
	}
	return plural.Cardinal.MatchPlural(tag, n, 0, 0, 0, 0)
}
