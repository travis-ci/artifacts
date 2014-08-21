package path

import (
	"fmt"
	"sync"
)

// PathSet is a set of paths and their behaviors
type PathSet struct {
	sync.Mutex
	paths map[string]*Path
}

// NewPathSet creates a new *PathSet
func NewPathSet() *PathSet {
	return &PathSet{
		paths: map[string]*Path{},
	}
}

// Add adds a path to the set
func (ps *PathSet) Add(p *Path) {
	ps.Lock()
	defer ps.Unlock()
	ps.paths[fmt.Sprintf("%#v", p)] = p
}

// All returns each path in the pathset
func (ps *PathSet) All() []*Path {
	ps.Lock()
	defer ps.Unlock()

	all := []*Path{}
	for _, p := range ps.paths {
		all = append(all, p)
	}

	return all
}
