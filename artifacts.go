package main

import (
	"fmt"

	"github.com/mitchellh/goamz/s3"
)

func main() {
	fmt.Printf("%#v\n", &s3.S3{})
}
