// Command go-typesafe-i18n generates typed message constructors from locale
// files.
//
// Add it to your module as a tool (Go 1.24+):
//
//	go get -tool github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n
//
// and invoke it via go:generate:
//
//	//go:generate go tool go-typesafe-i18n -out messages/messages.gen.go
//
// Alternatively, without the tool directive:
//
//	//go:generate go run github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n -out messages/messages.gen.go
//
// By default it reads locale files from the "locales" directory with "en"
// as the default language; see -dir, -default, -pkg, and -check.
package main

import (
	"fmt"
	"os"

	"github.com/mickamy/go-typesafe-i18n/internal/codegen"
)

func main() {
	if err := codegen.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "go-typesafe-i18n:", err)
		os.Exit(1)
	}
}
