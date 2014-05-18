package upload

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	apath "github.com/meatballhat/artifacts/path"
	"github.com/mitchellh/goamz/s3"
)

const (
	defaultCtype = "application/octet-stream"
)

type artifact struct {
	Path        *apath.Path
	Destination string
	Prefix      string
	Perm        s3.ACL

	Result *result
}

func newArtifact(path *apath.Path, prefix, destination string, perm s3.ACL) *artifact {
	return &artifact{
		Path:        path,
		Prefix:      prefix,
		Destination: destination,
		Perm:        perm,

		Result: &result{},
	}
}

func (a *artifact) ContentType() string {
	ctype := mime.TypeByExtension(path.Ext(a.Path.From))
	if ctype != "" {
		return ctype
	}

	f, err := os.Open(a.Path.Fullpath())
	if err != nil {
		return defaultCtype
	}

	var buf bytes.Buffer

	_, err = io.CopyN(&buf, f, int64(512))
	if err != nil && err != io.EOF {
		return defaultCtype
	}

	return http.DetectContentType(buf.Bytes())
}

func (a *artifact) Reader() (io.Reader, error) {
	f, err := os.Open(a.Path.Fullpath())
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (a *artifact) Size() (uint64, error) {
	fi, err := os.Stat(a.Path.Fullpath())
	if err != nil {
		return uint64(0), nil
	}

	return uint64(fi.Size()), nil
}

func (a *artifact) FullDestination() string {
	return strings.TrimLeft(filepath.Join(a.Prefix, a.Destination), "/")
}
