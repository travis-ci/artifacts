package upload

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestNewUploader(t *testing.T) {
	setenvs(map[string]string{
		"TRAVIS_BUILD_NUMBER":    "3",
		"TRAVIS_JOB_NUMBER":      "3.2",
		"ARTIFACTS_S3_BUCKET":    "foo",
		"ARTIFACTS_TARGET_PATHS": "baz;artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER",
		"ARTIFACTS_PATHS":        "bin/;derp",
	})

	log := logrus.New()
	log.Level = logrus.Panic

	u := newUploader(NewOptions(), log)
	if u == nil {
		t.Errorf("options are %v", u)
	}

	if u.Opts.BucketName != "foo" {
		t.Errorf("bucket name is %v", u.Opts.BucketName)
	}

	if len(u.Opts.TargetPaths) != 2 {
		t.Errorf("target paths length != 2: %v", len(u.Opts.TargetPaths))
	}

	if u.Opts.TargetPaths[0] != "baz" {
		t.Errorf("target paths[0] != baz: %v", u.Opts.TargetPaths)
	}

	if u.Opts.TargetPaths[1] != "artifacts/3/3.2" {
		t.Errorf("target paths[1] != artifacts/3/3.2: %v", u.Opts.TargetPaths)
	}

	if len(u.Paths.All()) != 2 {
		t.Errorf("all paths length != 2: %v", len(u.Paths.All()))
	}
}
