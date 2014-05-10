package path

import (
	"sync"
)

// PathSet is a set of paths and their behaviors
type PathSet struct {
	sync.Mutex
	paths map[*Path]int
}

// NewPathSet creates a new *PathSet
func NewPathSet() *PathSet {
	return &PathSet{
		paths: map[*Path]int{},
	}
}

// Add adds a path to the set
func (ps *PathSet) Add(p *Path) {
	ps.Lock()
	defer ps.Unlock()
	ps.paths[p] = 1
}

// All returns each path in the pathset
func (ps *PathSet) All() []*Path {
	ps.Lock()
	defer ps.Unlock()

	all := []*Path{}
	for key := range ps.paths {
		all = append(all, key)
	}

	return all
}
