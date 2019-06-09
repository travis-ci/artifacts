package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestFormat(t *testing.T) {
	formatter := &MultiLineFormatter{}
	log := logrus.New()
	entry := &logrus.Entry{
		Logger:  log,
		Level:   logrus.InfoLevel,
		Message: "something",
		Data:    logrus.Fields{"foo": "bar"},
	}

	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Error(err)
	}

	expected := "INFO: something\n  foo: bar\n\n"
	actual := string(bytes)

	if expected != actual {
		t.Logf("%q != %q", expected, actual)
		t.Fail()
	}
}
