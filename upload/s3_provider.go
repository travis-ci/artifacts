package upload

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
)

type s3Provider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger

	s3manager      s3Manager
	sess           *session.Session
	getSessionOnce sync.Once
}

type s3Manager interface {
	NewUploader(c client.ConfigProvider, options ...func(*s3manager.Uploader)) s3Uploader
}

type s3Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// awsS3Manager is the default S3 manager that uses the actual AWS SDK
type awsS3Manager struct{}

func (m *awsS3Manager) NewUploader(c client.ConfigProvider, options ...func(*s3manager.Uploader)) s3Uploader {
	return s3manager.NewUploader(c, options...)
}

func newS3Provider(opts *Options, log *logrus.Logger) *s3Provider {
	return &s3Provider{
		RetryInterval: defaultProviderRetryInterval,

		opts:      opts,
		log:       log,
		s3manager: new(awsS3Manager),
	}
}

func (s3p *s3Provider) Upload(id string, opts *Options, in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {
	err := s3p.getSession(opts)
	if err != nil {
		s3p.log.WithFields(logrus.Fields{
			"uploader": id,
			"err":      err,
		}).Error("uploader failed to get aws auth")
		done <- true
		return
	}

	for a := range in {
		err := s3p.uploadFile(opts, a)
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

func (s3p *s3Provider) uploadFile(opts *Options, a *artifact.Artifact) error {
	retries := uint64(0)

	for {
		err := s3p.rawUpload(opts, a)
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
}

func (s3p *s3Provider) rawUpload(opts *Options, a *artifact.Artifact) error {
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

	downloadHost := fmt.Sprintf("https://%s.s3-%s.amazonaws.com", opts.BucketName, opts.S3Region)

	s3p.log.WithFields(logrus.Fields{
		"download_url": fmt.Sprintf("%s/%s", downloadHost, dest),
	}).Info(fmt.Sprintf("uploading: %s (size: %s)", a.Source, humanize.Bytes(size)))

	s3p.log.WithFields(logrus.Fields{
		"percent_max_size": pctMax(size, opts.MaxSize),
		"max_size":         humanize.Bytes(opts.MaxSize),
		"source":           a.Source,
		"dest":             dest,
		"bucket":           opts.BucketName,
		"content_type":     ctype,
		"cache_control":    opts.CacheControl,
	}).Debug("more artifact details")

	uploadInput := &s3manager.UploadInput{
		ACL:          aws.String(opts.Perm),
		Body:         reader,
		Bucket:       aws.String(opts.BucketName),
		CacheControl: aws.String(opts.CacheControl),
		ContentType:  aws.String(ctype),
		Key:          aws.String(dest),
	}

	_, err = s3p.s3manager.NewUploader(s3p.sess).Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}

func (s3p *s3Provider) getSession(opts *Options) error {
	var err error
	s3p.getSessionOnce.Do(func() {
		s3p.sess, err = session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:      aws.String(opts.S3Region),
				Credentials: credentials.NewStaticCredentials(opts.AccessKey, opts.SecretKey, ""),
			},
		})
	})
	return err
}

func (s3p *s3Provider) Name() string {
	return "s3"
}
