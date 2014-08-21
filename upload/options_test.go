package upload

import (
	"os"
	"testing"
)

func TestOptionsValidate(t *testing.T) {
	os.Clearenv()
	opts := NewOptions()
	opts.Provider = "null"

	if opts.Validate() != nil {
		t.Fatalf("default options were invalid")
	}
}

func TestOptionsValidateS3(t *testing.T) {
	os.Clearenv()
	opts := NewOptions()
	opts.Provider = "s3"

	err := opts.Validate()
	if err == nil {
		t.Fatalf("default options were valid for s3")
	}

	if err.Error() != "no bucket name given" {
		t.Fatalf("default options did not fail on missing bucket name")
	}

	opts.BucketName = "foo"
	err = opts.Validate()
	if err == nil {
		t.Fatalf("options with only bucket name were valid for s3")
	}

	if err.Error() != "no access key given" {
		t.Fatalf("options did not fail on missing access key")
	}

	opts.AccessKey = "AZ123"
	err = opts.Validate()
	if err == nil {
		t.Fatalf("options with only bucket name were valid for s3")
	}

	if err.Error() != "no secret key given" {
		t.Fatalf("options did not fail on missing secret key")
	}

	opts.SecretKey = "ZYX321"
	if opts.Validate() != nil {
		t.Fatalf("valid s3 options were deemed invalid")
	}
}
