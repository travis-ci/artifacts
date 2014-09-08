package env

import (
	"os"
	"reflect"
	"testing"
)

func init() {
	os.Setenv("FOO", "1")
	os.Setenv("BAR", "")
	os.Setenv("BAZ", "a:b:c::")
	os.Setenv("MOAR", "32GB")
}

type sliceCase struct {
	expected []string
	actual   []string
}

func TestGet(t *testing.T) {
	for actual, expected := range map[string]string{
		Get("FOO", "3"): "1",
		Get("BAR", "2"): "2",
	} {
		if expected != actual {
			t.Fatalf("%v != %v", expected, actual)
		}
	}
}

func TestCascade(t *testing.T) {
	for actual, expected := range map[string]string{
		Cascade([]string{"DERP", "FOO"}, "2"):  "1",
		Cascade([]string{"FOO", "BAZ"}, "2"):   "1",
		Cascade([]string{"DERP", "NERP"}, "3"): "3",
	} {
		if expected != actual {
			t.Fatalf("%v != %v", expected, actual)
		}
	}
}

func TestBool(t *testing.T) {
	for actual, expected := range map[bool]bool{
		Bool("FOO", false): true,
		Bool("BAR", false): false,
		Bool("NOPE", true): true,
		Bool("BAZ", true):  true,
	} {
		if expected != actual {
			t.Fatalf("%v != %v", expected, actual)
		}
	}
}

func TestSlice(t *testing.T) {
	for _, c := range []sliceCase{
		sliceCase{
			expected: []string{"a", "b", "c"},
			actual:   Slice("BAZ", ":", []string{}),
		},
		sliceCase{
			expected: []string{"1"},
			actual:   Slice("FOO", ":", []string{}),
		},
		sliceCase{
			expected: []string{"z", "y", "x"},
			actual:   Slice("NOPE", ":", []string{"z", "y", "x"}),
		},
	} {
		if !reflect.DeepEqual(c.expected, c.actual) {
			t.Fatalf("%v != %v", c.expected, c.actual)
		}
	}
}

func TestUint(t *testing.T) {
	for actual, expected := range map[uint64]uint64{
		Uint("FOO", uint64(4)):  uint64(1),
		Uint("BAR", uint64(3)):  uint64(3),
		Uint("BAZ", uint64(5)):  uint64(5),
		Uint("NOPE", uint64(3)): uint64(3),
	} {
		if expected != actual {
			t.Fatalf("%v != %v", expected, actual)
		}
	}
}

func TestUintSize(t *testing.T) {
	for actual, expected := range map[uint64]uint64{
		UintSize("MOAR", uint64(4)): uint64(32000000000),
		UintSize("FOO", uint64(4)):  uint64(1),
		UintSize("BAZ", uint64(4)):  uint64(4),
		UintSize("NOPE", uint64(4)): uint64(4),
	} {
		if expected != actual {
			t.Fatalf("%v != %v", expected, actual)
		}
	}
}

func TestExpandSlice(t *testing.T) {
	for _, c := range []sliceCase{
		sliceCase{
			expected: []string{"1,a:b:c::", "32GB", ""},
			actual:   ExpandSlice([]string{"$FOO,${BAZ}", "$MOAR", "${BAR}"}),
		},
		sliceCase{
			expected: []string{"", ""},
			actual:   ExpandSlice([]string{"$NOPE", "${NOPE}"}),
		},
	} {
		if !reflect.DeepEqual(c.expected, c.actual) {
			t.Fatalf("%v != %v", c.expected, c.actual)
		}
	}
}
