package parser

// flatten converts a nested map into Result with messages and plurals.
func flatten(prefix string, m map[string]any, messages map[string]string, plurals map[string]map[string]string) {
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
					plural[form] = tmpl.(string)
				}
				plurals[key] = plural
			} else {
				flatten(key, val, messages, plurals)
			}
		}
	}
}
