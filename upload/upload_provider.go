package upload

import (
	"github.com/travis-ci/artifacts/artifact"
)

type uploadProvider interface {
	Upload(string, *Options,
		chan *artifact.Artifact, chan *artifact.Artifact, chan bool)
	Name() string
}
