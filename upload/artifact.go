package upload

import (
	"io"
	"os"
)

type artifact struct {
	Source      string
	Destination string
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
