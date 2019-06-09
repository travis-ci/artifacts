package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
	"github.com/travis-ci/artifacts/path"
)

const (
	defaultPublicCacheControl = "public, max-age=315360000"
)

type uploader struct {
	Opts          *Options
	Paths         *path.Set
	RetryInterval time.Duration
	Provider      uploadProvider

	log       *logrus.Logger
	curSize   *maxSizeTracker
	startTime time.Time
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
	var provider uploadProvider

	if opts.CacheControl == "" {
		opts.CacheControl = defaultPublicCacheControl
	}

	if opts.Provider == "" {
		opts.Provider = "s3"
	}

	switch opts.Provider {
	case "artifacts":
		provider = newArtifactsProvider(opts, log)
	case "s3":
		provider = newS3Provider(opts, log)
	case "null":
		provider = newNullProvider(nil, log)
	default:
		log.WithFields(logrus.Fields{
			"provider": opts.Provider,
		}).Warn("unrecognized provider, using s3 instead")
		provider = newS3Provider(opts, log)
	}

	u := &uploader{
		Opts:     opts,
		Paths:    path.NewSet(),
		Provider: provider,

		log:       log,
		startTime: time.Now(),
	}

	for _, s := range opts.Paths {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) < 2 {
			parts = append(parts, "")
		}

		p := path.New(opts.WorkingDir, parts[0], parts[1])
		log.WithFields(logrus.Fields{"path": p}).Debug("adding path")
		u.Paths.Add(p)
	}

	return u
}

func (u *uploader) Upload() error {
	u.log.Debug("starting upload")
	u.startTime = time.Now()
	done := make(chan bool)
	allDone := uint64(0)
	inChan := u.files()
	outChan := make(chan *artifact.Artifact)
	failed := []*artifact.Artifact{}

	defer func() {
		if len(failed) == 0 {
			return
		}

		for _, a := range failed {
			u.log.WithFields(logrus.Fields{
				"err": a.UploadResult.Err,
			}).Error(fmt.Sprintf("failed to upload: %s", a.Source))
		}
	}()

	u.log.WithFields(logrus.Fields{
		"bucket":        u.Opts.BucketName,
		"cache_control": u.Opts.CacheControl,
		"permissions":   u.Opts.Perm,
	}).Info("uploading with settings")

	u.log.WithFields(logrus.Fields{
		"working_dir":  u.Opts.WorkingDir,
		"target_paths": u.Opts.TargetPaths,
		"concurrency":  u.Opts.Concurrency,
		"max_size":     u.Opts.MaxSize,
		"retries":      u.Opts.Retries,
	}).Debug("other upload settings")

	for i := uint64(0); i < u.Opts.Concurrency; i++ {
		u.log.WithFields(logrus.Fields{
			"uploader": i,
		}).Debug("starting uploader worker")

		go u.Provider.Upload(fmt.Sprintf("%d", i), u.Opts, inChan, outChan, done)
	}

	for {
		select {
		case outArtifact := <-outChan:
			if outArtifact != nil && !outArtifact.UploadResult.OK {
				failed = append(failed, outArtifact)
			}
		case <-done:
			allDone++
			if allDone >= u.Opts.Concurrency {
				return nil
			}
		}
	}

	return nil
}

func (u *uploader) artifactFeederLoop(path *path.Path, artifacts chan *artifact.Artifact) error {
	to, from, root := path.To, path.From, path.Root
	u.log.WithField("path", path).Debug("incoming path")

	if path.IsDir() {
		root = filepath.Join(root, from)
		u.log.WithField("root", root).Debug("path is dir, so setting root to root+from")
	}

	artifactOpts := &artifact.Options{
		Perm:        u.Opts.Perm,
		RepoSlug:    u.Opts.RepoSlug,
		BuildNumber: u.Opts.BuildNumber,
		BuildID:     u.Opts.BuildID,
		JobNumber:   u.Opts.JobNumber,
		JobID:       u.Opts.JobID,
	}

	filepath.Walk(path.Fullpath(), func(source string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			u.log.WithField("path", source).Debug("skipping directory")
			return nil
		}

		relPath := strings.Replace(strings.Replace(source, root, "", -1), root+"/", "", -1)
		dest := relPath
		if len(to) > 0 {
			if path.IsDir() {
				dest = filepath.Join(to, relPath)
			} else {
				dest = to
			}
		}

		for _, targetPath := range u.Opts.TargetPaths {
			err := func() error {
				u.curSize.Lock()
				defer u.curSize.Unlock()

				a := artifact.New(targetPath, source, dest, artifactOpts)

				size, err := a.Size()
				if err != nil {
					return err
				}

				u.curSize.Current += size
				logFields := logrus.Fields{
					"current_size":     humanize.Bytes(u.curSize.Current),
					"max_size":         humanize.Bytes(u.Opts.MaxSize),
					"percent_max_size": pctMax(size, u.Opts.MaxSize),
					"artifact":         relPath,
					"artifact_size":    humanize.Bytes(size),
				}

				if u.curSize.Current > u.Opts.MaxSize {
					msg := "max-size would be exceeded"
					u.log.WithFields(logFields).Error(msg)
					return fmt.Errorf(msg)
				}

				u.log.WithFields(logFields).Debug("queueing artifact")
				artifacts <- a
				return nil
			}()
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func (u *uploader) artifactFeeder(artifacts chan *artifact.Artifact) error {
	u.curSize = &maxSizeTracker{Current: uint64(0)}

	i := 0
	for _, path := range u.Paths.All() {
		u.artifactFeederLoop(path, artifacts)
		i++
	}

	u.log.WithFields(logrus.Fields{
		"total_size":   humanize.Bytes(u.curSize.Current),
		"count":        i,
		"time_elapsed": time.Since(u.startTime),
	}).Debug("done feeding artifacts")

	close(artifacts)
	return nil
}

func (u *uploader) files() chan *artifact.Artifact {
	artifacts := make(chan *artifact.Artifact)
	go u.artifactFeeder(artifacts)
	return artifacts
}
