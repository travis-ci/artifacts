package path

import (
	"path/filepath"
	"strings"
)

// Path is path-like.  Bonkers.
type Path struct {
	Root string
	From string
	To   string
}

// NewPath makes a new *Path.  Crazy!
func NewPath(root, from, to string) *Path {
	return &Path{
		Root: root,
		From: from,
		To:   to,
	}
}

// Fullpath returns the full file/dir path
func (p *Path) Fullpath() string {
	if strings.HasPrefix(p.From, "/") {
		return p.From
	}

	return filepath.Join(p.Root, p.From)
}

// IsDir tells if the path is a directory!
func (p *Path) IsDir() bool {
	return false
}
