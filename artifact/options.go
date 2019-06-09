package artifact

// Options encapsulates stuff specific to artifacts.  Ugh.
type Options struct {
	RepoSlug    string
	BuildNumber string
	BuildID     string
	JobNumber   string
	JobID       string
	Perm        string
}
