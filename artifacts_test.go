package main

import (
	"testing"
)

func TestBuildApp(t *testing.T) {
	app := buildApp()
	if app == nil {
		t.Errorf("app is nil")
	}

	if app.Name != "artifacts" {
		t.Errorf("unexpected app name: %v", app.Name)
	}
}
