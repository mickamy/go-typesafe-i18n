package i18n

// Message represents a localizable message with its ID and arguments.
type Message struct {
	ID          string
	Args        map[string]any
	PluralCount *int
}
