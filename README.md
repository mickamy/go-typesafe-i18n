# go-typesafe-i18n

Type-safe internationalization library for Go. Generate type-safe functions from locale files.

## Features

- Supports YAML, JSON, and TOML formats
- Type-safe message functions generated at build time
- Typed placeholders: `{name}` (string), `{count:int}` (int), `{price:float}` (float64)
- Plural support using CLDR rules
- Language fallback using BCP 47 tags
- Escape support: `\{name\}` renders as literal `{name}`
- Collision detection at generation time

## Installation

### Runtime library

```bash
go get github.com/mickamy/go-typesafe-i18n
```

### Code generator

Global install:

```bash
go install github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n@latest
```

Or as a project-local tool (Go 1.24+):

```bash
go get -tool github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n@latest
go tool go-typesafe-i18n generate -base=en -pkg=messages -out=messages/messages_gen.go locales/
```

## Quick Start

### 1. Create locale files

```yaml
# locales/en.yaml
greeting: Hello
hello: "Hello, {name}!"
items_count: "{count:int} items"
total_price: "Total: ${price:float}"
user:
  not_found: User not found
  deleted: "User {name} has been deleted"
```

```yaml
# locales/ja.yaml
greeting: こんにちは
hello: "こんにちは、{name}さん！"
items_count: "{count:int}件のアイテム"
total_price: "合計: ¥{price:float}"
user:
  not_found: ユーザーが見つかりません
  deleted: "ユーザー {name} を削除しました"
```

### 2. Generate code

```bash
go-typesafe-i18n generate -base=en -pkg=messages -out=messages/messages_gen.go locales/
```

### 3. Use in your code

```go
package main

import (
    "fmt"

    "golang.org/x/text/language"

    "github.com/mickamy/go-typesafe-i18n"
    
	"your-module/messages"
)

func main() {
    bundle := i18n.NewBundle(language.English)
    bundle.MustLoadFile("locales/en.yaml")
    bundle.MustLoadFile("locales/ja.yaml")

    en := bundle.Localizer(language.English)
    fmt.Println(en.Localize(messages.Hello("World")))
    // Output: Hello, World!

    ja := bundle.Localizer(language.Japanese)
    fmt.Println(ja.Localize(messages.Hello("太郎")))
    // Output: こんにちは、太郎さん！
}
```

## CLI

```
Usage: go-typesafe-i18n <command> [options]

Commands:
  generate    Generate type-safe message functions from a locale directory
  lint        Check key consistency across locale files
```

### generate

```
Usage: go-typesafe-i18n generate -base=<locale> [options] <locale-dir>

Options:
  -base string
        base locale name (required)
  -pkg string
        package name for generated code (default "messages")
  -out string
        output file path (default "messages_gen.go")
```

### lint

Check that all locale files have the same keys as the base locale.

```
Usage: go-typesafe-i18n lint -base=<locale> <directory>
       go-typesafe-i18n lint -base=<file> <target-files...>

Options:
  -base string
        base locale name or file path (required)
```

Examples:

```bash
# Check all files in a directory against "en" locale
go-typesafe-i18n lint -base=en locales/

# Check specific files
go-typesafe-i18n lint -base=locales/en.yaml locales/ja.yaml locales/fr.yaml
```

## Placeholder Types

| Syntax          | Go Type   | Example               |
|-----------------|-----------|-----------------------|
| `{name}`        | `string`  | `"Hello, {name}!"`    |
| `{count:int}`   | `int`     | `"{count:int} items"` |
| `{price:float}` | `float64` | `"${price:float}"`    |

## Plural Support

Define plural forms using CLDR categories (`zero`, `one`, `two`, `few`, `many`, `other`):

```yaml
# locales/en.yaml
items_count:
  one: "You have 1 item"
  other: "You have {count:int} items"
```

```yaml
# locales/ja.yaml
items_count:
  one: "1件のアイテム"
  other: "{count:int}件のアイテム"
```

The generated function takes the count parameter and automatically selects the correct plural form:

```go
en := bundle.Localizer(language.English)
fmt.Println(en.Localize(messages.ItemsCount(1)))  // "You have 1 item"
fmt.Println(en.Localize(messages.ItemsCount(5)))  // "You have 5 items"

ja := bundle.Localizer(language.Japanese)
fmt.Println(ja.Localize(messages.ItemsCount(1)))  // "1件のアイテム"
fmt.Println(ja.Localize(messages.ItemsCount(5)))  // "5件のアイテム"
```

The first `int` typed placeholder is used to determine the plural form.

## Escaping

Use `\{` and `\}` to render literal braces:

```yaml
example: "Use \{name\} for placeholders"
# Output: Use {name} for placeholders
```

## License

[MIT](./LICENSE)
