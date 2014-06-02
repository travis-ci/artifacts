package upload

import (
	"fmt"
	"sort"

	"github.com/travis-ci/artifacts/artifact"
)

var (
	errUploadFailed = fmt.Errorf("upload failed")
)

type nullProvider struct {
	SourcesToFail []string
}

func newNullProvider(sourcesToFail []string) *nullProvider {
	return &nullProvider{
		SourcesToFail: sourcesToFail,
	}
}

func (np *nullProvider) Upload(id string, opts *Options,
	in chan *artifact.Artifact, out chan *artifact.Artifact, done chan bool) {

	sort.Strings(np.SourcesToFail)
	lenSrc := len(np.SourcesToFail)

	for a := range in {
		idx := sort.SearchStrings(np.SourcesToFail, a.Path.From)
		if idx < 0 || idx >= lenSrc {
			a.UploadResult.OK = false
			a.UploadResult.Err = errUploadFailed
		} else {
			a.UploadResult.OK = true
		}
		out <- a
	}

	done <- true
}
