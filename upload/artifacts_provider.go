package upload

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
	"github.com/travis-ci/artifacts/client"
)

var (
	defaultProviderRetryInterval = 3 * time.Second
)

type artifactsProvider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger

	overrideClient client.ArtifactPutter
}

func newArtifactsProvider(opts *Options, log *logrus.Logger) *artifactsProvider {
	return &artifactsProvider{
		RetryInterval: defaultProviderRetryInterval,

		opts: opts,
		log:  log,
	}
}

func (ap *artifactsProvider) Upload(id string, opts *Options,
	in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {

	cl := ap.getClient()

	for a := range in {
		err := ap.uploadFile(cl, a)
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

func (ap *artifactsProvider) uploadFile(cl client.ArtifactPutter, a *artifact.Artifact) error {
	retries := uint64(0)

	for {
		err := ap.rawUpload(cl, a)
		if err == nil {
			return nil
		}
		if retries < ap.opts.Retries {
			retries++
			ap.log.WithFields(logrus.Fields{
				"artifact": a.Source,
				"retry":    retries,
				"err":      err,
			}).Debug("retrying")
			time.Sleep(ap.RetryInterval)
			continue
		} else {
			return err
		}
	}
	return nil
}

func (ap *artifactsProvider) rawUpload(cl client.ArtifactPutter, a *artifact.Artifact) error {
	ctype := a.ContentType()
	size, err := a.Size()
	if err != nil {
		return err
	}

	ap.log.WithFields(logrus.Fields{
		"percent_max_size": pctMax(size, ap.opts.MaxSize),
		"max_size":         humanize.Bytes(ap.opts.MaxSize),
		"source":           a.Source,
		"dest":             a.FullDest(),
		"content_type":     ctype,
		"cache_control":    ap.opts.CacheControl,
	}).Debug("more artifact details")

	return cl.PutArtifact(a)
}

func (ap *artifactsProvider) getClient() client.ArtifactPutter {
	if ap.overrideClient != nil {
		ap.log.WithField("client", ap.overrideClient).Debug("using override client")
		return ap.overrideClient
	}

	ap.log.Debug("creating new client")
	return client.New(ap.opts.ArtifactsSaveHost, ap.opts.ArtifactsAuthToken, ap.log)
}

func (ap *artifactsProvider) Name() string {
	return "artifacts"
}
