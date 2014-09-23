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

func TestExpandSlice(t *testing.T) {
	for _, c := range []sliceCase{
		sliceCase{
			expected: []string{"1,a:b:c::", "32GB", ""},
			actual:   expandSlice([]string{"$FOO,${BAZ}", "$MOAR", "${BAR}"}),
		},
		sliceCase{
			expected: []string{"", ""},
			actual:   expandSlice([]string{"$NOPE", "${NOPE}"}),
		},
	} {
		if !reflect.DeepEqual(c.expected, c.actual) {
			t.Fatalf("%v != %v", c.expected, c.actual)
		}
	}
}
