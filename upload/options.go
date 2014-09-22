package upload

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/goamz/s3"
	"github.com/travis-ci/artifacts/env"
)

const (
	sizeChars = "BKMGTPEZYbkmgtpezy"
)

var (
	DefaultOptions = NewOptions()
)

// Options is used in the call to Upload
type Options struct {
	AccessKey    string `cli:"key, k" doc:"upload credentials key *REQUIRED*" env:"ARTIFACTS_KEY,ARTIFACTS_AWS_ACCESS_KEY,AWS_ACCESS_KEY_ID,AWS_ACCESS_KEY"`
	BucketName   string `cli:"bucket, b" doc:"destination bucket *REQUIRED*" env:"ARTIFACTS_BUCKET,ARTIFACTS_S3_BUCKET"`
	CacheControl string `cli:"cache-control" doc:"artifact cache-control header value" env:"ARTIFACTS_CACHE_CONTROL" default:"private"`
	Perm         s3.ACL `cli:"permissions" doc:"artifact access permissions" env:"ARTIFACTS_PERMISSIONS" default:"private"`
	SecretKey    string `cli:"secret, s" doc:"upload credentials secret *REQUIRED*" env:"ARTIFACTS_SECRET,ARTIFACTS_AWS_SECRET_KEY,AWS_SECRET_ACCESS_KEY,AWS_SECRET_KEY"`
	S3Region     string `cli:"s3-region" doc:"region used when storing to S3" env:"ARTIFACTS_S3_REGION" default:"us-east-1"`

	RepoSlug    string `cli:"repo-slug, r" doc:"repo owner/name slug" env:"ARTIFACTS_REPO_SLUG,TRAVIS_REPO_SLUG"`
	BuildNumber string `cli:"build-number" doc:"build number" env:"ARTIFACTS_BUILD_NUMBER,TRAVIS_BUILD_NUMBER"`
	BuildID     string `cli:"build-id" doc:"build id" env:"ARTIFACTS_BUILD_ID,TRAVIS_BUILD_ID"`
	JobNumber   string `cli:"job-number" doc:"job number" env:"ARTIFACTS_JOB_NUMBER,TRAVIS_JOB_NUMBER"`
	JobID       string `cli:"job-id" doc:"job id" env:"ARTIFACTS_JOB_ID,TRAVIS_JOB_ID"`

	Concurrency uint64   `cli:"concurrency" doc:"upload worker concurrency" env:"ARTIFACTS_CONCURRENCY" default:"5"`
	MaxSize     uint64   `cli:"max-size" doc:"max combined size of uploaded artifacts" env:"ARTIFACTS_MAX_SIZE" default:"1048576000"`
	Paths       []string `env:"ARTIFACTS_PATHS"`
	Provider    string   `cli:"upload-provider, p" doc:"artifact upload provider (artifacts, s3, null)" env:"ARTIFACTS_UPLOAD_PROVIDER" default:"s3"`
	Retries     uint64   `cli:"retries" doc:"number of upload retries per artifact" env:"ARTIFACTS_RETRIES" default:"2"`
	TargetPaths []string `cli:"target-paths, t" doc:"artifact target paths (':'-delimited) " env:"ARTIFACTS_TARGET_PATHS" default:"artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER"`
	WorkingDir  string   `cli:"working-dir" doc:"working directory" env:"ARTIFACTS_WORKING_DIR,TRAVIS_BUILD_DIR,PWD" default:"."`

	ArtifactsSaveHost  string `cli:"save-host, H" doc:"artifact save host" env:"ARTIFACTS_SAVE_HOST"`
	ArtifactsAuthToken string `cli:"auth-token, T" doc:"artifact save auth token" env:"ARTIFACTS_AUTH_TOKEN"`
}

// NewOptions makes some *Options with defaults!
func NewOptions() *Options {
	opts := &Options{}
	opts.reset()
	return opts
}

func (opts *Options) Flags() []cli.Flag {
	dflt := DefaultOptions
	flags := []cli.Flag{}

	s := reflect.ValueOf(dflt).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !f.CanSet() {
			continue
		}

		tag := t.Field(i).Tag
		name := tag.Get("cli")
		if name == "" {
			continue
		}

		flags = append(flags, cli.StringFlag{
			Name:   name,
			EnvVar: strings.Split(tag.Get("env"), ",")[0],
			Usage:  fmt.Sprintf("%v (default %q)", tag.Get("doc"), fmt.Sprintf("%v", f.Interface())),
		})
	}

	return flags
}

func (opts *Options) reset() {
	s := reflect.ValueOf(opts).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !f.CanSet() {
			continue
		}

		tag := t.Field(i).Tag
		dflt := os.ExpandEnv(tag.Get("default"))
		envKeys := strings.Split(tag.Get("env"), ",")
		value, envVar := env.CascadeMatch(envKeys, dflt)

		if value == "" {
			continue
		}

		if envVar == "" {
			envVar = envKeys[0]
		}

		value = os.ExpandEnv(value)

		k := f.Kind()
		switch k {
		case reflect.String:
			f.SetString(value)
		case reflect.Uint64:
			uintVal, err := strconv.ParseUint(dflt, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v", err)
			} else {
				f.SetUint(env.Uint(envVar, uintVal))
			}
		case reflect.Slice:
			sliceValue := env.Slice(envVar, ":", strings.Split(":", dflt))
			f.Set(reflect.ValueOf(sliceValue))
		default:
			panic(fmt.Sprintf("unknown kind wat: %v", k))
		}
	}
}

// UpdateFromCLI overlays a *cli.Context onto internal options
func (opts *Options) UpdateFromCLI(c *cli.Context) {
	// FIXME use reflection for this bit, too
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
