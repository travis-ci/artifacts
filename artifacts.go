package main

import (
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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
provided as either positional command-line arguments or as the ARTIFACTS_PATHS
environmental variable, which should be ';'-delimited.

Paths may be either files or directories.  Any path provided will be walked for
all child entries.  Each entry will have its mime type detected based first on
the file extension, then by sniffing up to the first 512 bytes via the net/http
function "DetectContentType".
`
)

func main() {
	app := cli.NewApp()
	app.Name = "artifacts"
	app.Usage = "manage your artifacts!"
	app.Version = VersionString
	app.Flags = []cli.Flag{
		cli.StringFlag{"log-format, f", "", "log output format (text or json)"},
		cli.BoolFlag{"debug, D", "set log level to debug"},
	}
	app.Commands = []cli.Command{
		{
			Name:        "upload",
			ShortName:   "u",
			Usage:       "upload some artifacts!",
			Description: uploadDescription,
			Flags: []cli.Flag{
				cli.StringFlag{"key, k", "", "upload credentials key [ARTIFACTS_KEY] *REQUIRED*"},
				cli.StringFlag{"secret, s", "", "upload credentials secret [ARTIFACTS_SECRET] *REQUIRED*"},
				cli.StringFlag{"bucket, b", "", "destination bucket [ARTIFACTS_BUCKET] *REQUIRED*"},
				cli.StringFlag{"cache-control", "", "artifact cache-control header value [ARTIFACTS_CACHE_CONTROL]"},
				cli.StringFlag{"concurrency", "", "upload worker concurrency [ARTIFACTS_CONCURRENCY]"},
				cli.StringFlag{"permissions", "", "artifact access permissions [ARTIFACTS_PERMISSIONS]"},
				cli.StringFlag{"retries", "", "number of upload retries per artifact [ARTIFACT_RETRIES]"},
				cli.StringFlag{"target-paths, t", "", "artifact target paths (';'-delimited) [ARTIFACTS_TARGET_PATHS]"},
				cli.StringFlag{"working-dir", "", "working directory [PWD, TRAVIS_BUILD_DIR]"},
			},
			Action: func(c *cli.Context) {
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
			},
		},
	}

	app.Run(os.Args)
}

func configureLog(log *logrus.Logger, c *cli.Context) {
	log.Formatter = &logrus.TextFormatter{}
	if c.String("format") == "json" || os.Getenv("ARTIFACTS_LOG_FORMAT") == "json" {
		log.Formatter = &logrus.JSONFormatter{}
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
			opts.Concurrency = int(intVal)
		}
	}
	if value := c.String("permissions"); value != "" {
		opts.Perm = s3.ACL(value)
	}
	if value := c.String("private"); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			opts.Private = boolVal
		}
	}
	if value := c.String("retries"); value != "" {
		intVal, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.Retries = int(intVal)
		}
	}
	if value := c.String("working-dir"); value != "" {
		opts.WorkingDir = value
	}
}
