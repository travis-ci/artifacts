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
)

const (
	defaultCtype = "application/octet-stream"
)

// Artifact is the thing that gets uploaded or whatever
type Artifact struct {
	RepoSlug    string
	BuildNumber string
	BuildID     string
	JobNumber   string
	JobID       string

	Source string
	Dest   string
	Prefix string
	Perm   string

	UploadResult *Result
}

// New creates a new *Artifact
func New(prefix, source, dest string, opts *Options) *Artifact {
	return &Artifact{
		Prefix: prefix,
		Source: source,
		Dest:   dest,

		RepoSlug:    opts.RepoSlug,
		BuildNumber: opts.BuildNumber,
		BuildID:     opts.BuildID,
		JobNumber:   opts.JobNumber,
		JobID:       opts.JobID,
		Perm:        opts.Perm,

		UploadResult: &Result{},
	}
}

// ContentType makes it easier to find the perfect match
func (a *Artifact) ContentType() string {
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
	if err != nil && err != io.EOF {
		return defaultCtype
	}

	return http.DetectContentType(buf.Bytes())
}

// Reader makes an io.Reader out of the filepath
func (a *Artifact) Reader() (io.Reader, error) {
	f, err := os.Open(a.Source)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Size reports the size of the artifact
func (a *Artifact) Size() (uint64, error) {
	fi, err := os.Stat(a.Source)
	if err != nil {
		return uint64(0), nil
	}

	return uint64(fi.Size()), nil
}

// FullDest calculates the full remote destination path
func (a *Artifact) FullDest() string {
	return strings.TrimLeft(filepath.Join(a.Prefix, a.Dest), "/")
}
