package upload

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/meatballhat/artifacts/path"
	"github.com/mitchellh/goamz/s3"
)

var (
	testArtifactPathDir = filepath.Join(testTmp, "artifact")
	testArtifactPaths   = map[string]string{
		filepath.Join(testArtifactPathDir, "foo"):     "text/plain; charset=utf-8",
		filepath.Join(testArtifactPathDir, "foo.csv"): "text/csv; charset=utf-8",
	}
)

func init() {
	err = os.MkdirAll(testArtifactPathDir, 0755)
	if err != nil {
		log.Panicf("game over: %v\n", err)
	}

	for filepath := range testArtifactPaths {
		fd, err := os.Create(filepath)
		if err != nil {
			log.Panicf("game over: %v\n", err)
		}

		defer fd.Close()
		for i := 0; i < 512; i++ {
			fmt.Fprintf(fd, "something\n")
		}
	}
}

func TestNewArtifact(t *testing.T) {
	p := path.NewPath("/", "foo", "bar")
	a := newArtifact(p, "bucket", "linux/foo", s3.PublicRead)
	if a == nil {
		t.Errorf("new artifact is nil")
	}

	if a.Path != p {
		t.Errorf("path not set correctly: %v", a.Path)
	}

	if a.Prefix != "bucket" {
		t.Errorf("prefix not set correctly: %v", a.Prefix)
	}

	if a.Destination != "linux/foo" {
		t.Errorf("destination not set correctly: %v", a.Destination)
	}

	if a.Perm != s3.PublicRead {
		t.Errorf("s3 perm not set correctly: %v", a.Perm)
	}

	if a.Result == nil {
		t.Errorf("result not initialized")
	}

	if a.Result.OK {
		t.Errorf("result initialized with OK as true")
	}

	if a.Result.Err != nil {
		t.Errorf("result initialized with non-nil Err")
	}
}

func TestArtifactContentType(t *testing.T) {
	for filepath, expectedCtype := range testArtifactPaths {
		p := path.NewPath("whatever", filepath, "somewhere")
		a := newArtifact(p, "bucket", "linux/foo", s3.PublicRead)
		if a == nil {
			t.Errorf("new artifact is nil")
		}

		actualCtype := a.ContentType()
		if expectedCtype != actualCtype {
			t.Errorf("%v != %v", expectedCtype, actualCtype)
		}
	}
}

func TestArtifactReader(t *testing.T) {
	for filepath := range testArtifactPaths {
		p := path.NewPath("whatever", filepath, "somewhere")
		a := newArtifact(p, "bucket", "linux/foo", s3.PublicRead)
		if a == nil {
			t.Errorf("new artifact is nil")
		}

		reader, err := a.Reader()
		if err != nil {
			t.Error(err)
		}

		_, err = ioutil.ReadAll(reader)
		if err != nil {
			t.Error(err)
		}
	}
}
