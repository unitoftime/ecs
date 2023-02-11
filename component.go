package ecs

import (
	"fmt"
	"sort"
)

type CompId uint16

type Component interface {
	Write(*archEngine, ArchId, Id)
	Name() CompId
}
// This type is used to box a component with all of its type info so that it implements the Component interface. I would like to get rid of this and simplify the APIs
type Box[T any] struct {
	Comp T
	compId CompId
}
func C[T any](comp T) Box[T] {
	return Box[T]{
		Comp: comp,
		compId: name(comp),
	}
}
func (c Box[T]) Write(engine *archEngine, archId ArchId, id Id) {
	writeArch[T](engine, archId, id, c.Comp)
}
func (c Box[T]) Name() CompId {
	if c.compId == invalidComponentId {
		c.compId = name(c.Comp)
	}
	return c.compId
}

func (c Box[T]) Get() T {
	return c.Comp
}

// Dynamic Component Registry
type componentRegistry struct {
	archCounter ArchId
	compCounter CompId
	archSet map[CompId]map[ArchId]bool // Contains the set of ArchIds that have this component
	trie *node
	generation int
}

func newComponentRegistry() *componentRegistry {
	r := &componentRegistry{
		archCounter: 0,
		compCounter: 0,
		archSet: make(map[CompId]map[ArchId]bool),
		generation: 1, // Start at 1 so that anyone with the default int value will always realize they are in the wrong generation
	}
	r.trie = newNode(r)
	return r
}

func (r *componentRegistry) print() {
	fmt.Println("--- componentRegistry ---")
	fmt.Println("archCounter", r.archCounter)
	fmt.Println("compCounter", r.compCounter)
	fmt.Println("-- archSet --")
	for name, set := range r.archSet {
		fmt.Printf("name(%d): archId: [ ", name)
		for archId := range set {
			fmt.Printf("%d ", archId)
		}
		fmt.Printf("]\n")
	}
}

func (r *componentRegistry) NewArchId() ArchId {
	r.generation++ // Increment the generation
	archId := r.archCounter
	r.archCounter++
	return archId
}

// 1. Map all components to their component Id
// 2. Sort all component ids so that we can index the prefix tree
// 3. Walk the prefix tree to find the ArchId
func (r *componentRegistry) GetArchId(comp ...Component) ArchId {
	list := make([]CompId, len(comp))
	for i := range comp {
		list[i] = r.Register(comp[i])
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})

	cur := r.trie
	for _, idx := range list {
		cur = cur.Get(r, idx)
	}

	// Add this ArchId to every component's archList
	for _, c := range comp {
		n := c.Name()
		r.archSet[n][cur.archId] = true
	}
	return cur.archId
}

// Registers a component to a component Id and returns the Id
// If already registered, just return the Id and don't make a new one
func (r *componentRegistry) Register(comp Component) CompId {
	compId := comp.Name()

	_, ok := r.archSet[compId]
	if !ok {
		r.archSet[compId] = make(map[ArchId]bool)
	}

	return compId
}

type node struct {
	archId ArchId
	child []*node
}

func newNode(r *componentRegistry) *node {
	return &node{
		archId: r.NewArchId(),
		child: make([]*node, 0),
	}
}

func (n *node) Get(r *componentRegistry, id CompId) *node {
	if id < CompId(len(n.child)) {
		if n.child[id] == nil {
			n.child[id] = newNode(r)
		}
		return n.child[id]
	}

	// Expand the slice to hold all required children
	n.child = append(n.child, make([]*node, 1 + int(id) - len(n.child))...)
	if n.child[id] == nil {
		n.child[id] = newNode(r)
	}
	return n.child[id]
}
