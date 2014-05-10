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
		os.Exit(1)
	}

	flag.Parse()
	cmd := flag.Arg(0)

	switch cmd {
	case "upload":
		upload.Upload(upload.NewOptions())
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
}
