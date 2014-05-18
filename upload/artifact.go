package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/goamz/s3"
)

var (
	errInvalidSource = fmt.Errorf("invalid source path")
)

const (
	defaultCtype = "application/octet-stream"
)

type artifact struct {
	Root           string
	RelativeSource string
	Source         string
	Destination    string
	Prefix         string
	Perm           s3.ACL

	Result *result
}

func newArtifact(root, relativeSource, prefix, destination string, perm s3.ACL) *artifact {
	return &artifact{
		Root:           root,
		RelativeSource: relativeSource,
		Source:         filepath.Join(root, relativeSource),
		Prefix:         prefix,
		Destination:    destination,
		Perm:           perm,

		Result: &result{},
	}
}

// ContentType infers the content type of the source file
func (a *artifact) ContentType() string {
	ctype := mime.TypeByExtension(path.Ext(a.Source))
	if ctype != "" {
		return ctype
	}

	f, err := os.Open(a.Source)
	if err != nil {
		return defaultCtype
	}

	var buf bytes.Buffer

	_, err = io.CopyN(&buf, f, int64(512))
	if err != nil {
		return defaultCtype
	}

	return http.DetectContentType(buf.Bytes())
}

// Reader returns an io.Reader suitable for stream-y things
func (a *artifact) Reader() (io.Reader, error) {
	src, err := a.FullSource()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Size attempts to get the file size from the source
func (a *artifact) Size() (uint64, error) {
	src, err := a.FullSource()
	if err != nil {
		return uint64(0), err
	}

	fi, err := os.Stat(src)
	if err != nil {
		return uint64(0), nil
	}

	return uint64(fi.Size()), nil
}

// FullDestination is the combined Prefix and Destination
func (a *artifact) FullDestination() string {
	return strings.TrimLeft(filepath.Join(a.Prefix, a.Destination), "/")
}

func (a *artifact) FullSource() (string, error) {
	for _, candidate := range []string{a.RelativeSource, a.Source} {
		_, err := os.Stat(candidate)
		if err == nil {
			return candidate, nil
		}
	}

	return "", errInvalidSource
}
