package upload

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type uploader struct {
	BucketName string
	Paths      []string
	TargetPath string
}

// Upload does the deed!
func Upload() {
	newUploader().Upload()
}

func newUploader() *uploader {
	u := &uploader{
		BucketName: os.Getenv("ARTIFACTS_AWS_S3_BUCKET"),
		TargetPath: os.Getenv("ARTIFACTS_AWS_TARGET_PATH"),
	}

	paths := os.Getenv("ARTIFACTS_PATHS")

	u.Paths = []string{}
	for _, s := range strings.Split(paths, ";") {
		u.Paths = append(u.Paths, strings.TrimSpace(s))
	}

	return u
}

func (u *uploader) Upload() error {
	auth, err := aws.GetAuth("", "")
	if err != nil {
		return err
	}

	conn := s3.New(auth, aws.USEast)
	bucket := conn.Bucket(u.BucketName)

	if bucket == nil {
		return fmt.Errorf("failed to get bucket")
	}

	fmt.Println("just kidding!")
	return nil
}
