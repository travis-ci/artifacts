package upload

import (
	"os"
	"testing"
)

func TestNewUploader(t *testing.T) {
	os.Setenv("ARTIFACTS_AWS_S3_BUCKET", "foo")
	os.Setenv("ARTIFACTS_AWS_TARGET_PATH", "/baz")
	os.Setenv("ARTIFACTS_PATHS", "bin/*;derp")

	u := newUploader()
	if u == nil {
		t.Fail()
	}

	if u.BucketName != "foo" {
		t.Fail()
	}

	if u.TargetPath != "/baz" {
		t.Fail()
	}

	if len(u.Paths) != 2 {
		t.Fail()
	}
}
