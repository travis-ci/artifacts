package upload

import (
	"os"
)

// Options is used in the call to Upload
type Options struct {
	Private bool
	// ClonePath    string
	CacheControl string
	BucketName   string
	TargetPath   string
	WorkingDir   string
	Paths        string
}

// NewOptions makes some *Options with defaults!
func NewOptions() *Options {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = os.Getenv("TRAVIS_BUILD_DIR")
	}

	return &Options{
		Private:      true,
		CacheControl: "private",
		BucketName:   os.Getenv("ARTIFACTS_AWS_S3_BUCKET"),
		TargetPath:   os.Getenv("ARTIFACTS_AWS_TARGET_PATH"),
		WorkingDir:   cwd,
		Paths:        os.Getenv("ARTIFACTS_PATHS"),
	}
}
