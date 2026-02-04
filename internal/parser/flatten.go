package parser

// flatten converts a nested map into messages and plurals.
func flatten(prefix string, m map[string]any) (map[string]string, map[string]map[string]string) {
	messages := make(map[string]string)
	plurals := make(map[string]map[string]string)

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case string:
			messages[key] = val
		case map[string]any:
			if isPluralMap(val) {
				plural := make(map[string]string)
				for form, tmpl := range val {
					// Safe to assert since isPluralMap checked the type
					plural[form] = tmpl.(string)
				}
				plurals[key] = plural
			} else {
				nestedMsgs, nestedPlurals := flatten(key, val)
				for k, v := range nestedMsgs {
					messages[k] = v
				}
				for k, v := range nestedPlurals {
					plurals[k] = v
				}
			}
		}
	}

	return messages, plurals
}
