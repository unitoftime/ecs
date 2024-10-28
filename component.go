package ecs

import (
	"fmt"
	"sync"
)

type storageBuilder interface {
	build() storage
}
type storageBuilderImp[T any] struct {
}

func (s storageBuilderImp[T]) build() storage {
	return &componentSliceStorage[T]{
		slice: make(map[archetypeId]*componentSlice[T], DefaultAllocation),
	}
}

func nameTyped[T any](comp T) componentId {
	compId := name(comp)
	registerComponentStorage[T](compId)
	return compId
}

var componentStorageLookupMut sync.RWMutex
var componentStorageLookup = make(map[componentId]storageBuilder)

func registerComponentStorage[T any](compId componentId) {
	componentStorageLookupMut.Lock()
	_, ok := componentStorageLookup[compId]
	if !ok {
		componentStorageLookup[compId] = storageBuilderImp[T]{}
	}
	componentStorageLookupMut.Unlock()
}

func newComponentStorage(c componentId) storage {
	componentStorageLookupMut.RLock()
	s, ok := componentStorageLookup[c]
	if !ok {
		panic(fmt.Sprintf("tried to build component storage with unregistered componentId: %d", c))
	}

	componentStorageLookupMut.RUnlock()
	return s.build()
}

type componentId uint16

type Component interface {
	SetOther(Component) // TODO: ???
	Clone() Component   // TODO: ???
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
	compId := nameTyped[T](comp)
	return Box[T]{
		Comp:   comp,
		compId: compId,
	}
}
func (c Box[T]) write(engine *archEngine, archId archetypeId, index int) {
	store := getStorageByCompId[T](engine, c.id())
	writeArch[T](engine, archId, index, store, c.Comp)
}
func (c Box[T]) writeVal(engine *archEngine, archId archetypeId, index int, val T) {
	store := getStorageByCompId[T](engine, c.id())
	writeArch[T](engine, archId, index, store, val)
}
func (c Box[T]) getPtr(engine *archEngine, archId archetypeId, index int) *T {
	store := getStorageByCompId[T](engine, c.id())
	slice := store.slice[archId]
	return &slice.comp[index]
}

func (c Box[T]) id() componentId {
	if c.compId == invalidComponentId {
		c.compId = name(c.Comp)
	}
	return c.compId
}

func (c Box[T]) SetOther(other Component) {
	o := other.(Box[T])
	o.Comp = c.Comp
}

func (c Box[T]) Get() T {
	return c.Comp
}

func (c Box[T]) Clone() Component {
	return c
}

func (c Box[T]) UnbundleVal(bun *Bundler, val T) {
	c.Comp = val
	c.Unbundle(bun)
}

func (c Box[T]) Unbundle(bun *Bundler) {
	compId := c.compId
	val := c.Comp
	bun.archMask.addComponent(compId)
	bun.Set[compId] = true
	if bun.Components[compId] == nil {
		// Note: We need a pointer so that we dont do an allocation every time we set it
		c2 := c // Note: make a copy, so the bundle doesn't contain a pointer to the original
		bun.Components[compId] = &c2
	} else {
		rwComp := bun.Components[compId].(*Box[T])
		rwComp.Comp = val
	}

	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
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

func (m *archetypeMask) addComponent(compId componentId) {
	// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
	idx := compId / 64
	offset := compId - (64 * idx)
	m[idx] |= (1 << offset)
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

// Checks to see if a mask m contains the supplied componentId
// Returns true if the bit location in that mask is set, else returns false
func (m archetypeMask) hasComponent(compId componentId) bool {
	// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
	idx := compId / 64
	offset := compId - (64 * idx)
	return (m[idx] & (1 << offset)) != 0
}

// Generates and returns a list of every componentId that this archetype contains
func (m archetypeMask) getComponentList() []componentId {
	ret := make([]componentId, 0)
	for compId := componentId(0); compId <= maxComponentId; compId++ {
		if m.hasComponent(compId) {
			ret = append(ret, compId)
		}
	}
	return ret
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

// func (r *componentRegistry) getArchetypeId(engine *archEngine, comps ...Component) archetypeId {
// 	mask := buildArchMask(comps...)
// 	archId, ok := r.archMask[mask]
// 	if !ok {
// 		componentIds := make([]componentId, 0)
// 		for i := range comps {
// 			componentIds = append(componentIds, comps[i].id())
// 		}

// 		archId = engine.newArchetypeId(mask, componentIds)
// 		r.archMask[mask] = archId

// 		if int(archId) != len(r.revArchMask) {
// 			panic(fmt.Sprintf("ecs: archId must increment. Expected: %d, Got: %d", len(r.revArchMask), archId))
// 		}
// 		r.revArchMask = append(r.revArchMask, mask)

// 		// Add this archetypeId to every component's archList
// 		for _, comp := range comps {
// 			compId := comp.id()
// 			r.archSet[compId] = append(r.archSet[compId], archId)
// 		}
// 	}
// 	return archId
// }

func (r *componentRegistry) getArchetypeIdFromMask(engine *archEngine, mask archetypeMask) archetypeId {
	archId, ok := r.archMask[mask]
	if !ok {
		componentIds := mask.getComponentList()
		archId = engine.newArchetypeId(mask, componentIds)
		r.archMask[mask] = archId

		if int(archId) != len(r.revArchMask) {
			panic(fmt.Sprintf("ecs: archId must increment. Expected: %d, Got: %d", len(r.revArchMask), archId))
		}
		r.revArchMask = append(r.revArchMask, mask)

		// Add this archetypeId to every component's archList
		for _, compId := range componentIds {
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
