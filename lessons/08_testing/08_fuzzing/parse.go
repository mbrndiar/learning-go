// Package fuzzing implements a tiny "key=value" parser used to demonstrate
// Go's built-in fuzzing support: seed corpus via f.Add, fuzz targets with
// f.Fuzz, and regression seeds stored under testdata/fuzz.
package fuzzing

import (
	"fmt"
	"strings"
)

// ParseKeyValue splits s on the first "=" and trims surrounding whitespace
// from both sides. It returns an error if s has no "=" or if the key is
// empty after trimming.
func ParseKeyValue(s string) (key, value string, err error) {
	rawKey, rawValue, found := strings.Cut(s, "=")
	if !found {
		return "", "", fmt.Errorf("parse key=value: no %q separator in %q", "=", s)
	}

	key = strings.TrimSpace(rawKey)
	value = strings.TrimSpace(rawValue)
	if key == "" {
		return "", "", fmt.Errorf("parse key=value: empty key in %q", s)
	}
	return key, value, nil
}
