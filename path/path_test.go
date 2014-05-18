package path

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	testTmp, err      = ioutil.TempDir("", "artifacts-test")
	testSomethingPath = filepath.Join(testTmp, "something")
	testSomethingBoop = filepath.Join(testSomethingPath, "boop")

	fullPathTests = map[string][]string{
		"/abc/ham":        []string{"/abc", "ham", "bone"},
		"/flim":           []string{"/nope", "/flim", "flam"},
		testSomethingBoop: []string{"/bogus", testSomethingBoop, "boop"},
	}
	isAbsTests = map[string]bool{
		"fiddle/faddle":   false,
		testSomethingBoop: true,
	}
	isDirTests = map[string]bool{
		testSomethingPath:          true,
		"this/had/better/not/work": false,
	}
)

func init() {
	if err != nil {
		log.Panicf("game over: %v\n", err)
	}

	err = os.MkdirAll(testSomethingPath, 0755)
	if err != nil {
		log.Panicf("game over: %v\n", err)
	}

	fd, err := os.Create(testSomethingBoop)
	if err != nil {
		log.Panicf("game over: %v\n", err)
	}

	defer fd.Close()
	fmt.Fprintf(fd, "something\n")
}

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

func TestPathIsAbs(t *testing.T) {
	for path, truth := range isAbsTests {
		p := NewPath("/whatever", path, "somewhere")
		if p.IsAbs() != truth {
			t.Errorf("path %v IsAbs != %v\n", path, truth)
		}
	}
}

func TestPathFullpath(t *testing.T) {
	for expected, args := range fullPathTests {
		actual := NewPath(args[0], args[1], args[2]).Fullpath()
		if expected != actual {
			t.Errorf("%v != %v", expected, actual)
		}
	}
}

func TestPathIsDir(t *testing.T) {
	for path, truth := range isDirTests {
		p := NewPath("/whatever", path, "somewhere")
		if p.IsDir() != truth {
			t.Errorf("path %v IsDir != %v\n", path, truth)
		}
	}
}
