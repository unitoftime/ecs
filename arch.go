package ecs

import (
	"fmt"
)

// This is the identifier for entities in the world
//
//cod:struct
type Id uint32

type archetypeId uint32

type entLoc struct {
	archId archetypeId
	index  uint32
}

// Provides generic storage for all archetypes
type archEngine struct {
	generation int

	lookup      []*lookupList // Indexed by archetypeId
	compStorage []storage     // Indexed by componentId
	dcr         *componentRegistry
}

func newArchEngine() *archEngine {
	return &archEngine{
		generation: 1, // Start at 1 so that anyone with the default int value will always realize they are in the wrong generation

		lookup:      make([]*lookupList, 0, DefaultAllocation),
		compStorage: make([]storage, maxComponentId+1),
		dcr:         newComponentRegistry(),
	}
}

func (e *archEngine) print() {
	// fmt.Printf("%+v\n", *e)
	for i := range e.lookup {
		fmt.Printf("  id: %+v\n", e.lookup[i].id)
		fmt.Printf("  holes: %+v\n", e.lookup[i].holes)
		fmt.Printf("  mask: %+v\n", e.lookup[i].mask)
		fmt.Printf("  components: %+v\n", e.lookup[i].components)
		fmt.Printf("--------------------------------------------------------------------------------\n")
	}

	for i := range e.compStorage {
		fmt.Printf("css: %d: %+v\n", i, e.compStorage[i])
	}
}

func (e *archEngine) newArchetypeId(archMask archetypeMask, components []CompId) archetypeId {
	e.generation++ // Increment the generation

	archId := archetypeId(len(e.lookup))
	e.lookup = append(e.lookup,
		&lookupList{
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

func getStorage[T any](e *archEngine) *componentStorage[T] {
	var val T
	n := name(val)
	return getStorageByCompId[T](e, n)
}

// Note: This will panic if the wrong compId doesn't match the generic type
func getStorageByCompId[T any](e *archEngine, compId CompId) *componentStorage[T] {
	ss := e.compStorage[compId]
	if ss == nil {
		ss = &componentStorage[T]{
			slice: newMap[archetypeId, *componentList[T]](DefaultAllocation),
		}
		e.compStorage[compId] = ss
	}
	storage := ss.(*componentStorage[T])

	return storage
}

func (e *archEngine) getOrAddLookupIndex(archId archetypeId, id Id) int {
	lookup := e.lookup[archId]

	// TODO: Removed this here. Need to relocate this somewhere better
	// // Check if we want to cleanup holes
	// // TODO: This is a defragmentation operation. I'm not really sure how to compute heuristically that we should repack our slices. Too big it causes a stall, too small it causes unecessary repacks. maybe make it percentage based on holes per total entities. Maybe repack one at a time. Currently this should only trigger if we delete more than 1024 of the same archetype
	// if len(lookup.holes) >= 1024 {
	// 	e.CleanupHoles(archId)
	// }

	// index, ok := lookup.index.Get(id)
	// if !ok {
	// 	// Because the Id hasn't been added to this arch, we need to add it
	// 	index = lookup.addToEasiestHole(id)
	// }
	index := lookup.addToEasiestHole(id)
	return index
}

// Writes all of the components to the archetype
func (e *archEngine) write(loc entLoc, id Id, comp ...Component) {
	e.writeIndex(loc, id, comp...)
}

// Writes all of the components to the archetype.
// Internally requires that the id is not added to the archetype
func (e *archEngine) spawn(archId archetypeId, id Id, comp ...Component) int {
	lookup := e.lookup[archId]
	// TODO: Doesn't cleanup holes?
	index := lookup.addToEasiestHole(id)
	loc := entLoc{archId, uint32(index)}
	e.writeIndex(loc, id, comp...)
	return index
}

func (e *archEngine) writeIndex(loc entLoc, id Id, comp ...Component) {
	// Loop through all components and add them to individual component slices
	wd := W{
		engine: e,
		archId: loc.archId,
		index:  int(loc.index),
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
	ss := e.compStorage[compId]
	if ss == nil {
		ss = newComponentStorage(compId)
		e.compStorage[compId] = ss
	}
	return ss
}

func writeArch[T any](e *archEngine, archId archetypeId, index int, store *componentStorage[T], val T) {
	cSlice := store.GetSlice(archId)
	cSlice.Write(index, val)
}

func readArch[T any](e *archEngine, loc entLoc, id Id) (T, bool) {
	var ret T
	lookup := e.lookup[loc.archId]
	if lookup == nil {
		return ret, false // TODO: when could this possibly happen?
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss := e.compStorage[n]
	if ss == nil {
		return ret, false
	}

	storage, ok := ss.(*componentStorage[T])
	if !ok {
		panic(fmt.Sprintf("Wrong componentSliceStorage[T] type: %d != %d", name(ss), name(ret)))
	}

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice.Get(loc.archId)
	if !ok {
		return ret, false
	}

	return cSlice.comp[loc.index], true
}

func readPtrArch[T any](e *archEngine, loc entLoc, id Id) *T {
	var ret T
	lookup := e.lookup[loc.archId]
	if lookup == nil {
		return nil
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss := e.compStorage[n]
	if ss == nil {
		return nil
	}

	storage, ok := ss.(*componentStorage[T])
	if !ok {
		panic(fmt.Sprintf("Wrong componentSliceStorage[T] type: %d != %d", name(ss), name(ret)))
	}

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice.Get(loc.archId)
	if !ok {
		return nil
	}

	return &cSlice.comp[loc.index]
}

// Returns the archetypeId of where the entity ends up
func (e *archEngine) rewriteArch(loc entLoc, id Id, comp ...Component) entLoc {
	// Calculate the new mask based on the bitwise or of the old and added masks
	lookup := e.lookup[loc.archId]
	oldMask := lookup.mask
	addMask := buildArchMask(comp...)
	newMask := oldMask.bitwiseOr(addMask)

	if oldMask == newMask {
		// Case 1: Archetype and index stays the same.
		// This means that we only need to write the newly added components because we wont be moving the base entity data
		e.write(loc, id, comp...)
		return loc
	} else {
		// 1. Move Archetype Data
		newLoc := e.moveArchetype(loc, newMask, id)

		// 2. Write new componts to new archetype/index location
		e.writeIndex(newLoc, id, comp...)

		return newLoc
	}
}

// Moves an entity from one archetype to another, copying all of the data from the old archetype to the new one
func (e *archEngine) moveArchetype(oldLoc entLoc, newMask archetypeMask, id Id) entLoc {
	newArchId := e.dcr.getArchetypeId(e, newMask)
	newIndex := e.allocate(newArchId, id)
	newLoc := entLoc{newArchId, uint32(newIndex)}

	oldLookup := e.lookup[oldLoc.archId]

	for _, compId := range oldLookup.components {
		store := e.compStorage[compId]
		store.moveArchetype(oldLoc, newLoc)
	}

	e.TagForDeletion(oldLoc, id)

	return entLoc{newArchId, uint32(newIndex)}
}

// Moves an entity from one archetype to another, copying all of the data required by the new archetype
func (e *archEngine) moveArchetypeDown(oldLoc entLoc, newMask archetypeMask, id Id) entLoc {
	newArchId := e.dcr.getArchetypeId(e, newMask)
	newIndex := e.allocate(newArchId, id)

	newLoc := entLoc{newArchId, uint32(newIndex)}
	newLookup := e.lookup[newArchId]
	for _, compId := range newLookup.components {
		store := e.compStorage[compId]
		store.moveArchetype(oldLoc, newLoc) //oldArchId, oldIndex, newArchId, newIndex)
	}

	e.TagForDeletion(oldLoc, id)

	return newLoc
}

func (e *archEngine) ReadEntity(loc entLoc, id Id) *Entity {
	lookup := e.lookup[loc.archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}
	index := int(loc.index)

	ent := NewEntity()
	for n := range e.compStorage {
		if e.compStorage[n] != nil {
			e.compStorage[n].ReadToEntity(ent, loc.archId, index)
		}
	}
	return ent
}

func (e *archEngine) ReadRawEntity(loc entLoc, id Id) *RawEntity {
	lookup := e.lookup[loc.archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	ent := NewRawEntity()
	for n := range e.compStorage {
		if e.compStorage[n] != nil {
			e.compStorage[n].ReadToRawEntity(ent, loc.archId, int(loc.index))
		}
	}
	return ent
}

// This creates a "hole" in the archetype at the specified Id
// Once we get enough holes, we can re-pack the entire slice
// TODO - How many holes before we repack? How many holes to pack at a time?
func (e *archEngine) TagForDeletion(loc entLoc, id Id) {
	lookup := e.lookup[loc.archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	// This indicates that the index needs to be cleaned up and should be skipped in any list processing
	lookup.id[loc.index] = InvalidEntity

	// This is used to track the current list of indices that need to be cleaned
	lookup.holes = append(lookup.holes, int(loc.index))
}

// TODO: I removed this. there's one segement that needs to be fixed. Ideally it'd be nice to have a repack operation. I kinda feel like it isn't useful until after you do a bunch of writes though? Maybe execute repacks every so often after a command buffer execute
// func (e *archEngine) CleanupHoles(archId archetypeId) {
// 	lookup := e.lookup[archId]
// 	if lookup == nil {
// 		panic("Archetype doesn't have lookup list")
// 	}

// 	for _, index := range lookup.holes {
// 		// Pop all holes off the end of the archetype
// 		for {
// 			lastIndex := len(lookup.id) - 1
// 			if lastIndex < 0 {
// 				break
// 			} // Break if the index we are trying to pop off is -1
// 			lastId := lookup.id[lastIndex]
// 			if lastId == InvalidEntity {
// 				// If the last id is a hole, then slice it off
// 				lookup.id = lookup.id[:lastIndex]
// 				// for n := range e.compStorage {
// 				// 	if e.compStorage[n] != nil {
// 				// 		e.compStorage[n].Delete(archId, lastIndex)
// 				// 	}
// 				// }
// 				for n := range e.compStorage {
// 					if e.compStorage[n] != nil {
// 						e.compStorage[n].Delete(archId, lastIndex)
// 					}
// 				}

// 				continue // Try again
// 			}

// 			break
// 		}

// 		// Check bounds because we may have popped past our original index
// 		if index >= len(lookup.id) {
// 			continue
// 		}

// 		// Swap lastIndex (which is not a hole) with index (which is a hole)
// 		lastIndex := len(lookup.id) - 1
// 		lastId := lookup.id[lastIndex]
// 		if lastId == InvalidEntity {
// 			panic("Bug: This shouldn't happen")
// 		}

// 		// TODO: To fix this, you need to bubble the index swap up to the entLoc map. You probably want to relocate how the "CleanupHoles" gets called. I kinda feel like it shouldn't get executed on write?

// 		lookup.id[index] = lastId
// 		lookup.id = lookup.id[:lastIndex]
// 		lookup.index.Put(lastId, index)
// 		for n := range e.compStorage {
// 			if e.compStorage[n] != nil {
// 				e.compStorage[n].Delete(archId, index)
// 			}
// 		}
// 	}

// 	// Clear holes slice
// 	lookup.holes = lookup.holes[:0]
// }
