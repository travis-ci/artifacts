package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
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

type maxSizeTracker struct {
	sync.Mutex
	Current uint64
}

// Upload does the deed!
func Upload(opts *Options, log *logrus.Logger) error {
	return newUploader(opts, log).Upload()
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
	u.log.Debug("starting upload")
	done := make(chan bool)
	allDone := uint64(0)
	fileChan := u.files()

	for i := uint64(0); i < u.Opts.Concurrency; i++ {
		u.log.WithFields(logrus.Fields{
			"uploader": i,
		}).Debug("starting uploader worker")

		go func() {
			uploader := 0 + i
			auth, err := aws.GetAuth(u.Opts.AccessKey, u.Opts.SecretKey)
			if err != nil {
				u.log.WithFields(logrus.Fields{
					"uploader": uploader,
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

func (u *uploader) artifactFeederLoop(path *path.Path, artifacts chan *artifact, curSize *maxSizeTracker) error {
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
			return func() error {
				curSize.Lock()
				defer curSize.Unlock()

				artifact := newArtifact(root, relPath, targetPath, destination, u.Opts.Perm)
				size := artifact.Size()
				curSize.Current += size
				logFields := logrus.Fields{
					"current_size":     humanize.Bytes(curSize.Current),
					"max_size":         humanize.Bytes(u.Opts.MaxSize),
					"percent_max_size": pctMax(size, u.Opts.MaxSize),
					"artifact":         relPath,
					"artifact_size":    humanize.Bytes(artifact.Size()),
				}

				if curSize.Current > u.Opts.MaxSize {
					msg := "max-size would be exceeded"
					u.log.WithFields(logFields).Error(msg)
					return fmt.Errorf(msg)
				}

				u.log.WithFields(logFields).Debug("queueing artifact")
				artifacts <- artifact
				return nil
			}()
		}
		return nil
	})

	return nil
}

func (u *uploader) artifactFeeder(artifacts chan *artifact) error {
	curSize := &maxSizeTracker{Current: uint64(0)}

	i := 0
	for _, path := range u.Paths.All() {
		u.artifactFeederLoop(path, artifacts, curSize)
		i += 1
	}

	u.log.WithFields(logrus.Fields{
		"total_size": curSize.Current,
		"count":      i,
	}).Debug("done feeding artifacts")

	close(artifacts)
	return nil
}

func (u *uploader) files() chan *artifact {
	artifacts := make(chan *artifact)
	go u.artifactFeeder(artifacts)
	return artifacts
}

func (u *uploader) uploadFile(b *s3.Bucket, a *artifact) error {
	retries := uint64(0)

	for {
		err := u.rawUpload(b, a)
		if err != nil {
			if retries < u.Opts.Retries {
				retries += 1
				u.log.WithFields(logrus.Fields{
					"artifact": a.Source,
					"retry":    retries,
				}).Debug("retrying")
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
		"artifact_size":    humanize.Bytes(a.Size()),
		"percent_max_size": pctMax(a.Size(), u.Opts.MaxSize),
		"max_size":         humanize.Bytes(u.Opts.MaxSize),
		"source":           a.Source,
		"dest":             destination,
		"bucket":           b.Name,
		"content_type":     ctype,
		"cache_control":    u.Opts.CacheControl,
	}).Info("uploading to s3")

	err = b.PutReaderHeader(destination, reader, int64(a.Size()),
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

func pctMax(artifactSize, maxSize uint64) float64 {
	return float64(100.0) * (float64(artifactSize) / float64(maxSize))
}
