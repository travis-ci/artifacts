package artifact

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/goamz/s3"
	"github.com/travis-ci/artifacts/path"
)

var (
	testTmp, err        = ioutil.TempDir("", "artifacts-test-upload")
	testArtifactPathDir = filepath.Join(testTmp, "artifact")
	testArtifactPaths   = map[string]string{
		filepath.Join(testArtifactPathDir, "foo"):     "text/plain; charset=utf-8",
		filepath.Join(testArtifactPathDir, "foo.csv"): "text/csv; charset=utf-8",
	}
)

func init() {
	if err != nil {
		log.Panicf("game over: %v\n", err)
	}

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
	p := path.New("/", "foo", "bar")
	a := New(p, "bucket", "linux/foo", &Options{
		Perm:     s3.PublicRead,
		RepoSlug: "owner/foo",
	})
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

	if a.UploadResult == nil {
		t.Errorf("result not initialized")
	}

	if a.UploadResult.OK {
		t.Errorf("result initialized with OK as true")
	}

	if a.UploadResult.Err != nil {
		t.Errorf("result initialized with non-nil Err")
	}
}

func TestArtifactContentType(t *testing.T) {
	for filepath, expectedCtype := range testArtifactPaths {
		p := path.New("whatever", filepath, "somewhere")
		a := New(p, "bucket", "linux/foo", &Options{
			Perm:     s3.PublicRead,
			RepoSlug: "owner/foo",
		})
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
		p := path.New("whatever", filepath, "somewhere")
		a := New(p, "bucket", "linux/foo", &Options{
			Perm:     s3.PublicRead,
			RepoSlug: "owner/foo",
		})
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
