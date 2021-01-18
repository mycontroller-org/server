package normalize

import (
	"regexp"
	"strings"
)

var normalizeRegexp = regexp.MustCompile(`[^a-zA-Z0-9_/\.]+`)

// Key removes period and other invalid chars from the key
func Key(key string) string {
	key = strings.ToLower(key)
	key = normalizeRegexp.ReplaceAllString(key, "")
	key = strings.ReplaceAll(key, ".", "/") // replace all the periods (.)
	return key
}
