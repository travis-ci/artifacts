package client

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
)

func TestNew(t *testing.T) {
	log := logrus.New()
	log.Level = logrus.PanicLevel
	c := New("host.example.com", "foo-bar", log)

	if c.SaveHost != "host.example.com" {
		t.Fatalf("client save host does not match")
	}

	if c.Token != "foo-bar" {
		t.Fatalf("client token does not match")
	}

	if c.RetryInterval != (3 * time.Second) {
		t.Fatalf("default retry interval does not match")
	}
}
