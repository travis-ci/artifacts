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
	TargetPaths  []string
	WorkingDir   string
	Paths        []string
	Concurrency  int
	Retries      int
	// ClonePath    string
}

// NewOptions makes some *Options with defaults!
func NewOptions() *Options {
	cwd, _ := os.Getwd()
	cwd = env.Get("TRAVIS_BUILD_DIR", cwd)

	targetPaths := env.ExpandSlice(env.Slice("ARTIFACTS_TARGET_PATHS", ";", []string{}))
	if len(targetPaths) == 0 {
		targetPaths = append(targetPaths, filepath.Join("artifacts",
			env.Get("TRAVIS_BUILD_NUMBER", ""),
			env.Get("TRAVIS_JOB_NUMBER", "")))
	}

	private := env.Bool("ARTIFACTS_PRIVATE", true)

	return &Options{
		Private:     private,
		BucketName:  env.Get("ARTIFACTS_S3_BUCKET", ""),
		TargetPaths: targetPaths,
		WorkingDir:  cwd,
		Paths:       env.ExpandSlice(env.Slice("ARTIFACTS_PATHS", ";", []string{})),
		Concurrency: env.Int("ARTIFACTS_CONCURRENCY", 3),
		Retries:     env.Int("ARTIFACTS_RETRIES", 2),
	}
}
