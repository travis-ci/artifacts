package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/travis-ci/artifacts/env"
)

var (
	// DefaultCacheControl is the default value for each artifact's Cache-Control header
	DefaultCacheControl = "private"

	// DefaultConcurrency is the default number of concurrent goroutines used during upload
	DefaultConcurrency = uint64(5)

	// DefaultMaxSize is the default maximum allowed bytes for all artifacts
	DefaultMaxSize = uint64(1024 * 1024 * 1000)

	// DefaultPaths is the default slice of local paths to upload (empty)
	DefaultPaths = []string{}

	// DefaultPerm is the default ACL applied to each artifact
	DefaultPerm = "private"

	// DefaultS3Region is the default region used when storing to S3
	DefaultS3Region = aws.USEast

	// DefaultRepoSlug is the repo slug detected from the env
	DefaultRepoSlug = ""

	// DefaultBuildNumber is the build number detected from the env
	DefaultBuildNumber = ""

	// DefaultBuildID is the build id detected from the env
	DefaultBuildID = ""

	// DefaultJobNumber is the build number detected from the env
	DefaultJobNumber = ""

	// DefaultJobID is the build id detected from the env
	DefaultJobID = ""

	// DefaultRetries is the default number of times a given artifact upload will be retried
	DefaultRetries = uint64(2)

	// DefaultTargetPaths is the default upload prefix for each artifact
	DefaultTargetPaths = []string{}

	// DefaultUploadProvider is the provider used to upload (nuts)
	DefaultUploadProvider = "s3"
	// TODO: DefaultUploadProvider = "artifacts"

	// DefaultWorkingDir is the default working directory ... wow.
	DefaultWorkingDir, _ = os.Getwd()
)

const (
	sizeChars = "BKMGTPEZYbkmgtpezy"
)

// Options is used in the call to Upload
type Options struct {
	AccessKey    string `cli:"key" env:"ARTIFACTS_KEY"`
	BucketName   string `cli:"bucket" env:"ARTIFACTS_BUCKET"`
	CacheControl string `cli:"cache-control" env:"ARTIFACTS_CACHE_CONTROL"`
	Perm         s3.ACL `cli:"permissions" env:"ARTIFACTS_PERMISSIONS"`
	SecretKey    string `cli:"secret" env:"ARTIFACTS_SECRET"`
	S3Region     string `cli:"s3-region" env:"ARTIFACTS_S3_REGION"`

	RepoSlug    string `cli:"repo-slug" env:"TRAVIS_REPO_SLUG"`
	BuildNumber string `cli:"build-number" env:"TRAVIS_BUILD_NUMBER"`
	BuildID     string `cli:"build-id" env:"TRAVIS_BUILD_ID"`
	JobNumber   string `cli:"job-number" env:"TRAVIS_JOB_NUMBER"`
	JobID       string `cli:"job-id" env:"TRAVIS_JOB_ID"`

	Concurrency uint64   `cli:"concurrency" env:"ARTIFACTS_CONCURRENCY"`
	MaxSize     uint64   `cli:"max-size" env:"ARTIFACTS_MAX_SIZE"`
	Paths       []string `env:"ARTIFACTS_PATHS"`
	Provider    string   `cli:"upload-provider" env:"ARTIFACTS_UPLOAD_PROVIDER"`
	Retries     uint64   `cli:"retries" env:"ARTIFACTS_RETRIES"`
	TargetPaths []string `cli:"target-paths" env:"ARTIFACTS_TARGET_PATHS"`
	WorkingDir  string   `cli:"working-dir" env:"TRAVIS_BUILD_DIR"`

	ArtifactsSaveHost  string `cli:"save-host" env:"ARTIFACTS_SAVE_HOST"`
	ArtifactsAuthToken string `cli:"auth-token" env:"ARTIFACTS_AUTH_TOKEN"`
}

func init() {
	DefaultTargetPaths = append(DefaultTargetPaths, filepath.Join("artifacts",
		env.Get("TRAVIS_BUILD_NUMBER", ""),
		env.Get("TRAVIS_JOB_NUMBER", "")))

	DefaultRepoSlug = env.Get("TRAVIS_REPO_SLUG", "")
	DefaultBuildNumber = env.Get("TRAVIS_BUILD_NUMBER", "")
	DefaultBuildID = env.Get("TRAVIS_BUILD_ID", "")
	DefaultJobNumber = env.Get("TRAVIS_JOB_NUMBER", "")
	DefaultJobID = env.Get("TRAVIS_JOB_ID", "")
}

// NewOptions makes some *Options with defaults!
func NewOptions() *Options {
	cwd, _ := os.Getwd()
	cwd = env.Get("TRAVIS_BUILD_DIR", cwd)

	targetPaths := env.ExpandSlice(env.Slice("ARTIFACTS_TARGET_PATHS", ":", []string{}))
	if len(targetPaths) == 0 {
		targetPaths = DefaultTargetPaths
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
		BucketName: strings.TrimSpace(env.Cascade([]string{
			"ARTIFACTS_BUCKET",
			"ARTIFACTS_S3_BUCKET",
		}, "")),
		CacheControl: strings.TrimSpace(env.Get("ARTIFACTS_CACHE_CONTROL", DefaultCacheControl)),
		Perm:         s3.ACL(env.Get("ARTIFACTS_PERMISSIONS", DefaultPerm)),
		S3Region:     env.Get("ARTIFACTS_S3_REGION", DefaultS3Region.Name),

		RepoSlug:    DefaultRepoSlug,
		BuildNumber: DefaultBuildNumber,
		BuildID:     DefaultBuildID,
		JobNumber:   DefaultJobNumber,
		JobID:       DefaultJobID,

		Concurrency: env.Uint("ARTIFACTS_CONCURRENCY", DefaultConcurrency),
		MaxSize:     env.UintSize("ARTIFACTS_MAX_SIZE", DefaultMaxSize),
		Paths:       env.ExpandSlice(env.Slice("ARTIFACTS_PATHS", ":", DefaultPaths)),
		Provider:    env.Get("ARTIFACTS_UPLOAD_PROVIDER", DefaultUploadProvider),
		Retries:     env.Uint("ARTIFACTS_RETRIES", DefaultRetries),
		TargetPaths: targetPaths,
		WorkingDir:  cwd,

		ArtifactsSaveHost:  env.Get("ARTIFACTS_SAVE_HOST", ""),
		ArtifactsAuthToken: env.Get("ARTIFACTS_AUTH_TOKEN", ""),
	}
}

// UpdateFromCLI overlays a *cli.Context onto internal options
func (opts *Options) UpdateFromCLI(c *cli.Context) {
	if value := c.String("key"); value != "" {
		opts.AccessKey = value
	}
	if value := c.String("bucket"); value != "" {
		opts.BucketName = value
	}
	if value := c.String("cache-control"); value != "" {
		opts.CacheControl = value
	}
	if value := c.String("permissions"); value != "" {
		opts.Perm = s3.ACL(value)
	}
	if value := c.String("secret"); value != "" {
		opts.SecretKey = value
	}
	if value := c.String("s3-region"); value != "" {
		opts.S3Region = value
	}

	if value := c.String("repo-slug"); value != "" {
		opts.RepoSlug = value
	}
	if value := c.String("build-number"); value != "" {
		opts.BuildNumber = value
	}
	if value := c.String("build-id"); value != "" {
		opts.BuildID = value
	}
	if value := c.String("job-number"); value != "" {
		opts.JobNumber = value
	}
	if value := c.String("job-id"); value != "" {
		opts.JobID = value
	}

	if value := c.String("concurrency"); value != "" {
		intVal, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.Concurrency = intVal
		}
	}
	if value := c.String("max-size"); value != "" {
		if strings.ContainsAny(value, sizeChars) {
			b, err := humanize.ParseBytes(value)
			if err == nil {
				opts.MaxSize = b
			}
		} else {
			intVal, err := strconv.ParseUint(value, 10, 64)
			if err == nil {
				opts.MaxSize = intVal
			}
		}
	}
	if value := c.String("upload-provider"); value != "" {
		opts.Provider = value
	}
	if value := c.String("retries"); value != "" {
		intVal, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.Retries = intVal
		}
	}
	if value := c.String("target-paths"); value != "" {
		tp := []string{}
		for _, part := range strings.Split(value, ":") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				tp = append(tp, trimmed)
			}
		}
		opts.TargetPaths = tp
	}
	if value := c.String("working-dir"); value != "" {
		opts.WorkingDir = value
	}

	if value := c.String("save-host"); value != "" {
		opts.ArtifactsSaveHost = value
	}
	if value := c.String("auth-token"); value != "" {
		opts.ArtifactsAuthToken = value
	}

	for _, arg := range c.Args() {
		opts.Paths = append(opts.Paths, arg)
	}
}

// Validate checks for validity!
func (opts *Options) Validate() error {
	if opts.Provider == "s3" {
		return opts.validateS3()
	}

	return nil
}

func (opts *Options) validateS3() error {
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
