package ecs

import (
	"fmt"
	// "sort"
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
}

// Createst the boxed component type
func C[T any](comp T) Box[T] {
	return Box[T]{
		Comp:   comp,
		compId: name(comp),
	}
}
func (c Box[T]) write(engine *archEngine, archId archetypeId, index int) {
	store := getStorageByCompId[T](engine, c.id())
	writeArch[T](engine, archId, index, store, c.Comp)
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


// Note: you can increase max component size by increasing maxComponentId and archetypeMask
// TODO: I should have some kind of panic if you go over maximum component size
const maxComponentId = 255
// Supports maximum 256 unique component types
type archetypeMask [4]uint64 // TODO: can/should I make this configurable?
func buildArchMask(comps ...Component) archetypeMask {
	var mask archetypeMask
	for _, comp := range comps {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		c := comp.id()
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1<<offset)
	}
	return mask
}

// Performs a bitwise or on the base mask `m` with the added mask `a`
func (m archetypeMask) bitwiseOr(a archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] | a[i]
	}
	return m
}

// TODO: You should move to this (ie archetype graph (or bitmask?). maintain the current archetype node, then traverse to nodes (and add new ones) based on which components are added): https://ajmmertens.medium.com/building-an-ecs-2-archetypes-and-vectorization-fe21690805f9
// Dynamic component Registry
type componentRegistry struct {
	archSet     [][]archetypeId // Contains the set of archetypeIds that have this component
	archMask    map[archetypeMask]archetypeId // Contains a mapping of archetype bitmasks to archetypeIds
}

func newComponentRegistry() *componentRegistry {
	r := &componentRegistry{
		archSet:     make([][]archetypeId, maxComponentId + 1), // TODO: hardcoded to max component
		archMask:    make(map[archetypeMask]archetypeId),
	}
	return r
}

func (r *componentRegistry) print() {
	fmt.Println("--- componentRegistry ---")
	fmt.Println("-- archSet --")
	for name, set := range r.archSet {
		fmt.Printf("name(%d): archId: [ ", name)
		for archId := range set {
			fmt.Printf("%d ", archId)
		}
		fmt.Printf("]\n")
	}
}

func (r *componentRegistry) getArchetypeId(engine *archEngine, comps ...Component) archetypeId {
	mask := buildArchMask(comps...)
	archId, ok := r.archMask[mask]
	if !ok {
		archId = engine.newArchetypeId(mask)
		r.archMask[mask] = archId

		// Add this archetypeId to every component's archList
		for _, comp := range comps {
			compId := comp.id()
			r.archSet[compId] = append(r.archSet[compId], archId)
		}
	}
	return archId
}
