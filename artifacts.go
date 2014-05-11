package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/meatballhat/artifacts/upload"
)

var (
	// VersionString contains the compiled-in version number
	VersionString = ""
	// RevisionString contains the compiled-in git rev
	RevisionString = ""

	versionFlag   = flag.Bool("v", false, "Show version and exit")
	logFormatFlag = flag.String("f", "text", "Log output format (text or json)")

	log = logrus.New()
)

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	if *versionFlag {
		fmt.Printf("artifacts version=%v rev=%v\n", VersionString, RevisionString)
		os.Exit(0)
	}

	log.Formatter = &logrus.TextFormatter{}
	if *logFormatFlag == "json" || os.Getenv("ARTIFACTS_LOG_FORMAT") == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}

	cmd := flag.Arg(0)

	switch cmd {
	case "upload":
		opts := upload.NewOptions()
		for i, arg := range flag.Args() {
			if i == 0 {
				continue
			}
			opts.Paths = append(opts.Paths, arg)
		}
		if strings.TrimSpace(opts.BucketName) == "" {
			fmt.Println("no bucket name given")
			os.Exit(2)
		}
		upload.Upload(opts)
	default:
		fmt.Println("what kind of command is", cmd, "...?")
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf(`Usage: artifacts <command>

Commands:
  upload - upload some artifacts!

`)
	flag.PrintDefaults()
}
