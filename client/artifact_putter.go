package client

import "github.com/travis-ci/artifacts/artifact"

// ArtifactPutter is the interface used to put artifacts
type ArtifactPutter interface {
	PutArtifact(*artifact.Artifact) error
}
