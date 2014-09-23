package env

import (
	"os"
	"strconv"
	"strings"
)

// CascadeMatch is like Cascade, but also returns which env var
// was used to retrieve the value
func CascadeMatch(keys []string, dflt string) (string, string) {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		return value, key
	}

	return dflt, ""
}

// Slice returns a string slice from the env given a delimiter
func Slice(key, delim string, dflt []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
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

	return expandSlice(ret)
}

// Uint returns an uint from the env
func Uint(key string, dflt uint64) uint64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return dflt
	}

	uintVal, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return dflt
	}

	return uintVal
}

func expandSlice(vars []string) []string {
	expanded := []string{}
	for _, s := range vars {
		expanded = append(expanded, os.ExpandEnv(s))
	}
	return expanded
}
