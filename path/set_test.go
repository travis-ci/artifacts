package path

import "testing"

func TestNewSet(t *testing.T) {
	ps := NewSet()
	if len(ps.All()) > 0 {
		t.Fatalf("new Set has non-empty paths")
	}
}

func TestSetDeDuping(t *testing.T) {
	ps := NewSet()

	ps.Add(New("/", "foo", "bar"))
	ps.Add(New("/", "foo", "bar"))

	if len(ps.All()) > 1 {
		t.Fatalf("duplicate path was not de-duped")
	}
}
