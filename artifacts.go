package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/meatballhat/artifacts/logging"
	"github.com/meatballhat/artifacts/upload"
	"github.com/mitchellh/goamz/s3"
)

var (
	// VersionString contains the compiled-in version number
	VersionString = "?"
	// RevisionString contains the compiled-in git rev
	RevisionString = "?"

	log = logrus.New()
)

const (
	uploadDescription = `
Upload a set of local paths to an artifact repository.  The paths may be
provided as either positional command-line arguments or as the $ARTIFACTS_PATHS
environmental variable, which should be :-delimited.

Paths may be either files or directories.  Any path provided will be walked for
all child entries.  Each entry will have its mime type detected based first on
the file extension, then by sniffing up to the first 512 bytes via the net/http
function "DetectContentType".
`
)

var (
	uploadFlags = []cli.Flag{
		cli.StringFlag{"key, k", "", "upload credentials key ($ARTIFACTS_KEY) *REQUIRED*"},
		cli.StringFlag{"secret, s", "", "upload credentials secret ($ARTIFACTS_SECRET) *REQUIRED*"},
		cli.StringFlag{"bucket, b", "", "destination bucket ($ARTIFACTS_BUCKET) *REQUIRED*"},
		cli.StringFlag{"cache-control", "", fmt.Sprintf("artifact cache-control header value ($ARTIFACTS_CACHE_CONTROL) (default %q)", upload.DefaultCacheControl)},
		cli.StringFlag{"concurrency", "", fmt.Sprintf("upload worker concurrency ($ARTIFACTS_CONCURRENCY) (default %v)", upload.DefaultConcurrency)},
		cli.StringFlag{"max-size", "", fmt.Sprintf("max combined size of uploaded artifacts ($ARTIFACTS_MAX_SIZE) (default %v)", humanize.Bytes(upload.DefaultMaxSize))},
		cli.StringFlag{"permissions", "", fmt.Sprintf("artifact access permissions ($ARTIFACTS_PERMISSIONS) (default %q)", upload.DefaultPerm)},
		cli.StringFlag{"retries", "", fmt.Sprintf("number of upload retries per artifact ($ARTIFACT_RETRIES) (default %v)", upload.DefaultRetries)},
		cli.StringFlag{"target-paths, t", "", fmt.Sprintf("artifact target paths (':'-delimited) ($ARTIFACTS_TARGET_PATHS) (default %#v)", upload.DefaultTargetPaths)},
		cli.StringFlag{"working-dir", "", "working directory ($TRAVIS_BUILD_DIR) (default $PWD)"},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "artifacts"
	app.Usage = "manage your artifacts!"
	app.Version = VersionString
	app.Flags = []cli.Flag{
		cli.StringFlag{"log-format, f", "", "log output format (text, json, or multiline)"},
		cli.BoolFlag{"debug, D", "set log level to debug"},
	}
	app.Commands = []cli.Command{
		{
			Name:        "upload",
			ShortName:   "u",
			Usage:       "upload some artifacts!",
			Description: uploadDescription,
			Flags:       uploadFlags,
			Action:      runUpload,
		},
	}

	app.Run(os.Args)
}

func runUpload(c *cli.Context) {
	configureLog(log, c)

	opts := upload.NewOptions()
	overlayFlags(opts, c)

	for i, arg := range c.Args() {
		if i == 0 {
			continue
		}
		opts.Paths = append(opts.Paths, arg)
	}

	if err := opts.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := upload.Upload(opts, log); err != nil {
		log.Fatal(err)
	}
}

func configureLog(log *logrus.Logger, c *cli.Context) {
	log.Formatter = &logrus.TextFormatter{}

	formatArg := c.GlobalString("log-format")
	formatEnv := os.Getenv("ARTIFACTS_LOG_FORMAT")

	if formatArg == "json" || formatEnv == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}
	if formatArg == "multiline" || formatEnv == "multiline" {
		log.Formatter = &logging.MultiLineFormatter{}
	}
	if c.Bool("debug") || os.Getenv("ARTIFACTS_DEBUG") != "" {
		log.Level = logrus.Debug
	}
}

func overlayFlags(opts *upload.Options, c *cli.Context) {
	if value := c.String("key"); value != "" {
		opts.AccessKey = value
	}
	if value := c.String("secret"); value != "" {
		opts.SecretKey = value
	}
	if value := c.String("bucket"); value != "" {
		opts.BucketName = value
	}
	if value := c.String("cache-control"); value != "" {
		opts.CacheControl = value
	}
	if value := c.String("concurrency"); value != "" {
		intVal, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.Concurrency = intVal
		}
	}
	if value := c.String("max-size"); value != "" {
		if strings.ContainsAny(value, "BKMGTPEZYbkmgtpezy") {
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
	if value := c.String("permissions"); value != "" {
		opts.Perm = s3.ACL(value)
	}
	if value := c.String("retries"); value != "" {
		intVal, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.Retries = intVal
		}
	}
	if value := c.String("working-dir"); value != "" {
		opts.WorkingDir = value
	}
}
