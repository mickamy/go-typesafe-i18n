package parser

// flatten converts a nested map into a flat map with dot-separated keys.
func flatten(prefix string, m map[string]any, result map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case string:
			result[key] = val
		case map[string]any:
			flatten(key, val, result)
		}
	}
}
