package upload

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/travis-ci/artifacts/artifact"
)

var (
	nilAuth aws.Auth
)

type s3Provider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger

	overrideConn *s3.S3
	overrideAuth aws.Auth
}

func newS3Provider(opts *Options, log *logrus.Logger) *s3Provider {
	return &s3Provider{
		RetryInterval: defaultProviderRetryInterval,

		opts: opts,
		log:  log,

		overrideAuth: nilAuth,
	}
}

func (s3p *s3Provider) Upload(id string, opts *Options, in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {
	auth, err := s3p.getAuth(opts.AccessKey, opts.SecretKey)

	if err != nil {
		s3p.log.WithFields(logrus.Fields{
			"uploader": id,
			"err":      err,
		}).Error("uploader failed to get aws auth")
		done <- true
		return
	}

	conn := s3p.getConn(auth)
	bucket := conn.Bucket(opts.BucketName)

	if bucket == nil {
		s3p.log.WithFields(logrus.Fields{
			"uploader": id,
		}).Warn("uploader failed to get bucket")
		done <- true
		return
	}

	for a := range in {
		err := s3p.uploadFile(opts, bucket, a)
		if err != nil {
			a.UploadResult.OK = false
			a.UploadResult.Err = err
		} else {
			a.UploadResult.OK = true
		}
		out <- a
	}

	done <- true
	return
}

func (s3p *s3Provider) uploadFile(opts *Options, b *s3.Bucket, a *artifact.Artifact) error {
	retries := uint64(0)

	for {
		err := s3p.rawUpload(opts, b, a)
		if err == nil {
			return nil
		}
		if retries < opts.Retries {
			retries++
			s3p.log.WithFields(logrus.Fields{
				"artifact": a.Source,
				"retry":    retries,
				"err":      err,
			}).Debug("retrying")
			time.Sleep(s3p.RetryInterval)
			continue
		} else {
			return err
		}
	}
	return nil
}

func (s3p *s3Provider) rawUpload(opts *Options, b *s3.Bucket, a *artifact.Artifact) error {
	dest := a.FullDest()
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
		"download_url": fmt.Sprintf("%s/%s/%s", s3p.getRegion().S3Endpoint, b.Name, dest),
	}).Info(fmt.Sprintf("uploading: %s (size: %s)", a.Source, humanize.Bytes(size)))

	s3p.log.WithFields(logrus.Fields{
		"percent_max_size": pctMax(size, opts.MaxSize),
		"max_size":         humanize.Bytes(opts.MaxSize),
		"source":           a.Source,
		"dest":             dest,
		"bucket":           b.Name,
		"content_type":     ctype,
		"cache_control":    opts.CacheControl,
	}).Debug("more artifact details")

	err = b.PutReaderHeader(dest, reader, int64(size),
		map[string][]string{
			"Content-Type":  []string{ctype},
			"Cache-Control": []string{opts.CacheControl},
		}, a.Perm)
	if err != nil {
		return err
	}

	return nil
}

func (s3p *s3Provider) getConn(auth aws.Auth) *s3.S3 {
	if s3p.overrideConn != nil {
		s3p.log.WithField("conn", s3p.overrideConn).Debug("using override connection")
		return s3p.overrideConn
	}

	return s3.New(auth, s3p.getRegion())
}

func (s3p *s3Provider) getAuth(accessKey, secretKey string) (aws.Auth, error) {
	if s3p.overrideAuth != nilAuth {
		s3p.log.WithField("auth", s3p.overrideAuth).Debug("using override auth")
		return s3p.overrideAuth, nil
	}

	s3p.log.Debug("creating new auth")
	return aws.GetAuth(accessKey, secretKey)
}

func (s3p *s3Provider) getRegion() aws.Region {
	region, ok := aws.Regions[s3p.opts.S3Region]

	if !ok {
		s3p.log.WithFields(logrus.Fields{
			"region":  s3p.opts.S3Region,
			"default": DefaultOptions.S3Region,
		}).Warn(fmt.Sprintf("invalid region, defaulting to %s", DefaultOptions.S3Region))
		region = aws.Regions[DefaultOptions.S3Region]
	}

	return region
}

func (s3p *s3Provider) Name() string {
	return "s3"
}
