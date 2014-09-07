package upload

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
			ContentType: "application/octet-stream",
			Valid:       false,
		},
		&testPath{
			Path:        filepath.Join(testArtifactPathDir, "unreadable"),
			ContentType: "application/octet-stream",
			Valid:       false,
		},
	}
)

func init() {
	os.Clearenv()

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

func setenvs(e map[string]string) error {
	for k, v := range e {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
