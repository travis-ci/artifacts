package artifact

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
)

type testPath struct {
	Path        string
	ContentType string
	Valid       bool
}

var (
	testTmp, err        = ioutil.TempDir("", "artifacts-test-upload")
	testArtifactPathDir = filepath.Join(testTmp, "artifact")
	testArtifactPaths   = []*testPath{
		&testPath{
			Path:        filepath.Join(testArtifactPathDir, "foo"),
			ContentType: "text/plain; charset=utf-8",
			Valid:       true,
		},
		&testPath{
			Path:        filepath.Join(testArtifactPathDir, "foo.csv"),
			ContentType: "text/csv; charset=utf-8",
			Valid:       true,
		},
		&testPath{
			Path:        filepath.Join(testArtifactPathDir, "nonexistent"),
			ContentType: defaultCtype,
			Valid:       false,
		},
		&testPath{
			Path:        filepath.Join(testArtifactPathDir, "unreadable"),
			ContentType: defaultCtype,
			Valid:       false,
		},
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

	for _, p := range testArtifactPaths {
		if filepath.Base(p.Path) == "nonexistent" {
			continue
		}

		fd, err := os.Create(p.Path)
		if err != nil {
			log.Panicf("game over: %v\n", err)
		}

		defer fd.Close()

		for i := 0; i < 512; i++ {
			fmt.Fprintf(fd, "something\n")
		}

		if filepath.Base(p.Path) == "unreadable" {
			fd.Chmod(0000)
		}
	}
}

func TestNewArtifact(t *testing.T) {
	a := New("bucket", "/foo/bar", "linux/foo", &Options{
		Perm:     s3.BucketCannedACLPublicRead,
		RepoSlug: "owner/foo",
	})
	if a == nil {
		t.Fatalf("new artifact is nil")
	}

	if a.Prefix != "bucket" {
		t.Fatalf("prefix not set correctly: %v", a.Prefix)
	}

	if a.Dest != "linux/foo" {
		t.Fatalf("destination not set correctly: %v", a.Dest)
	}

	if a.Perm != s3.BucketCannedACLPublicRead {
		t.Fatalf("s3 perm not set correctly: %v", a.Perm)
	}

	if a.UploadResult == nil {
		t.Fatalf("result not initialized")
	}

	if a.UploadResult.OK {
		t.Fatalf("result initialized with OK as true")
	}

	if a.UploadResult.Err != nil {
		t.Fatalf("result initialized with non-nil Err")
	}
}

func TestArtifactContentType(t *testing.T) {
	for _, p := range testArtifactPaths {
		a := New("bucket", p.Path, "linux/foo", &Options{
			Perm:     s3.BucketCannedACLPublicRead,
			RepoSlug: "owner/foo",
		})
		if a == nil {
			t.Fatalf("new artifact is nil")
		}

		actualCtype := a.ContentType()
		if p.ContentType != actualCtype {
			t.Fatalf("%v: %v != %v", p.Path, p.ContentType, actualCtype)
		}
	}
}

func TestArtifactReader(t *testing.T) {
	for _, p := range testArtifactPaths {
		if !p.Valid {
			continue
		}

		a := New("bucket", p.Path, "linux/foo", &Options{
			Perm:     s3.BucketCannedACLPublicRead,
			RepoSlug: "owner/foo",
		})
		if a == nil {
			t.Fatalf("new artifact is nil")
		}

		reader, err := a.Reader()
		if err != nil {
			t.Fatalf("error getting reader: %v", err)
		}

		_, err = ioutil.ReadAll(reader)
		if err != nil {
			t.Error(err)
		}
	}
}
