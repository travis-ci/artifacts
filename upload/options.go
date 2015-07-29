package upload

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/travis-ci/artifacts/env"
)

const (
	sizeChars = "BKMGTPEZYbkmgtpezy"

	// CommandDescription is the string used to describe the
	// "upload" command in the command line help system
	CommandDescription = `
Upload a set of local paths to an artifact repository.  The paths may be
provided as either positional command-line arguments or as the $ARTIFACTS_PATHS
environment variable, which should be :-delimited.

Paths may be either files or directories.  Any path provided will be walked for
all child entries.  Each entry will have its mime type detected based first on
the file extension, then by sniffing up to the first 512 bytes via the net/http
function "DetectContentType".
`
)

var (
	DefaultOptions = NewOptions()

	optsMaps = map[string]map[string]string{
		"cli": map[string]string{
			"AccessKey":    "key, k",
			"BucketName":   "bucket, b",
			"CacheControl": "cache-control",
			"Perm":         "permissions",
			"SecretKey":    "secret, s",
			"S3Region":     "s3-region",

			"RepoSlug":    "repo-slug, r",
			"BuildNumber": "build-number",
			"BuildID":     "build-id",
			"JobNumber":   "job-number",
			"JobID":       "job-id",

			"Concurrency": "concurrency",
			"MaxSize":     "max-size",
			"Paths":       "",
			"Provider":    "upload-provider, p",
			"Retries":     "retries",
			"TargetPaths": "target-paths, t",
			"WorkingDir":  "working-dir",

			"ArtifactsSaveHost":  "save-host, H",
			"ArtifactsAuthToken": "auth-token, T",
		},
		"doc": map[string]string{
			"AccessKey":    "upload credentials key *REQUIRED*",
			"BucketName":   "destination bucket *REQUIRED*",
			"CacheControl": "artifact cache-control header value",
			"Perm":         "artifact access permissions",
			"SecretKey":    "upload credentials secret *REQUIRED*",
			"S3Region":     "region used when storing to S3",

			"RepoSlug":    "repo owner/name slug",
			"BuildNumber": "build number",
			"BuildID":     "build id",
			"JobNumber":   "job number",
			"JobID":       "job id",

			"Concurrency": "upload worker concurrency",
			"MaxSize":     "max combined size of uploaded artifacts",
			"Paths":       "",
			"Provider":    "artifact upload provider (artifacts, s3, null)",
			"Retries":     "number of upload retries per artifact",
			"TargetPaths": "artifact target paths (':'-delimited)",
			"WorkingDir":  "working directory",

			"ArtifactsSaveHost":  "artifact save host",
			"ArtifactsAuthToken": "artifact save auth token",
		},
		"env": map[string]string{
			"AccessKey":    "ARTIFACTS_KEY,ARTIFACTS_AWS_ACCESS_KEY,AWS_ACCESS_KEY_ID,AWS_ACCESS_KEY",
			"BucketName":   "ARTIFACTS_BUCKET,ARTIFACTS_S3_BUCKET",
			"CacheControl": "ARTIFACTS_CACHE_CONTROL",
			"Perm":         "ARTIFACTS_PERMISSIONS",
			"SecretKey":    "ARTIFACTS_SECRET,ARTIFACTS_AWS_SECRET_KEY,AWS_SECRET_ACCESS_KEY,AWS_SECRET_KEY",
			"S3Region":     "ARTIFACTS_REGION,ARTIFACTS_S3_REGION",

			"RepoSlug":    "ARTIFACTS_REPO_SLUG,TRAVIS_REPO_SLUG",
			"BuildNumber": "ARTIFACTS_BUILD_NUMBER,TRAVIS_BUILD_NUMBER",
			"BuildID":     "ARTIFACTS_BUILD_ID,TRAVIS_BUILD_ID",
			"JobNumber":   "ARTIFACTS_JOB_NUMBER,TRAVIS_JOB_NUMBER",
			"JobID":       "ARTIFACTS_JOB_ID,TRAVIS_JOB_ID",

			"Concurrency": "ARTIFACTS_CONCURRENCY",
			"MaxSize":     "ARTIFACTS_MAX_SIZE",
			"Paths":       "ARTIFACTS_PATHS",
			"Provider":    "ARTIFACTS_UPLOAD_PROVIDER",
			"Retries":     "ARTIFACTS_RETRIES",
			"TargetPaths": "ARTIFACTS_TARGET_PATHS",
			"WorkingDir":  "ARTIFACTS_WORKING_DIR,TRAVIS_BUILD_DIR,PWD",

			"ArtifactsSaveHost":  "ARTIFACTS_SAVE_HOST",
			"ArtifactsAuthToken": "ARTIFACTS_AUTH_TOKEN",
		},
		"default": map[string]string{
			"AccessKey":    "",
			"BucketName":   "",
			"CacheControl": "private",
			"Perm":         "private",
			"SecretKey":    "",
			"S3Region":     "us-east-1",

			"RepoSlug":    "",
			"BuildNumber": "",
			"BuildID":     "",
			"JobNumber":   "",
			"JobID":       "",

			"Concurrency": "5",
			"MaxSize":     fmt.Sprintf("%d", 1024*1024*1000),
			"Paths":       "",
			"Provider":    "s3",
			"Retries":     "2",
			"TargetPaths": "artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER",
			"WorkingDir":  ".",

			"ArtifactsSaveHost":  "",
			"ArtifactsAuthToken": "",
		},
	}
)

// Options is used in the call to Upload
type Options struct {
	AccessKey    string
	BucketName   string
	CacheControl string
	Perm         string
	SecretKey    string
	S3Region     string

	RepoSlug    string
	BuildNumber string
	BuildID     string
	JobNumber   string
	JobID       string

	Concurrency uint64
	MaxSize     uint64
	Paths       []string
	Provider    string
	Retries     uint64
	TargetPaths []string
	WorkingDir  string

	ArtifactsSaveHost  string
	ArtifactsAuthToken string
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

		tf := t.Field(i)
		name := optsMaps["cli"][tf.Name]
		if name == "" {
			continue
		}

		flags = append(flags, cli.StringFlag{
			Name:   name,
			EnvVar: strings.Split(optsMaps["env"][tf.Name], ",")[0],
			Usage: fmt.Sprintf("%v (default %q)",
				optsMaps["doc"][tf.Name],
				fmt.Sprintf("%v", f.Interface())),
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

		tf := t.Field(i)
		dflt := os.ExpandEnv(optsMaps["default"][tf.Name])
		envKeys := strings.Split(optsMaps["env"][tf.Name], ",")
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
	s := reflect.ValueOf(opts).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !f.CanSet() {
			continue
		}

		tf := t.Field(i)

		names := optsMaps["cli"][tf.Name]
		nameParts := strings.Split(names, ",")
		if len(nameParts) < 1 {
			continue
		}

		name := nameParts[0]
		value := c.String(name)
		if value == "" {
			continue
		}

		switch name {
		case "concurrency", "retries":
			intVal, err := strconv.ParseUint(value, 10, 64)
			if err == nil {
				f.SetUint(intVal)
			}
		case "max-size":
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
		case "target-paths":
			tp := []string{}
			for _, part := range strings.Split(value, ":") {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					tp = append(tp, trimmed)
				}
			}
			opts.TargetPaths = tp
		default:
			if f.Kind() == reflect.String {
				f.SetString(value)
			}
		}
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
