package env

import (
	"os"
	"strconv"
	"strings"
)

// Get returns a string from the env
func Get(key, dflt string) string {
	value := os.Getenv(key)
	if value == "" {
		return dflt
	}

	return value
}

// Bool returns a bool from the env
func Bool(key string, dflt bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return dflt
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return dflt
	}

	return boolVal
}

// Getslice returns a string slice from the env given a delimiter
func Getslice(key, delim string, dflt []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return dflt
	}

	ret := []string{}
	for _, part := range strings.Split(value, delim) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			ret = append(ret, trimmed)
		}
	}

	return ret
}
