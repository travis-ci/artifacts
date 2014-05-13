package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/meatballhat/artifacts/upload"
)

var (
	// VersionString contains the compiled-in version number
	VersionString = "?"
	// RevisionString contains the compiled-in git rev
	RevisionString = "?"

	log = logrus.New()
)

func main() {
	app := cli.NewApp()
	app.Name = "artifacts"
	app.Usage = "manage your artifacts!"
	app.Version = VersionString
	app.Flags = []cli.Flag{
		cli.StringFlag{"log-format, f", "text", "log output format (text or json)"},
		cli.BoolFlag{"debug, D", "set log level to debug"},
	}
	app.Commands = []cli.Command{
		{
			Name:      "upload",
			ShortName: "u",
			Usage:     "upload some artifacts!",
			Flags: []cli.Flag{
				cli.StringFlag{"key, k", "", "upload credentials key [ARTIFACTS_KEY]"},
				cli.StringFlag{"secret, s", "", "upload credentials secret [ARTIFACTS_SECRET]"},
			},
			Action: func(c *cli.Context) {
				configureLog(log, c)

				opts := upload.NewOptions()
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
