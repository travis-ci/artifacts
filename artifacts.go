package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/meatballhat/artifacts/upload"
)

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		flag.PrintDefaults()
		die()
	}

	flag.Parse()
	cmd := flag.Arg(0)

	switch cmd {
	case "upload":
		upload.Upload()
	default:
		die("what kind of command is", cmd, "...?")
	}
}

func usage() {
	fmt.Printf(`Usage: artifacts <command>

Commands:
  upload - upload some artifacts!
`)
}

func die(whatever ...interface{}) {
	if len(whatever) > 0 {
		fmt.Println(whatever...)
	}
	os.Exit(1)
}
