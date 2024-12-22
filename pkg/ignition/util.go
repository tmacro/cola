package ignition

func keysAreUnique[T any](col []T, getKey func(T) string) bool {
	seen := make(map[string]bool)
	for _, item := range col {
		key := getKey(item)
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}
