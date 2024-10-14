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

var blankArchMask archetypeMask

// Supports maximum 256 unique component types
type archetypeMask [4]uint64 // TODO: can/should I make this configurable?
func buildArchMask(comps ...Component) archetypeMask {
	var mask archetypeMask
	for _, comp := range comps {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		c := comp.id()
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}
func buildArchMaskFromAny(comps ...any) archetypeMask {
	var mask archetypeMask
	for _, comp := range comps {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		c := name(comp)
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}
func buildArchMaskFromId(compIds ...componentId) archetypeMask {
	var mask archetypeMask
	for _, c := range compIds {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}

// Performs a bitwise OR on the base mask `m` with the added mask `a`
func (m archetypeMask) bitwiseOr(a archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] | a[i]
	}
	return m
}

// Performs a bitwise AND on the base mask `m` with the added mask `a`
func (m archetypeMask) bitwiseAnd(a archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] & a[i]
	}
	return m
}

// Checks to ensure archetype m contains archetype a
// Returns true if every bit in m is also set in a
// Returns false if at least one set bit in m is not set in a
func (m archetypeMask) contains(a archetypeMask) bool {
	// Logic: Bitwise AND on every segment, if the 'check' result doesn't match m[i] for that segment
	// then we know there was a bit in a[i] that was not set
	var check uint64
	for i := range m {
		check = m[i] & a[i]
		if check != m[i] {
			return false
		}
	}
	return true
}

// TODO: You should move to this (ie archetype graph (or bitmask?). maintain the current archetype node, then traverse to nodes (and add new ones) based on which components are added): https://ajmmertens.medium.com/building-an-ecs-2-archetypes-and-vectorization-fe21690805f9
// Dynamic component Registry
type componentRegistry struct {
	archSet  [][]archetypeId               // Contains the set of archetypeIds that have this component
	archMask map[archetypeMask]archetypeId // Contains a mapping of archetype bitmasks to archetypeIds

	revArchMask []archetypeMask // Contains the reverse mapping of archetypeIds to archetype masks. Indexed by archetypeId
}

func newComponentRegistry() *componentRegistry {
	r := &componentRegistry{
		archSet:     make([][]archetypeId, maxComponentId+1), // TODO: hardcoded to max component
		archMask:    make(map[archetypeMask]archetypeId),
		revArchMask: make([]archetypeMask, 0),
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

		if int(archId) != len(r.revArchMask) {
			panic(fmt.Sprintf("ecs: archId must increment. Expected: %d, Got: %d", len(r.revArchMask), archId))
		}
		r.revArchMask = append(r.revArchMask, mask)

		// Add this archetypeId to every component's archList
		for _, comp := range comps {
			compId := comp.id()
			r.archSet[compId] = append(r.archSet[compId], archId)
		}
	}
	return archId
}

// This is mostly for the without filter
func (r *componentRegistry) archIdOverlapsMask(archId archetypeId, compArchMask archetypeMask) bool {
	archMaskToCheck := r.revArchMask[archId]

	resultArchMask := archMaskToCheck.bitwiseAnd(compArchMask)
	if resultArchMask != blankArchMask {
		// If the resulting arch mask is nonzero, it means that both the component mask and the base mask had the same bit set, which means the arch had one of the components
		return true
	}
	return false
}
