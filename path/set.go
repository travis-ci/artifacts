package path

import (
	"fmt"
	"sync"
)

// Set is a set of paths and their behaviors
type Set struct {
	sync.Mutex
	paths map[string]*Path
}

// NewSet creates a new *Set
func NewSet() *Set {
	return &Set{
		paths: map[string]*Path{},
	}
}

// Add adds a path to the set
func (ps *Set) Add(p *Path) {
	ps.Lock()
	defer ps.Unlock()
	ps.paths[fmt.Sprintf("%#v", p)] = p
}

// All returns each path in the pathset
func (ps *Set) All() []*Path {
	ps.Lock()
	defer ps.Unlock()

	all := []*Path{}
	for _, p := range ps.paths {
		all = append(all, p)
	}

	return all
}
