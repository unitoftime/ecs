package ecs

import (
	"fmt"
)

// This is the identifier for entities in the world
//
//cod:struct
type Id uint32

type archetypeId uint32

// Provides generic storage for all archetypes
type archEngine struct {
	generation int

	lookup           []*lookupList // Indexed by archetypeId
	compSliceStorage []storage     // Indexed by componentId
	dcr              *componentRegistry
}

func newArchEngine() *archEngine {
	return &archEngine{
		generation: 1, // Start at 1 so that anyone with the default int value will always realize they are in the wrong generation

		lookup:           make([]*lookupList, 0, DefaultAllocation),
		compSliceStorage: make([]storage, maxComponentId+1),
		dcr:              newComponentRegistry(),
	}
}

func (e *archEngine) newArchetypeId(archMask archetypeMask, components []CompId) archetypeId {
	e.generation++ // Increment the generation

	archId := archetypeId(len(e.lookup))
	e.lookup = append(e.lookup,
		&lookupList{
			index:      newMap[Id, int](0),
			id:         make([]Id, 0, DefaultAllocation),
			holes:      make([]int, 0, DefaultAllocation),
			mask:       archMask,
			components: components,
		},
	)

	return archId
}

func (e *archEngine) getGeneration() int {
	return e.generation
}

func (e *archEngine) count(anything ...any) int {
	comps := make([]CompId, len(anything))
	for i, c := range anything {
		comps[i] = name(c)
	}

	archIds := make([]archetypeId, 0)
	archIds = e.FilterList(archIds, comps)

	total := 0
	for _, archId := range archIds {
		lookup := e.lookup[archId]
		if lookup == nil {
			panic(fmt.Sprintf("Couldnt find archId in archEngine lookup table: %d", archId))
		}

		// Each id represents an entity that holds the requested component(s)
		// Each hole represents a deleted entity that used to hold the requested component(s)
		total = total + len(lookup.id) - len(lookup.holes)
	}
	return total
}

func (e *archEngine) getArchetypeId(mask archetypeMask) archetypeId {
	return e.dcr.getArchetypeId(e, mask)
}

// Returns replaces archIds with a list of archids that match the compId list
func (e *archEngine) FilterList(archIds []archetypeId, comp []CompId) []archetypeId {
	// Idea 3: Loop through every registered archMask to see if it matches
	// Problem - Forces you to check every arch mask, even if the
	// The good side is that you dont need to deduplicate your list, and you dont need to allocate
	requiredArchMask := buildArchMaskFromId(comp...)

	archIds = archIds[:0]
	for archId := range e.dcr.revArchMask {
		if requiredArchMask.contains(e.dcr.revArchMask[archId]) {
			archIds = append(archIds, archetypeId(archId))
		}
	}
	return archIds

	//--------------------------------------------------------------------------------
	// Idea 2: Loop through every archMask that every componentId points to
	// // TODO: could I maybe do something more optimal with archetypeMask? Something like this could work.
	// requiredArchMask := buildArchMaskFromId(comp...)

	// archCount := make(map[archetypeId]struct{})

	// archIds = archIds[:0]
	// for _, compId := range comp {
	// 	for _, archId := range e.dcr.archSet[compId] {
	// 		archMask, ok := e.dcr.revArchMask[archId]
	// 		if !ok {
	// 			panic("AAA")
	// 			continue
	// 		} // TODO: This shouldn't happen?
	// 		if requiredArchMask.contains(archMask) {
	// 			archCount[archId] = struct{}{}
	// 		}
	// 	}
	// }

	// for archId := range archCount {
	// 	archIds = append(archIds, archId)
	// }
	// return archIds
}

func getStorage[T any](e *archEngine) *componentSliceStorage[T] {
	var val T
	n := name(val)
	return getStorageByCompId[T](e, n)
}

// Note: This will panic if the wrong compId doesn't match the generic type
func getStorageByCompId[T any](e *archEngine, compId CompId) *componentSliceStorage[T] {
	ss := e.compSliceStorage[compId]
	if ss == nil {
		ss = &componentSliceStorage[T]{
			slice: make(map[archetypeId]*componentSlice[T], DefaultAllocation),
		}
		e.compSliceStorage[compId] = ss
	}
	storage := ss.(*componentSliceStorage[T])

	return storage
}

func (e *archEngine) getOrAddLookupIndex(archId archetypeId, id Id) int {
	lookup := e.lookup[archId]

	// Check if we want to cleanup holes
	// TODO: This is a defragmentation operation. I'm not really sure how to compute heuristically that we should repack our slices. Too big it causes a stall, too small it causes unecessary repacks. maybe make it percentage based on holes per total entities. Maybe repack one at a time. Currently this should only trigger if we delete more than 1024 of the same archetype
	if len(lookup.holes) >= 1024 {
		e.CleanupHoles(archId)
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		// Because the Id hasn't been added to this arch, we need to add it
		index = lookup.addToEasiestHole(id)
	}
	return index
}

// Writes all of the components to the archetype
func (e *archEngine) write(archId archetypeId, id Id, comp ...Component) {
	// Add to lookup list
	index := e.getOrAddLookupIndex(archId, id)
	e.writeIndex(archId, id, index, comp...)
}

func (e *archEngine) writeIndex(archId archetypeId, id Id, index int, comp ...Component) {
	// Loop through all components and add them to individual component slices
	wd := W{
		engine: e,
		archId: archId,
		index:  index,
	}
	for i := range comp {
		comp[i].CompWrite(wd)
	}
}

// Allocates a slot for the supplied archId
func (e *archEngine) allocate(archId archetypeId, id Id) int {
	// Add to lookup list
	index := e.getOrAddLookupIndex(archId, id)

	// for compId registered to archId
	lookup := e.lookup[archId]
	for _, compId := range lookup.components {
		s := e.getStorage(compId)
		s.Allocate(archId, index)
	}
	return index
}

func (e *archEngine) getStorage(compId CompId) storage {
	ss := e.compSliceStorage[compId]
	if ss == nil {
		ss = newComponentStorage(compId)
		e.compSliceStorage[compId] = ss
	}
	return ss
}

func writeArch[T any](e *archEngine, archId archetypeId, index int, store *componentSliceStorage[T], val T) {
	// Get the underlying Archetype's componentSlice
	cSlice, ok := store.slice[archId]
	if !ok {
		cSlice = &componentSlice[T]{
			comp: make([]T, 0, DefaultAllocation),
		}
		store.slice[archId] = cSlice
	}

	cSlice.Write(index, val)
}

func readArch[T any](e *archEngine, archId archetypeId, id Id) (T, bool) {
	var ret T
	lookup := e.lookup[archId]
	if lookup == nil {
		return ret, false // TODO: when could this possibly happen?
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		return ret, false
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss := e.compSliceStorage[n]
	if ss == nil {
		return ret, false
	}

	// fmt.Printf("componentSliceStorage[T] type: %s != %s", name(ss), name(ret))
	storage, ok := ss.(*componentSliceStorage[T])
	if !ok {
		panic(fmt.Sprintf("Wrong componentSliceStorage[T] type: %d != %d", name(ss), name(ret)))
	}

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		return ret, false
	}

	return cSlice.comp[index], true
}

func readPtrArch[T any](e *archEngine, archId archetypeId, id Id) *T {
	var ret T
	lookup := e.lookup[archId]
	if lookup == nil {
		return nil
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		return nil
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss := e.compSliceStorage[n]
	if ss == nil {
		return nil
	}

	storage, ok := ss.(*componentSliceStorage[T])
	if !ok {
		panic(fmt.Sprintf("Wrong componentSliceStorage[T] type: %d != %d", name(ss), name(ret)))
	}

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		return nil
	}

	return &cSlice.comp[index]
}

// Returns the archetypeId of where the entity ends up
func (e *archEngine) rewriteArch(archId archetypeId, id Id, comp ...Component) archetypeId {
	// Calculate the new mask based on the bitwise or of the old and added masks
	lookup := e.lookup[archId]
	oldMask := lookup.mask
	addMask := buildArchMask(comp...)
	newMask := oldMask.bitwiseOr(addMask)

	if oldMask == newMask {
		// Case 1: Archetype stays the same.
		// This means that we only need to write the newly added components because we wont be moving the base entity data
		e.write(archId, id, comp...)
		return archId
	} else {
		// 1. Move Archetype Data
		newArchId, newIndex := e.moveArchetype(archId, newMask, id)

		// 2. Write new componts to new archetype/index location
		e.writeIndex(newArchId, id, newIndex, comp...)

		return newArchId
	}
}

// Moves an entity from one archetype to another, copying all of the data from the old archetype to the new one
func (e *archEngine) moveArchetype(oldArchId archetypeId, newMask archetypeMask, id Id) (archetypeId, int) {
	newArchId := e.dcr.getArchetypeId(e, newMask)

	newIndex := e.allocate(newArchId, id)

	oldLookup := e.lookup[oldArchId]
	oldIndex, ok := oldLookup.index.Get(id)
	if !ok {
		panic("bug: id missing from lookup list")
	}

	for _, compId := range oldLookup.components {
		store := e.compSliceStorage[compId]
		store.moveArchetype(oldArchId, oldIndex, newArchId, newIndex)
	}

	e.TagForDeletion(oldArchId, id)

	return newArchId, newIndex
}

// Moves an entity from one archetype to another, copying all of the data required by the new archetype
func (e *archEngine) moveArchetypeDown(oldArchId archetypeId, newMask archetypeMask, id Id) archetypeId {
	newArchId := e.dcr.getArchetypeId(e, newMask)

	newIndex := e.allocate(newArchId, id)

	oldLookup := e.lookup[oldArchId]
	oldIndex, ok := oldLookup.index.Get(id)
	if !ok {
		panic("bug: id missing from lookup list")
	}

	newLookup := e.lookup[newArchId]
	for _, compId := range newLookup.components {
		store := e.compSliceStorage[compId]
		store.moveArchetype(oldArchId, oldIndex, newArchId, newIndex)
	}

	e.TagForDeletion(oldArchId, id)

	return newArchId
}

func (e *archEngine) ReadEntity(archId archetypeId, id Id) *Entity {
	lookup := e.lookup[archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		panic("Archetype doesn't contain ID")
	}

	ent := NewEntity()
	for n := range e.compSliceStorage {
		if e.compSliceStorage[n] != nil {
			e.compSliceStorage[n].ReadToEntity(ent, archId, index)
		}
	}
	return ent
}

func (e *archEngine) ReadRawEntity(archId archetypeId, id Id) *RawEntity {
	lookup := e.lookup[archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		panic("Archetype doesn't contain ID")
	}

	ent := NewRawEntity()
	for n := range e.compSliceStorage {
		if e.compSliceStorage[n] != nil {
			e.compSliceStorage[n].ReadToRawEntity(ent, archId, index)
		}
	}
	return ent
}

// This creates a "hole" in the archetype at the specified Id
// Once we get enough holes, we can re-pack the entire slice
// TODO - How many holes before we repack? How many holes to pack at a time?
func (e *archEngine) TagForDeletion(archId archetypeId, id Id) {
	lookup := e.lookup[archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	index, ok := lookup.index.Get(id)
	if !ok {
		panic("Archetype doesn't contain ID")
	}

	// This indicates that the index needs to be cleaned up and should be skipped in any list processing
	lookup.id[index] = InvalidEntity
	lookup.index.Delete(id)

	// This is used to track the current list of indices that need to be cleaned
	lookup.holes = append(lookup.holes, index)
}

func (e *archEngine) CleanupHoles(archId archetypeId) {
	lookup := e.lookup[archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	for _, index := range lookup.holes {
		// Pop all holes off the end of the archetype
		for {
			lastIndex := len(lookup.id) - 1
			if lastIndex < 0 {
				break
			} // Break if the index we are trying to pop off is -1
			lastId := lookup.id[lastIndex]
			if lastId == InvalidEntity {
				// If the last id is a hole, then slice it off
				lookup.id = lookup.id[:lastIndex]
				for n := range e.compSliceStorage {
					if e.compSliceStorage[n] != nil {
						e.compSliceStorage[n].Delete(archId, lastIndex)
					}
				}

				continue // Try again
			}

			break
		}

		// Check bounds because we may have popped past our original index
		if index >= len(lookup.id) {
			continue
		}

		// Swap lastIndex (which is not a hole) with index (which is a hole)
		lastIndex := len(lookup.id) - 1
		lastId := lookup.id[lastIndex]
		if lastId == InvalidEntity {
			panic("Bug: This shouldn't happen")
		}

		lookup.id[index] = lastId
		lookup.id = lookup.id[:lastIndex]
		lookup.index.Put(lastId, index)
		for n := range e.compSliceStorage {
			if e.compSliceStorage[n] != nil {
				e.compSliceStorage[n].Delete(archId, index)
			}
		}
	}

	// Clear holes slice
	lookup.holes = lookup.holes[:0]
}
