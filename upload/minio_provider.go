package upload

import (
	"time"

	"github.com/Sirupsen/logrus"
	minio "github.com/minio/minio-go"
	"github.com/travis-ci/artifacts/artifact"
)

type minioProvider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger
	clnt *minio.Client
}

func newMinioProvider(opts *Options, log *logrus.Logger) *minioProvider {
	return &minioProvider{
		RetryInterval: defaultProviderRetryInterval,
		opts:          opts,
		log:           log,
	}
}

func (mp *minioProvider) retryingUpload(opts *Options, a *artifact.Artifact) error {
	retries := uint64(0)

	for {
		reader, err := a.Reader()
		if err != nil {
			return err
		}

		_, err = mp.clnt.PutObject(opts.BucketName, a.FullDest(), reader, a.ContentType())
		if err == nil {
			return nil
		}

		if retries < opts.Retries {
			retries++
			mp.log.WithFields(logrus.Fields{
				"artifact": a.Source,
				"retry":    retries,
				"err":      err,
			}).Debug("retrying")
			time.Sleep(mp.RetryInterval)
			continue
		} else {
			return err
		}
	}

	return nil
}

func (mp *minioProvider) Upload(id string, opts *Options,
	in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {

	clnt, err := minio.NewWithRegion(opts.Endpoint, opts.AccessKey, opts.SecretKey, false, opts.S3Region)
	if err != nil {
		mp.log.Errorf("Failed to connect to Minio server at %s", opts.Endpoint)
		return
	}

	// Set minio client object.
	mp.clnt = clnt

	for a := range in {
		err := mp.retryingUpload(opts, a)
		if err != nil {
			a.UploadResult.OK = false
			a.UploadResult.Err = err
		} else {
			a.UploadResult.OK = true
		}
		out <- a
	}

	done <- true
}

func (mp *minioProvider) Name() string {
	return "Minio"
}
