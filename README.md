# go-typesafe-i18n

[![CI](https://github.com/mickamy/go-typesafe-i18n/actions/workflows/ci.yaml/badge.svg)](https://github.com/mickamy/go-typesafe-i18n/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mickamy/go-typesafe-i18n.svg)](https://pkg.go.dev/github.com/mickamy/go-typesafe-i18n)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Type-safe internationalization for Go: locale files in, compile-time-checked message constructors out.

```go
fmt.Println(en.Localize(messages.Hello("World")))  // Hello, World!
fmt.Println(ja.Localize(messages.ItemsCount(5)))   // アイテムが5個あります。
```

Message keys and parameter types are checked by the compiler: renaming a key, dropping a parameter, or changing its type breaks the build instead of shipping a broken message.

## How it works

1. Write locale files in YAML or TOML.
2. `go generate` runs `cmd/go-typesafe-i18n`, which validates every locale and generates a typed constructor per message from the default language.
3. At runtime, a `Bundle` loads the locale files and hands out `Localizer`s that render messages with CLDR plural rules, locale-aware number formatting, and language fallback.

## Quick start

### 1. Write locale files

```yaml
# locales/en.yaml
greeting: "Hello!"
hello: "Hello, {name}!"
items_count:
  one: "You have {count} item."
  other: "You have {count} items."
total_price: "Total: {price:number}"
user:
  not_found: "User not found."
  deleted: "User {name} has been deleted."
```

```yaml
# locales/ja.yaml
greeting: "こんにちは！"
hello: "こんにちは、{name}さん！"
items_count:
  other: "アイテムが{count}個あります。"
total_price: "合計: {price:number}"
user:
  not_found: "ユーザーが見つかりません。"
  deleted: "ユーザー{name}を削除しました。"
```

### 2. Generate typed constructors

Add the generator to your module as a tool (Go 1.24+) and declare a `go:generate` directive:

```sh
go get -tool github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n
```

```go
//go:generate go tool go-typesafe-i18n -out messages/messages.gen.go
package main
```

```sh
go generate ./...
```

This generates functions like `messages.Hello(name string)` and `messages.ItemsCount(count int)`, and fails when locales disagree with the default language (unknown keys, mismatched parameters, plural shape differences). Keys missing from a translation are warnings, because the runtime falls back to the default language.

### 3. Localize

```go
package main

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n"
	"example.com/yourmodule/messages"
)

func main() {
	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile("locales/en.yaml")
	bundle.MustLoadFile("locales/ja.yaml")

	en := bundle.Localizer(language.English)
	fmt.Println(en.Localize(messages.Hello("World")))      // Hello, World!
	fmt.Println(en.Localize(messages.ItemsCount(1)))       // You have 1 item.
	fmt.Println(en.Localize(messages.TotalPrice(1234.56))) // Total: 1,234.56

	ja := bundle.Localizer(language.Japanese)
	fmt.Println(ja.Localize(messages.Hello("太郎")))        // こんにちは、太郎さん！
	fmt.Println(ja.Localize(messages.ItemsCount(1)))       // アイテムが1個あります。
}
```

See [examples/basic](examples/basic) for the full program.

## Locale file format

- Nested mappings flatten into dot-joined keys: `user.not_found` generates `UserNotFound()`.
- A mapping whose keys are all CLDR plural categories (`zero`/`one`/`two`/`few`/`many`/`other`) is a plural group and must define `other`. Plural messages always take `count int` as their first parameter, and the form is selected by the CLDR rules of the rendering language (Japanese only ever uses `other`; English distinguishes `one`).
- Placeholders are `{name}` or `{name:kind}`, where kind is `string` (default), `int` (rendered plain), or `number` (`float64`, rendered with locale conventions: `1,234.56` in en, `1.234,56` in de).
- A bare placeholder inherits an explicit kind annotated elsewhere in the same message, including other plural variants; only conflicting explicit annotations are an error.
- Literal braces are escaped as `{{` and `}}`.
- TOML works the same way, chosen by file extension:

```toml
# locales/en.toml
greeting = "Hello!"

[items_count]
one = "You have {count} item."
other = "You have {count} items."
```

The language of a file comes from its name: `en.yaml`, `ja.toml`, `zh-Hans.yaml`. Data from other sources (such as `embed.FS`) loads via `bundle.LoadYAML(lang, data)` / `bundle.LoadTOML(lang, data)`.

## Runtime behavior

- `bundle.Localizer(tag)` picks the best loaded locale via `golang.org/x/text/language.NewMatcher`, so `en-US` resolves to `en`.
- `Localize` never fails. A message missing from the matched locale falls back through the loaded parents of its language tag (`en-GB` falls back to `en`), then the default language, and finally renders the key itself. A missing argument renders as the placeholder name in braces.
- Localizers are cheap value types, safe for concurrent use once loading is done: load every locale first, then hand out Localizers.

## CLI

```
go-typesafe-i18n [flags]
  -dir string      directory containing locale files (default "locales")
  -default string  default language defining the generated signatures (default "en")
  -out string      output file path (default "messages/messages.gen.go")
  -pkg string      package name of the generated file (default: base name of the output directory)
  -check           validate the locales without writing the output file
```

`-check` suits CI: it runs the full validation and reports cross-locale problems without touching the working tree.

## License

[MIT](LICENSE)
