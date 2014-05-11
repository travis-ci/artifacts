package upload

import (
	"os"
	"testing"
)

func TestNewUploader(t *testing.T) {
	os.Setenv("TRAVIS_BUILD_NUMBER", "3")
	os.Setenv("TRAVIS_JOB_NUMBER", "3.2")

	os.Setenv("ARTIFACTS_S3_BUCKET", "foo")
	os.Setenv("ARTIFACTS_TARGET_PATHS", "baz;artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER")
	os.Setenv("ARTIFACTS_PATHS", "bin/;derp")

	u := newUploader(NewOptions())
	if u == nil {
		t.Errorf("options are %v", u)
	}

	if u.BucketName != "foo" {
		t.Errorf("bucket name is %v", u.BucketName)
	}

	if len(u.TargetPaths) != 2 {
		t.Errorf("target paths length != 2: %v", len(u.TargetPaths))
	}

	if u.TargetPaths[0] != "baz" {
		t.Errorf("target paths[0] != baz: %v", u.TargetPaths)
	}

	if u.TargetPaths[1] != "artifacts/3/3.2" {
		t.Errorf("target paths[1] != artifacts/3/3.2: %v", u.TargetPaths)
	}

	if len(u.Paths.All()) != 2 {
		t.Errorf("all paths length != 2: %v", len(u.Paths.All()))
	}
}
