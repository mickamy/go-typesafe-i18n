package i18n_test

import (
	"testing"

	"golang.org/x/text/language"

	i18n "github.com/mickamy/go-typesafe-i18n"
)

const enYAML = `
greeting: "Hello!"
hello: "Hello, {name}!"
items_count:
  one: "You have {count} item."
  other: "You have {count} items."
total_price: "Total: {price:number}"
user:
  not_found: "User not found."
  deleted: "User {name} has been deleted."
`

// ja.user.deleted is intentionally missing to exercise the fallback chain,
// and items_count has only "other" because Japanese has no "one" form.
const jaTOML = `
greeting = "こんにちは！"
hello = "こんにちは、{name}さん！"
total_price = "合計: {price:number}"

[items_count]
other = "アイテムが{count}個あります。"

[user]
not_found = "ユーザーが見つかりません。"
`

func newBundle(t *testing.T) *i18n.Bundle {
	t.Helper()
	b := i18n.NewBundle(language.English)
	if err := b.LoadYAML(language.English, []byte(enYAML)); err != nil {
		t.Fatal(err)
	}
	if err := b.LoadTOML(language.Japanese, []byte(jaTOML)); err != nil {
		t.Fatal(err)
	}
	return b
}

func msg(key string, args ...i18n.Arg) i18n.Message {
	return i18n.Message{Key: key, Args: args}
}

func TestLocalizer_Localize(t *testing.T) {
	t.Parallel()

	b := newBundle(t)
	en := b.Localizer(language.English)
	ja := b.Localizer(language.Japanese)

	tests := []struct {
		name string
		loc  i18n.Localizer
		msg  i18n.Message
		want string
	}{
		{name: "en static", loc: en, msg: msg("greeting"), want: "Hello!"},
		{
			name: "en string arg",
			loc:  en,
			msg:  msg("hello", i18n.Arg{Name: "name", Value: "World"}),
			want: "Hello, World!",
		},
		{
			name: "en plural one",
			loc:  en,
			msg:  msg("items_count", i18n.Arg{Name: "count", Value: 1}),
			want: "You have 1 item.",
		},
		{
			name: "en plural other",
			loc:  en,
			msg:  msg("items_count", i18n.Arg{Name: "count", Value: 5}),
			want: "You have 5 items.",
		},
		{
			name: "en plural zero uses other",
			loc:  en,
			msg:  msg("items_count", i18n.Arg{Name: "count", Value: 0}),
			want: "You have 0 items.",
		},
		{
			name: "en plural negative count",
			loc:  en,
			msg:  msg("items_count", i18n.Arg{Name: "count", Value: -1}),
			want: "You have -1 item.",
		},
		{
			name: "en number formatting",
			loc:  en,
			msg:  msg("total_price", i18n.Arg{Name: "price", Value: 1234.56}),
			want: "Total: 1,234.56",
		},
		{name: "en nested key", loc: en, msg: msg("user.not_found"), want: "User not found."},
		{name: "ja static", loc: ja, msg: msg("greeting"), want: "こんにちは！"},
		{
			name: "ja string arg",
			loc:  ja,
			msg:  msg("hello", i18n.Arg{Name: "name", Value: "太郎"}),
			want: "こんにちは、太郎さん！",
		},
		{
			name: "ja plural count 1 uses other",
			loc:  ja,
			msg:  msg("items_count", i18n.Arg{Name: "count", Value: 1}),
			want: "アイテムが1個あります。",
		},
		{
			name: "ja number formatting",
			loc:  ja,
			msg:  msg("total_price", i18n.Arg{Name: "price", Value: 1234.56}),
			want: "合計: 1,234.56",
		},
		{
			name: "ja missing key falls back to default language",
			loc:  ja,
			msg:  msg("user.deleted", i18n.Arg{Name: "name", Value: "Alice"}),
			want: "User Alice has been deleted.",
		},
		{name: "unknown key returns the key", loc: en, msg: msg("nope.key"), want: "nope.key"},
		{name: "missing arg renders placeholder", loc: en, msg: msg("hello"), want: "Hello, {name}!"},
		{
			name: "unexpected arg type falls back to fmt",
			loc:  en,
			msg:  msg("hello", i18n.Arg{Name: "name", Value: 42}),
			want: "Hello, 42!",
		},
		{
			name: "plural without count uses other",
			loc:  en,
			msg:  msg("items_count"),
			want: "You have {count} items.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.loc.Localize(tt.msg); got != tt.want {
				t.Errorf("Localize(%q) = %q, want %q", tt.msg.Key, got, tt.want)
			}
		})
	}
}

func TestBundle_Localizer_matching(t *testing.T) {
	t.Parallel()

	b := newBundle(t)

	tests := []struct {
		name string
		tag  language.Tag
		want string
	}{
		{name: "exact match", tag: language.Japanese, want: "こんにちは！"},
		{name: "region narrows to base language", tag: language.MustParse("ja-JP"), want: "こんにちは！"},
		{name: "en-US matches en", tag: language.AmericanEnglish, want: "Hello!"},
		{name: "unloaded language falls back to default", tag: language.French, want: "Hello!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := b.Localizer(tt.tag).Localize(msg("greeting")); got != tt.want {
				t.Errorf("Localizer(%v).Localize(greeting) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestBundle_Localizer_germanNumberFormat(t *testing.T) {
	t.Parallel()

	b := i18n.NewBundle(language.English)
	if err := b.LoadYAML(language.German, []byte("total_price: \"Gesamt: {price:number}\"\n")); err != nil {
		t.Fatal(err)
	}
	de := b.Localizer(language.German)
	got := de.Localize(msg("total_price", i18n.Arg{Name: "price", Value: 1234.56}))
	if want := "Gesamt: 1.234,56"; got != want {
		t.Errorf("Localize(total_price) = %q, want %q", got, want)
	}
}

func TestBundle_Localizer_withoutDefaultCatalog(t *testing.T) {
	t.Parallel()

	b := i18n.NewBundle(language.English) // English itself is never loaded
	if err := b.LoadTOML(language.Japanese, []byte("greeting = \"こんにちは！\"\n")); err != nil {
		t.Fatal(err)
	}
	if got := b.Localizer(language.French).Localize(msg("greeting")); got != "こんにちは！" {
		t.Errorf("Localize(greeting) = %q, want こんにちは！", got)
	}
}

func TestLocalizer_zeroValue(t *testing.T) {
	t.Parallel()

	var zero i18n.Localizer
	if got := zero.Localize(msg("greeting")); got != "greeting" {
		t.Errorf("Localize(greeting) = %q, want %q", got, "greeting")
	}
	if got := i18n.NewBundle(language.English).Localizer(language.English).Localize(msg("x")); got != "x" {
		t.Errorf("empty bundle Localize(x) = %q, want %q", got, "x")
	}
}
