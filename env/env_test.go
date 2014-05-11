package env

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("FOO", "1")
	os.Setenv("BAR", "")
	os.Setenv("BAZ", "a;b;c;;")
}

func TestGet(t *testing.T) {
	if Get("FOO", "3") != "1" {
		t.Fail()
	}

	if Get("BAR", "2") != "2" {
		t.Fail()
	}
}

func TestBool(t *testing.T) {
	if Bool("FOO", false) != true {
		t.Fail()
	}

	if Bool("BAR", false) != false {
		t.Fail()
	}
}

func TestSlice(t *testing.T) {
	s := Slice("BAZ", ";", []string{})
	if len(s) != 3 {
		t.Fail()
	}

	if s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Fail()
	}
}

func TestInt(t *testing.T) {
	if Int("FOO", 4) != 1 {
		t.Fail()
	}

	if Int("BAR", 3) != 3 {
		t.Fail()
	}
}
