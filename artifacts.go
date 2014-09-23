package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/travis-ci/artifacts/logging"
	"github.com/travis-ci/artifacts/upload"
)

var (
	// VersionString contains the compiled-in version number
	VersionString = "?"
	// RevisionString contains the compiled-in git rev
	RevisionString = "?"
)

func main() {
	app := buildApp()
	app.Run(os.Args)
}

func buildApp() *cli.App {
	app := cli.NewApp()
	app.Name = "artifacts"
	app.Usage = "manage your artifacts!"
	app.Version = fmt.Sprintf("%s revision=%s", VersionString, RevisionString)
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
			Description: upload.CommandDescription,
			Flags:       upload.DefaultOptions.Flags(),
			Action:      runUpload,
		},
	}

	return app
}

func runUpload(c *cli.Context) {
	log := configureLog(c)

	opts := upload.NewOptions()
	opts.UpdateFromCLI(c)

	if err := opts.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := upload.Upload(opts, log); err != nil {
		log.Fatal(err)
	}
}

func configureLog(c *cli.Context) *logrus.Logger {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{}

	formatArg := c.GlobalString("log-format")

	if formatArg == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}
	if formatArg == "multiline" {
		log.Formatter = &logging.MultiLineFormatter{}
	}
	if c.GlobalBool("debug") {
		log.Level = logrus.DebugLevel
		log.Debug("setting log level to debug")
	}

	return log
}
