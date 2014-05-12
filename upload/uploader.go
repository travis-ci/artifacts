package upload

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/meatballhat/artifacts/path"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type uploader struct {
	Opts          *Options
	Paths         *path.PathSet
	RetryInterval time.Duration

	log *logrus.Logger
}

// Upload does the deed!
func Upload(opts *Options, log *logrus.Logger) {
	newUploader(opts, log).Upload()
}

func newUploader(opts *Options, log *logrus.Logger) *uploader {
	u := &uploader{
		Opts:          opts,
		Paths:         path.NewPathSet(),
		RetryInterval: 3 * time.Second,

		log: log,
	}

	if opts.Private {
		opts.CacheControl = "private"
	}

	if opts.CacheControl == "" {
		opts.CacheControl = "public, max-age=315360000"
	}

	for _, s := range opts.Paths {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) < 2 {
			parts = append(parts, "")
		}
		u.Paths.Add(path.NewPath(opts.WorkingDir, parts[0], parts[1]))
	}

	return u
}

func (u *uploader) Upload() error {
	done := make(chan bool)
	allDone := 0
	fileChan := u.files()

	for i := 0; i < u.Opts.Concurrency; i++ {
		go func() {
			auth, err := aws.GetAuth(u.Opts.AccessKey, u.Opts.SecretKey)
			if err != nil {
				u.log.WithFields(logrus.Fields{
					"uploader": i,
					"err":      err,
				}).Error("uploader failed to get aws auth")
				done <- true
				return
			}

			conn := s3.New(auth, aws.USEast)
			bucket := conn.Bucket(u.Opts.BucketName)

			if bucket == nil {
				u.log.WithFields(logrus.Fields{
					"uploader": i,
				}).Warn("uploader failed to get bucket")
				done <- true
				return
			}

			for artifact := range fileChan {
				u.uploadFile(bucket, artifact)
			}

			done <- true
		}()
	}

	for {
		select {
		case <-done:
			allDone += 1
			if allDone >= u.Opts.Concurrency {
				return nil
			}
		}
	}

	return nil
}

func (u *uploader) files() chan *artifact {
	artifacts := make(chan *artifact)

	go func() {
		for _, path := range u.Paths.All() {
			to, from, root := path.To, path.From, path.Root
			if path.IsDir() {
				root = filepath.Join(root, from)
				if strings.HasSuffix(root, "/") {
					root = root + "/"
				}
			}

			filepath.Walk(path.Fullpath(), func(f string, info os.FileInfo, err error) error {
				if info != nil && info.IsDir() {
					return nil
				}

				relPath := strings.Replace(strings.Replace(f, root, "", -1), root+"/", "", -1)
				destination := relPath
				if len(to) > 0 {
					if path.IsDir() {
						destination = filepath.Join(to, relPath)
					} else {
						destination = to
					}
				}

				for _, targetPath := range u.Opts.TargetPaths {
					artifacts <- newArtifact(root, relPath, targetPath, destination, u.Opts.Perm)
				}
				return nil
			})

		}
		close(artifacts)
	}()

	return artifacts
}

func (u *uploader) uploadFile(b *s3.Bucket, a *artifact) error {
	retries := 0

	for {
		err := u.rawUpload(b, a)
		if err != nil {
			if retries < u.Opts.Retries {
				retries += 1
				time.Sleep(u.RetryInterval)
				continue
			} else {
				return err
			}
		}
		return nil
	}
}

func (u *uploader) rawUpload(b *s3.Bucket, a *artifact) error {
	destination := a.FullDestination()
	reader, err := a.Reader()
	if err != nil {
		return err
	}

	ctype := a.ContentType()

	u.log.WithFields(logrus.Fields{
		"source":        a.Source,
		"dest":          destination,
		"bucket":        b.Name,
		"content_type":  ctype,
		"cache_control": u.Opts.CacheControl,
	}).Info("uploading to s3")

	err = b.PutReaderHeader(destination, reader, a.Size(),
		map[string][]string{
			"Content-Type":  []string{ctype},
			"Cache-Control": []string{u.Opts.CacheControl},
		}, a.Perm)
	if err != nil {
		u.log.WithFields(logrus.Fields{"err": err}).Error("failed to upload")
		return err
	}

	return nil
}
