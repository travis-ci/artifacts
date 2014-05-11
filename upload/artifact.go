package upload

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type artifact struct {
	Root           string
	RelativeSource string
	Source         string
	Destination    string
	Prefix         string
}

func newArtifact(root, relativeSource, prefix, destination string) *artifact {
	return &artifact{
		Root:           root,
		RelativeSource: relativeSource,
		Source:         filepath.Join(root, relativeSource),
		Prefix:         prefix,
		Destination:    destination,
	}
}

func (a *artifact) ContentType() string {
	return "application/octet-stream"
}

func (a *artifact) Reader() (io.Reader, error) {
	f, err := os.Open(a.Source)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (a *artifact) Size() int64 {
	fi, err := os.Stat(a.Source)
	if err != nil {
		return int64(0)
	}

	return fi.Size()
}

func (a *artifact) FullDestination() string {
	return strings.TrimLeft(filepath.Join(a.Prefix, a.Destination), "/")
}
