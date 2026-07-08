package i18n

import "golang.org/x/text/language"

// Loaded reports whether the catalog of lang has been loaded.
func (b *Bundle) Loaded(lang language.Tag) bool {
	_, ok := b.catalogs[lang]
	return ok
}
