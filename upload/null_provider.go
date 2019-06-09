package upload

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
)

var (
	errUploadFailed = fmt.Errorf("upload failed")
)

type nullProvider struct {
	SourcesToFail []string

	Log *logrus.Logger
}

func newNullProvider(sourcesToFail []string, log *logrus.Logger) *nullProvider {
	if sourcesToFail == nil {
		sourcesToFail = []string{}
	}
	if log == nil {
		log = logrus.New()
	}
	return &nullProvider{
		SourcesToFail: sourcesToFail,

		Log: log,
	}
}

func (np *nullProvider) Upload(id string, opts *Options,
	in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {

	sort.Strings(np.SourcesToFail)
	lenSrc := len(np.SourcesToFail)

	for a := range in {
		idx := sort.SearchStrings(np.SourcesToFail, a.Source)
		if idx < 0 || idx >= lenSrc {
			a.UploadResult.OK = false
			a.UploadResult.Err = errUploadFailed
		} else {
			a.UploadResult.OK = true
		}
		np.Log.WithField("artifact", a).Debug("not really uploading")
		out <- a
	}

	done <- true
}

func (np *nullProvider) Name() string {
	return "null"
}
