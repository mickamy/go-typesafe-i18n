package main

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/mickamy/go-typesafe-i18n"
	"github.com/mickamy/go-typesafe-i18n/examples/basic/messages"
)

func main() {
	// Create a bundle with English as the default language
	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadFile("locales/en.yaml")
	bundle.MustLoadFile("locales/ja.yaml")

	// English localizer
	fmt.Println("=== English ===")
	en := bundle.Localizer(language.English)
	fmt.Println(en.Localize(messages.Greeting()))
	fmt.Println(en.Localize(messages.Hello("World")))
	fmt.Println(en.Localize(messages.ItemsCount(5)))
	fmt.Println(en.Localize(messages.TotalPrice(1234.56)))
	fmt.Println(en.Localize(messages.UserNotFound()))
	fmt.Println(en.Localize(messages.UserDeleted("Alice")))

	// Japanese localizer
	fmt.Println("\n=== Japanese ===")
	ja := bundle.Localizer(language.Japanese)
	fmt.Println(ja.Localize(messages.Greeting()))
	fmt.Println(ja.Localize(messages.Hello("太郎")))
	fmt.Println(ja.Localize(messages.ItemsCount(5)))
	fmt.Println(ja.Localize(messages.TotalPrice(1234.56)))
	fmt.Println(ja.Localize(messages.UserNotFound()))
	fmt.Println(ja.Localize(messages.UserDeleted("Alice")))
}
