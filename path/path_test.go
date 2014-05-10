package path

import (
	"testing"
)

func TestNewPath(t *testing.T) {
	p := NewPath("/xyz", "foo", "bar")

	if p.Root != "/xyz" {
		t.Fail()
	}

	if p.From != "foo" {
		t.Fail()
	}

	if p.To != "bar" {
		t.Fail()
	}
}

func TestPathFullpath(t *testing.T) {
	if NewPath("/abc", "ham", "bone").Fullpath() != "/abc/ham" {
		t.Fail()
	}

	if NewPath("/nope", "/flim", "flam").Fullpath() != "/flim" {
		t.Fail()
	}
}
