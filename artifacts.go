package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/goamz/s3"
	"github.com/travis-ci/artifacts/logging"
	"github.com/travis-ci/artifacts/upload"
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
		cli.StringFlag{
			Name:   "key, k",
			EnvVar: "ARTIFACTS_KEY",
			Usage:  "upload credentials key *REQUIRED*",
		},
		cli.StringFlag{
			Name:   "bucket, b",
			EnvVar: "ARTIFACTS_BUCKET",
			Usage:  "destination bucket *REQUIRED*",
		},
		cli.StringFlag{
			Name:   "cache-control",
			EnvVar: "ARTIFACTS_CACHE_CONTROL",
			Usage:  fmt.Sprintf("artifact cache-control header value (default %q)", upload.DefaultCacheControl),
		},
		cli.StringFlag{
			Name:   "permissions",
			EnvVar: "ARTIFACTS_PERMISSIONS",
			Usage:  fmt.Sprintf("artifact access permissions (default %q)", upload.DefaultPerm),
		},
		cli.StringFlag{
			Name:   "secret, s",
			EnvVar: "ARTIFACTS_SECRET",
			Usage:  "upload credentials secret *REQUIRED*",
		},

		cli.StringFlag{
			Name:   "concurrency",
			EnvVar: "ARTIFACTS_CONCURRENCY",
			Usage:  fmt.Sprintf("upload worker concurrency (default %v)", upload.DefaultConcurrency),
		},
		cli.StringFlag{
			Name:   "max-size",
			EnvVar: "ARTIFACTS_MAX_SIZE",
			Usage:  fmt.Sprintf("max combined size of uploaded artifacts (default %v)", humanize.Bytes(upload.DefaultMaxSize)),
		},
		cli.StringFlag{
			Name:   "retries",
			EnvVar: "ARTIFACTS_RETRIES",
			Usage:  fmt.Sprintf("number of upload retries per artifact (default %v)", upload.DefaultRetries),
		},
		cli.StringFlag{
			Name:   "target-paths, t",
			EnvVar: "ARTIFACTS_TARGET_PATHS",
			Usage:  fmt.Sprintf("artifact target paths (':'-delimited) (default %#v)", upload.DefaultTargetPaths),
		},
		cli.StringFlag{
			Name:   "working-dir",
			EnvVar: "TRAVIS_BUILD_DIR",
			Usage:  "working directory ($TRAVIS_BUILD_DIR) (default $PWD)",
		},

		cli.StringFlag{
			Name:   "upload-provider, p",
			EnvVar: "ARTIFACTS_UPLOAD_PROVIDER",
			Usage:  fmt.Sprintf("artifact upload provider (artifacts, s3, null) (default %#v)", upload.DefaultUploadProvider),
		},

		cli.StringFlag{
			Name:   "save-host, H",
			EnvVar: "ARTIFACTS_SAVE_HOST",
			Usage:  "artifact save host",
		},
		cli.StringFlag{
			Name:   "auth-token, T",
			EnvVar: "ARTIFACTS_AUTH_TOKEN",
			Usage:  "artifact save auth token",
		},
	}
)

func main() {
	app := buildApp()
	app.Run(os.Args)
}

func buildApp() *cli.App {
	app := cli.NewApp()
	app.Name = "artifacts"
	app.Usage = "manage your artifacts!"
	app.Version = VersionString
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-format, f",
			EnvVar: "ARTIFACTS_LOG_FORMAT",
			Usage:  "log output format (text, json, or multiline)",
		},
		cli.BoolFlag{
			Name:   "debug, D",
			EnvVar: "ARTIFACTS_DEBUG",
			Usage:  "set log level to debug",
		},
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

	return app
}

func runUpload(c *cli.Context) {
	configureLog(log, c)

	opts := upload.NewOptions()
	overlayFlags(opts, c)

	for _, arg := range c.Args() {
		log.WithFields(logrus.Fields{
			"path": arg,
		}).Debug("adding path from command line args")
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

	if formatArg == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}
	if formatArg == "multiline" {
		log.Formatter = &logging.MultiLineFormatter{}
	}
	if c.Bool("debug") {
		log.Level = logrus.Debug
		log.Debug("setting log level to debug")
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
	if value := c.String("upload-provider"); value != "" {
		opts.Provider = value
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
}
