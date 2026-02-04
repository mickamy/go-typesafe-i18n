# go-typesafe-i18n

Type-safe internationalization library for Go. Generate type-safe functions from YAML locale files.

## Features

- Type-safe message functions generated from YAML
- Typed placeholders: `{name}` (string), `{count:int}` (int), `{price:float}` (float64)
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
go tool go-typesafe-i18n -pkg=messages -out=messages/messages_gen.go locales/en.yaml
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
go-typesafe-i18n -pkg=messages -out=messages/messages_gen.go locales/en.yaml
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

## CLI Options

```
Usage: go-typesafe-i18n [options] <yaml-file>

Options:
  -pkg string
        package name for generated code (default "messages")
  -out string
        output file path (default "messages_gen.go")
  -version
        show version
```

## Placeholder Types

| Syntax          | Go Type   | Example               |
|-----------------|-----------|-----------------------|
| `{name}`        | `string`  | `"Hello, {name}!"`    |
| `{count:int}`   | `int`     | `"{count:int} items"` |
| `{price:float}` | `float64` | `"${price:float}"`    |

## Escaping

Use `\{` and `\}` to render literal braces:

```yaml
example: "Use \{name\} for placeholders"
# Output: Use {name} for placeholders
```

## License

[MIT](./LICENSE)
