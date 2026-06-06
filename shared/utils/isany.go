package utils

// IsAny check if a comparable value matches any of the provided valid values (generic helper for enum validation).
func IsAny[T comparable](v T, valid ...T) bool {
	for _, x := range valid {
		if x == v {
			return true
		}
	}
	return false
}
