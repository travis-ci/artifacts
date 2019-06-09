package client

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	log := logrus.New()
	log.Level = logrus.PanicLevel
	c := New("host.example.com", "foo-bar", log)

	if c.SaveHost != "host.example.com" {
		t.Fatalf("SaveHost %v != host.example.com", c.SaveHost)
	}

	if c.Token != "foo-bar" {
		t.Fatalf("Token %v != foo-bar", c.Token)
	}

	if c.RetryInterval != defaultRetryInterval {
		t.Fatalf("RetryInterval %v != %v", c.RetryInterval, defaultRetryInterval)
	}
}
