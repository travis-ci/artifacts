package upload

import (
	"os"
	"path/filepath"

	"github.com/meatballhat/artifacts/env"
)

// Options is used in the call to Upload
type Options struct {
	Private      bool
	CacheControl string
	BucketName   string
	TargetPath   string
	WorkingDir   string
	Paths        []string
	// ClonePath    string
}

// NewOptions makes some *Options with defaults!
func NewOptions() *Options {
	cwd, _ := os.Getwd()
	cwd = env.Get("TRAVIS_BUILD_DIR", cwd)

	targetPath := env.Get("ARTIFACTS_AWS_TARGET_PATH",
		filepath.Join("artifacts",
			env.Get("TRAVIS_BUILD_NUMBER", ""),
			env.Get("TRAVIS_JOB_NUMBER", "")))

	private := env.Bool("ARTIFACTS_PRIVATE", true)

	return &Options{
		Private:    private,
		BucketName: env.Get("ARTIFACTS_AWS_S3_BUCKET", ""),
		TargetPath: targetPath,
		WorkingDir: cwd,
		Paths:      env.Getslice("ARTIFACTS_PATHS", ";", []string{}),
	}
}
