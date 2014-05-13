package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meatballhat/artifacts/env"
	"github.com/mitchellh/goamz/s3"
)

// Options is used in the call to Upload
type Options struct {
	AccessKey    string
	BucketName   string
	CacheControl string
	Concurrency  int
	Paths        []string
	Perm         s3.ACL
	Private      bool
	Retries      int
	SecretKey    string
	TargetPaths  []string
	WorkingDir   string
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

	return &Options{
		AccessKey: env.Cascade([]string{
			"ARTIFACTS_KEY",
			"ARTIFACTS_AWS_ACCESS_KEY",
			"AWS_ACCESS_KEY_ID",
			"AWS_ACCESS_KEY",
		}, ""),
		SecretKey: env.Cascade([]string{
			"ARTIFACTS_SECRET",
			"ARTIFACTS_AWS_SECRET_KEY",
			"AWS_SECRET_ACCESS_KEY",
			"AWS_SECRET_KEY",
		}, ""),

		BucketName:   strings.TrimSpace(env.Get("ARTIFACTS_S3_BUCKET", "")),
		CacheControl: strings.TrimSpace(env.Get("ARTIFACTS_CACHE_CONTROL", "private")),
		Concurrency:  env.Int("ARTIFACTS_CONCURRENCY", 3),
		Paths:        env.ExpandSlice(env.Slice("ARTIFACTS_PATHS", ";", []string{})),
		Perm:         s3.ACL(env.Get("ARTIFACTS_PERMISSIONS", "private")),
		Private:      env.Bool("ARTIFACTS_PRIVATE", true),
		Retries:      env.Int("ARTIFACTS_RETRIES", 2),
		TargetPaths:  targetPaths,
		WorkingDir:   cwd,
	}
}

// Validate checks for validity!
func (opts *Options) Validate() error {
	if opts.BucketName == "" {
		return fmt.Errorf("no bucket name given")
	}

	if opts.AccessKey == "" {
		return fmt.Errorf("no access key given")
	}

	if opts.SecretKey == "" {
		return fmt.Errorf("no secret key given")
	}

	return nil
}
