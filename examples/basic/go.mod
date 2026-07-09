module github.com/mickamy/go-typesafe-i18n/examples/basic

go 1.26.4

tool github.com/mickamy/go-typesafe-i18n/cmd/go-typesafe-i18n

require (
	github.com/mickamy/go-typesafe-i18n v0.0.0
	golang.org/x/text v0.39.0
)

require (
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/mickamy/go-typesafe-i18n => ../..
