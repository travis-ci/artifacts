package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

// Get returns a string from the env
func Get(key, dflt string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return dflt
	}

	return value
}

// Cascade is like Get, but with a bunch of tries
func Cascade(keys []string, dflt string) string {
	value, _ := CascadeMatch(keys, dflt)
	return value
}

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

// Bool returns a bool from the env
func Bool(key string, dflt bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return dflt
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return dflt
	}

	return boolVal
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

	return ExpandSlice(ret)
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

// UintSize returns a size-like uint from the env
func UintSize(key string, dflt uint64) uint64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return dflt
	}

	if strings.ContainsAny(value, "BKMGTPEZYbkmgtpezy") {
		b, err := humanize.ParseBytes(value)
		if err == nil {
			return b
		}
	}

	uintVal, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return dflt
	}

	return uintVal
}

// ExpandSlice evaluates each string in the slice through os.ExpandEnv
func ExpandSlice(vars []string) []string {
	expanded := []string{}
	for _, s := range vars {
		expanded = append(expanded, os.ExpandEnv(s))
	}
	return expanded
}
