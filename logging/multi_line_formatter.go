package logging

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
)

// MultiLineFormatter is a logrus-compatible formatter for multi-line output
type MultiLineFormatter struct{}

// Format creates a formatted entry
func (f *MultiLineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var serialized []byte

	levelText := strings.ToUpper(entry.Data["level"].(string))
	serialized = append(serialized, []byte(fmt.Sprintf("%s: %s\n", levelText, entry.Data["msg"]))...)

	keys := make([]string, 0)
	for k := range entry.Data {
		if k != "level" && k != "time" && k != "msg" {
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
