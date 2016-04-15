package artifact

import (
	"github.com/goamz/goamz/s3"
)

// Options encapsulates stuff specific to artifacts.  Ugh.
type Options struct {
	RepoSlug    string
	BuildNumber string
	BuildID     string
	JobNumber   string
	JobID       string
	Perm        s3.ACL
}
