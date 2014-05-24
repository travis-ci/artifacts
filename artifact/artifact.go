package artifact

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

// Artifact is the thing that gets uploaded or whatever
type Artifact struct {
	RepoSlug    string
	BuildNumber string
	JobNumber   string

	Path        *apath.Path
	Destination string
	Prefix      string
	Perm        s3.ACL

	UploadResult *Result
}

// New creates a new *Artifact
func New(path *apath.Path, repoSlug, prefix, destination string, perm s3.ACL) *Artifact {
	return &Artifact{
		RepoSlug:    repoSlug,
		Path:        path,
		Prefix:      prefix,
		Destination: destination,
		Perm:        perm,

		UploadResult: &Result{},
	}
}

// ContentType makes it easier to find the perfect match
func (a *Artifact) ContentType() string {
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

// Reader makes an io.Reader out of the filepath
func (a *Artifact) Reader() (io.Reader, error) {
	f, err := os.Open(a.Path.Fullpath())
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Size reports the size of the artifact
func (a *Artifact) Size() (uint64, error) {
	fi, err := os.Stat(a.Path.Fullpath())
	if err != nil {
		return uint64(0), nil
	}

	return uint64(fi.Size()), nil
}

// FullDestination calculates the full remote destination path
func (a *Artifact) FullDestination() string {
	return strings.TrimLeft(filepath.Join(a.Prefix, a.Destination), "/")
}
