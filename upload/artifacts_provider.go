package upload

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/meatballhat/artifacts/artifact"
	"github.com/meatballhat/artifacts/client"
)

type artifactsProvider struct {
	RetryInterval time.Duration

	opts *Options
	log  *logrus.Logger
}

func newArtifactsProvider(opts *Options, log *logrus.Logger) *artifactsProvider {
	return &artifactsProvider{
		RetryInterval: 3 * time.Second,

		opts: opts,
		log:  log,
	}
}

func (ap *artifactsProvider) Upload(id string, opts *Options,
	in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {

	cl := client.New(ap.opts.ArtifactsSaveURL, ap.opts.ArtifactsAuthToken, ap.log)

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

func (ap *artifactsProvider) uploadFile(cl *client.Client, a *artifact.Artifact) error {
	retries := uint64(0)

	for {
		err := ap.rawUpload(cl, a)
		if err == nil {
			return nil
		}
		if retries < ap.opts.Retries {
			retries++
			ap.log.WithFields(logrus.Fields{
				"artifact": a.Path.From,
				"retry":    retries,
			}).Debug("retrying")
			time.Sleep(ap.RetryInterval)
			continue
		} else {
			return err
		}
	}
	return nil
}

func (ap *artifactsProvider) rawUpload(cl *client.Client, a *artifact.Artifact) error {
	ctype := a.ContentType()
	size, err := a.Size()
	if err != nil {
		return err
	}

	ap.log.WithFields(logrus.Fields{
		"percent_max_size": pctMax(size, ap.opts.MaxSize),
		"max_size":         humanize.Bytes(ap.opts.MaxSize),
		"source":           a.Path.From,
		"dest":             a.FullDestination(),
		"content_type":     ctype,
		"cache_control":    ap.opts.CacheControl,
	}).Debug("more artifact details")

	return cl.PutArtifact(a)
}
