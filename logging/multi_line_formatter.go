package logging

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// MultiLineFormatter is a logrus-compatible formatter for multi-line output
type MultiLineFormatter struct{}

// Format creates a formatted entry
func (f *MultiLineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var serialized []byte

	levelText := strings.ToUpper(entry.Level.String())

	msg := fmt.Sprintf("%s: %s\n", levelText, entry.Message)
	if levelText == "ERROR" {
		msg = fmt.Sprintf("\033[31;1m%s\033[0m", msg)
	}
	serialized = append(serialized, []byte(msg)...)

	keys := []string{}
	for k := range entry.Data {
		if k != "time" {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := entry.Data[k]
		serialized = append(serialized, []byte(fmt.Sprintf("  %v: %v\n", k, v))...)
	}

	return append(serialized, '\n'), nil
}
