package normalize

import (
	"regexp"
	"strings"
)

var (
	normalizeRegexp        = regexp.MustCompile(`[^a-zA-Z0-9_/\.]+`)
	camelCaseMatchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	camelCaseMatchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// Key removes period and other invalid chars from the key
func Key(key string) string {
	key = strings.ToLower(key)
	key = normalizeRegexp.ReplaceAllString(key, "")
	key = strings.ReplaceAll(key, ".", "/") // replace all the periods (.)
	return key
}

// ToSnakeCase converts camelCase to snake_case
func ToSnakeCase(str string) string {
	snake := camelCaseMatchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = camelCaseMatchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
