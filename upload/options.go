package upload

import (
	"os"
	"path/filepath"
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

	targetPath := os.Getenv("ARTIFACTS_AWS_TARGET_PATH")
	if len(targetPath) == 0 {
		targetPath = filepath.Join("artifacts",
			os.Getenv("TRAVIS_BUILD_NUMBER"),
			os.Getenv("TRAVIS_JOB_NUMBER"))
	}

	return &Options{
		Private:      true,
		CacheControl: "private",
		BucketName:   os.Getenv("ARTIFACTS_AWS_S3_BUCKET"),
		TargetPath:   targetPath,
		WorkingDir:   cwd,
		Paths:        os.Getenv("ARTIFACTS_PATHS"),
	}
}
