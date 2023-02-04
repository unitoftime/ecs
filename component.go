package ecs

import (
	"fmt"
	"sort"

	"reflect"
)

// TODO I think a lot of things can be cleaned up/optimized in this file

type CompId uint16

type Component interface {
	Write(*ArchEngine, ArchId, Id)
	Name() reflect.Type
}
// TODO -I could get rid of reflect if there ends up being some way to compile-time reflect on generics
type CompBox[T any] struct {
	Comp T
}
func C[T any](comp T) CompBox[T] {
	return CompBox[T]{comp}
}
func (c CompBox[T]) Write(engine *ArchEngine, archId ArchId, id Id) {
	WriteArch[T](engine, archId, id, c.Comp)
}
func (c CompBox[T]) Name() reflect.Type {
	return name(c.Comp)
}

func (c CompBox[T]) Get() T {
	return c.Comp
}

// Dynamic Component Registry
type DCR struct {
	archCounter ArchId
	compCounter CompId
	mapping map[reflect.Type]CompId // Contains the CompId for the component name
	archSet map[reflect.Type]map[ArchId]bool // Contains the set of ArchIds that have this component
	// componentStorageType map[string]any
	trie *node
}

func NewDCR() *DCR {
	r := &DCR{
		archCounter: 0,
		compCounter: 0,
		mapping: make(map[reflect.Type]CompId),
		archSet: make(map[reflect.Type]map[ArchId]bool),
	}
	r.trie = NewNode(r)
	return r
}

func (r *DCR) print() {
	fmt.Println("--- DCR ---")
	fmt.Println("archCounter", r.archCounter)
	fmt.Println("compCounter", r.compCounter)
	fmt.Println("-- mapping --")
	for name, compId := range r.mapping {
		fmt.Printf("name(%s) - compId(%d)\n", name, compId)
	}
	fmt.Println("-- archSet --")
	for name, set := range r.archSet {
		fmt.Printf("name(%s): archId: [ ", name)
		for archId := range set {
			fmt.Printf("%d ", archId)
		}
		fmt.Printf("]\n")
	}
}

func (r *DCR) NewArchId() ArchId {
	archId := r.archCounter
	r.archCounter++
	return archId
}

// 1. Map all components to their component Id
// 2. Sort all component ids so that we can index the prefix tree
// 3. Walk the prefix tree to find the ArchId
func (r *DCR) GetArchId(comp ...Component) ArchId {
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
		// n := name(c)
		n := c.Name()
		r.archSet[n][cur.archId] = true

		// r.archList[n] = append(r.archList[n], cur.archId)
		// TODO - sort these to improve my filter speed?
		// sort.Slice(r.archList[n], func(i, j int) bool {
		// 	return r.archList[n][i] < r.archList[n][j]
		// })
	}
	return cur.archId
}

// Registers a component to a component Id and returns the Id
// If already registered, just return the Id and don't make a new one
func (r *DCR) Register(comp Component) CompId {
	// n := name(comp)
	n := comp.Name()

	id, ok := r.mapping[n]
	if !ok {
		// add empty ArchList
		r.archSet[n] = make(map[ArchId]bool)

		r.compCounter++
		r.mapping[n] = r.compCounter
		return r.compCounter
	}

	return id
}

type node struct {
	archId ArchId
//	parent *node
	child []*node
}

func NewNode(r *DCR) *node {
	return &node{
		archId: r.NewArchId(),
		child: make([]*node, 0),
	}
}

func (n *node) Get(r *DCR, id CompId) *node {
	if id < CompId(len(n.child)) {
		if n.child[id] == nil {
			n.child[id] = NewNode(r)
		}
		return n.child[id]
	}

	// Expand the slice to hold all required children
	n.child = append(n.child, make([]*node, 1 + int(id) - len(n.child))...)
	if n.child[id] == nil {
		n.child[id] = NewNode(r)
	}
	return n.child[id]
}
