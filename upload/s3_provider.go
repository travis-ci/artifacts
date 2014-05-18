package upload

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type s3Provider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger
}

func newS3Provider(opts *Options, log *logrus.Logger) *s3Provider {
	return &s3Provider{
		RetryInterval: 3 * time.Second,

		opts: opts,
		log:  log,
	}
}

func (s3p *s3Provider) Upload(id string, opts *Options, in chan *artifact, out chan *artifact, done chan bool) {
	auth, err := aws.GetAuth(opts.AccessKey, opts.SecretKey)

	if err != nil {
		s3p.log.WithFields(logrus.Fields{
			"uploader": id,
			"err":      err,
		}).Error("uploader failed to get aws auth")
		done <- true
		return
	}

	conn := s3.New(auth, aws.USEast)
	bucket := conn.Bucket(opts.BucketName)

	if bucket == nil {
		s3p.log.WithFields(logrus.Fields{
			"uploader": id,
		}).Warn("uploader failed to get bucket")
		done <- true
		return
	}

	for artifact := range in {
		err := s3p.uploadFile(opts, bucket, artifact)
		if err != nil {
			artifact.Result.OK = false
			artifact.Result.Err = err
		} else {
			artifact.Result.OK = true
		}
		out <- artifact
	}

	done <- true
	return
}

func (s3p *s3Provider) Name() string {
	return "s3"
}

func (s3p *s3Provider) uploadFile(opts *Options, b *s3.Bucket, a *artifact) error {
	retries := uint64(0)

	for {
		err := s3p.rawUpload(opts, b, a)
		if err != nil {
			if retries < opts.Retries {
				retries += 1
				s3p.log.WithFields(logrus.Fields{
					"artifact": a.Path.From,
					"retry":    retries,
				}).Debug("retrying")
				time.Sleep(s3p.RetryInterval)
				continue
			} else {
				return err
			}
		}
		return nil
	}
}

func (s3p *s3Provider) rawUpload(opts *Options, b *s3.Bucket, a *artifact) error {
	destination := a.FullDestination()
	reader, err := a.Reader()
	if err != nil {
		return err
	}

	ctype := a.ContentType()
	size, err := a.Size()
	if err != nil {
		return err
	}

	s3p.log.WithFields(logrus.Fields{
		"download_url": fmt.Sprintf("https://s3.amazonaws.com/%s/%s", b.Name, destination),
	}).Info(fmt.Sprintf("uploading: %s (size: %s)", a.Path.From, humanize.Bytes(size)))

	s3p.log.WithFields(logrus.Fields{
		"percent_max_size": pctMax(size, opts.MaxSize),
		"max_size":         humanize.Bytes(opts.MaxSize),
		"source":           a.Path.From,
		"dest":             destination,
		"bucket":           b.Name,
		"content_type":     ctype,
		"cache_control":    opts.CacheControl,
	}).Debug("more artifact details")

	err = b.PutReaderHeader(destination, reader, int64(size),
		map[string][]string{
			"Content-Type":  []string{ctype},
			"Cache-Control": []string{opts.CacheControl},
		}, a.Perm)
	if err != nil {
		return err
	}

	return nil
}

func pctMax(artifactSize, maxSize uint64) float64 {
	return float64(100.0) * (float64(artifactSize) / float64(maxSize))
}
