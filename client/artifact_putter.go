package client

import "github.com/travis-ci/artifacts/artifact"

type ArtifactPutter interface {
	PutArtifact(*artifact.Artifact) error
}
