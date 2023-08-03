package ecs

import (
	"fmt"
	"sort"
)

type componentId uint16

type Component interface {
	write(*archEngine, archetypeId, int)
	id() componentId
}

// This type is used to box a component with all of its type info so that it implements the component interface. I would like to get rid of this and simplify the APIs
type Box[T any] struct {
	Comp   T
	compId componentId
	storage *componentSliceStorage[T]
}

// Createst the boxed component type
func C[T any](comp T) Box[T] {
	return Box[T]{
		Comp:   comp,
		compId: name(comp),
	}
}
func (c Box[T]) write(engine *archEngine, archId archetypeId, index int) {
	if c.storage == nil {
		c.storage = getStorageByCompId[T](engine, c.id())
	}
	writeArch[T](engine, archId, index, c.storage, c.Comp)
}
func (c Box[T]) id() componentId {
	if c.compId == invalidComponentId {
		c.compId = name(c.Comp)
	}
	return c.compId
}

func (c Box[T]) Get() T {
	return c.Comp
}

// TODO: You should move to this (ie archetype graph (or bitmask?). maintain the current archetype node, then traverse to nodes (and add new ones) based on which components are added): https://ajmmertens.medium.com/building-an-ecs-2-archetypes-and-vectorization-fe21690805f9
// Dynamic component Registry
type componentRegistry struct {
	archCounter archetypeId
	archSet     map[componentId]map[archetypeId]bool // Contains the set of archetypeIds that have this component
	trie        *node
	generation  int
}

func newComponentRegistry() *componentRegistry {
	r := &componentRegistry{
		archCounter: 0,
		archSet:     make(map[componentId]map[archetypeId]bool),
		generation:  1, // Start at 1 so that anyone with the default int value will always realize they are in the wrong generation
	}
	r.trie = newNode(r)
	return r
}

func (r *componentRegistry) print() {
	fmt.Println("--- componentRegistry ---")
	fmt.Println("archCounter", r.archCounter)
	fmt.Println("-- archSet --")
	for name, set := range r.archSet {
		fmt.Printf("name(%d): archId: [ ", name)
		for archId := range set {
			fmt.Printf("%d ", archId)
		}
		fmt.Printf("]\n")
	}
}

func (r *componentRegistry) NewarchetypeId() archetypeId {
	r.generation++ // Increment the generation
	archId := r.archCounter
	r.archCounter++
	return archId
}

// 1. Map all components to their component Id
// 2. Sort all component ids so that we can index the prefix tree
// 3. Walk the prefix tree to find the archetypeId
func (r *componentRegistry) GetarchetypeId(comp ...componentId) archetypeId {
	list := make([]componentId, len(comp))
	for i, compId := range comp {
		r.Register(compId)
		list[i] = compId
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})

	cur := r.trie
	for _, idx := range list {
		cur = cur.Get(r, idx)
	}

	// Add this archetypeId to every component's archList
	for _, compId := range comp {
		if !r.archSet[compId][cur.archId] {
			r.archSet[compId][cur.archId] = true

			// If this was the first time we've associated this archetype to this component, then we need to bump the generation, so that all views get regenerated based on this update. This could maybe be moved to somewhere else.
			r.generation++
		}
	}
	return cur.archId
}

// Registers the component
func (r *componentRegistry) Register(compId componentId) {
	_, ok := r.archSet[compId]
	if !ok {
		r.archSet[compId] = make(map[archetypeId]bool)
	}
}

type node struct {
	archId archetypeId
	child  []*node
}

func newNode(r *componentRegistry) *node {
	return &node{
		archId: r.NewarchetypeId(),
		child:  make([]*node, 0),
	}
}

func (n *node) Get(r *componentRegistry, id componentId) *node {
	if id < componentId(len(n.child)) {
		if n.child[id] == nil {
			n.child[id] = newNode(r)
		}
		return n.child[id]
	}

	// Expand the slice to hold all required children
	n.child = append(n.child, make([]*node, 1+int(id)-len(n.child))...)
	if n.child[id] == nil {
		n.child[id] = newNode(r)
	}
	return n.child[id]
}
