package path

import (
	"os"
	"path/filepath"
	"strings"
)

// Path is path-like.  Bonkers.
type Path struct {
	Root string
	From string
	To   string
}

// New makes a new *Path.  Crazy!
func New(root, from, to string) *Path {
	return &Path{
		Root: root,
		From: from,
		To:   to,
	}
}

// Fullpath returns the full file/dir path
func (p *Path) Fullpath() string {
	if p.IsAbs() || strings.HasPrefix(p.From, "/") {
		return p.From
	}

	return filepath.Join(p.Root, p.From)
}

// IsDir tells if the path is a directory!
func (p *Path) IsDir() bool {
	fi, err := os.Stat(p.From)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// IsExists tells if the path exists locally
func (p *Path) IsExists() bool {
	_, err := os.Stat(p.From)
	return err == nil
}

// IsAbs tells if the path is absolute!
func (p *Path) IsAbs() bool {
	asRelpath := filepath.Join(p.Root, p.From)
	_, err := os.Stat(asRelpath)
	if err == nil {
		return false
	}

	return p.IsExists()
}
