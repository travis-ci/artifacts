package path

import "testing"

func TestNewPathSet(t *testing.T) {
	ps := NewPathSet()
	if len(ps.All()) > 0 {
		t.Fatalf("new PathSet has non-empty paths")
	}
}

func TestPathSetDeDuping(t *testing.T) {
	ps := NewPathSet()

	ps.Add(New("/", "foo", "bar"))
	ps.Add(New("/", "foo", "bar"))

	if len(ps.All()) > 1 {
		t.Fatalf("duplicate path was not de-duped")
	}
}
