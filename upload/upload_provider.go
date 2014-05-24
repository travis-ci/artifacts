package upload

import (
	"github.com/meatballhat/artifacts/artifact"
)

type uploadProvider interface {
	Upload(string, *Options,
		chan *artifact.Artifact, chan *artifact.Artifact, chan bool)
}
