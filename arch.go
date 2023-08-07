package ecs

import (
	"fmt"
	"sync"
	"reflect"
)

// This is the identifier for entities in the world
//cod:struct
type Id uint32

type archetypeId uint32

var componentIdMutex sync.Mutex
var registeredComponents = make(map[reflect.Type]componentId, maxComponentId)
var invalidComponentId componentId = 0
var componentRegistryCounter componentId = 1

func name(t any) componentId {
	// Note: We have to lock here in case there are multiple worlds
	// TODO!! - This probably causes some performance penalty
	componentIdMutex.Lock()
	defer componentIdMutex.Unlock()

	typeof := reflect.TypeOf(t)
	compId, ok := registeredComponents[typeof]
	if !ok {
		compId = componentRegistryCounter
		registeredComponents[typeof] = compId
		componentRegistryCounter++
	}
	return compId
}

type componentSlice[T any] struct {
	comp []T
}

// Note: This will panic if you write past the buffer by more than 1
func (s *componentSlice[T]) Write(index int, val T) {
	if index == len(s.comp) {
		// Case: index causes a single append (new element added)
		s.comp = append(s.comp, val)
	} else {
		// Case: index is inside the length
		// Edge: (Causes Panic): Index is greater than 1 plus length
		s.comp[index] = val
	}
}

// TODO: Rename, this is kind of like an archetype header
type lookupList struct {
	index map[Id]int // A mapping from entity ids to array indices
	id    []Id       // An array of every id in the arch list (essentially a reverse mapping from index to Id)
	holes []int      // List of indexes that have ben deleted
	mask archetypeMask
}

// Adds ourselves to the last available hole, else appends
// Returns the index
func (l *lookupList) addToEasiestHole(id Id) int {
	if len(l.holes) > 0 {
		lastHoleIndex := len(l.holes)-1
		index := l.holes[lastHoleIndex]
		l.id[index] = id
		l.index[id] = index

		l.holes = l.holes[:lastHoleIndex]
		return index
	} else {
		// Because the Id hasn't been added to this arch, we need to append it to the end
		l.id = append(l.id, id)
		index := len(l.id) - 1
		l.index[id] = index
		return index
	}
}


type storage interface {
	ReadToEntity(*Entity, archetypeId, int) bool
	ReadToRawEntity(*RawEntity, archetypeId, int) bool
	Delete(archetypeId, int)
	print(int)
}

type componentSliceStorage[T any] struct {
	slice map[archetypeId]*componentSlice[T]
}

func (ss *componentSliceStorage[T]) ReadToEntity(entity *Entity, archId archetypeId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return false
	}
	entity.Add(C(cSlice.comp[index]))
	return true
}

func (ss *componentSliceStorage[T]) ReadToRawEntity(entity *RawEntity, archId archetypeId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return false
	}
	entity.Add(&cSlice.comp[index])
	return true
}

// Delete is somewhat special because it deletes the index of the archId for the componentSlice
// but then plugs the hole by pushing the last element of the componentSlice into index
func (ss *componentSliceStorage[T]) Delete(archId archetypeId, index int) {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return
	}

	lastVal := cSlice.comp[len(cSlice.comp)-1]
	cSlice.comp[index] = lastVal
	cSlice.comp = cSlice.comp[:len(cSlice.comp)-1]
}

func (s *componentSliceStorage[T]) print(amount int) {
	for archId, compSlice := range s.slice {
		fmt.Printf("archId(%d) - %v\n", archId, *compSlice)
	}
}

// Provides generic storage for all archetypes
type archEngine struct {
	generation  int
	// archCounter archetypeId

	lookup []*lookupList // Indexed by archetypeId
	compSliceStorage []storage // Indexed by componentId
	dcr *componentRegistry

	// TODO - using this makes things not thread safe inside the engine
	archCount map[archetypeId]int
}

func newArchEngine() *archEngine {
	return &archEngine{
		generation:  1, // Start at 1 so that anyone with the default int value will always realize they are in the wrong generation

		lookup:           make([]*lookupList, 0, DefaultAllocation),
		compSliceStorage: make([]storage, maxComponentId + 1),
		dcr:              newComponentRegistry(),
		archCount:        make(map[archetypeId]int),
	}
}

func (e *archEngine) newArchetypeId(archMask archetypeMask) archetypeId {
	e.generation++ // Increment the generation

	archId := archetypeId(len(e.lookup))
	e.lookup = append(e.lookup,
		&lookupList{
			index: make(map[Id]int),
			id:    make([]Id, 0, DefaultAllocation),
			holes: make([]int, 0, DefaultAllocation),
			mask: archMask,
		},
	)

	return archId
}

func (e *archEngine) getGeneration() int {
	return e.generation
}

// func (e *archEngine) Print(amount int) {
// 	fmt.Println("--- archEngine ---")
// 	max := amount
// 	for archId, lookup := range e.lookup {
// 		fmt.Printf("archId(%d) - lookup(%v)\n", archId, lookup)
// 		max--; if max <= 0 { break }
// 	}
// 	for name, storage := range e.compSliceStorage {
// 		fmt.Printf("name(%s) -\n", name)
// 		storage.print(amount)
// 		max--; if max <= 0 { break }
// 	}
// 	e.dcr.print()
// }

func (e *archEngine) count(anything ...any) int {
	comps := make([]componentId, len(anything))
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

func (e *archEngine) getArchetypeId(comp ...Component) archetypeId {
	return e.dcr.getArchetypeId(e, comp...)
}

// TODO - map might be slower than just having an array. I could probably do a big bitmask and then just do a logical OR
func (e *archEngine) FilterList(archIds []archetypeId, comp []componentId) []archetypeId {
	// TODO: could I maybe do something more optimal with archetypeMask?
	// New way: With archSets that are just slices
	// Logic: Go thorugh and keep track of how many times we see each archetype. Then only keep the archetypes that we've seen an amount of times equal to the number of components. If we have 5 components and see 5 for a specific archId, it means that each component has that archId
	// TODO: this may be more efficient to use a slice?

	// Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	for k := range e.archCount {
		delete(e.archCount, k)
	}

	for _, compId := range comp {
		for _, archId := range e.dcr.archSet[compId] {
			e.archCount[archId] = e.archCount[archId] + 1
		}
	}

	numComponents := len(comp)

	archIds = archIds[:0]
	for archId, count := range e.archCount {
		if count >= numComponents {
			archIds = append(archIds, archId)

			// // TODO: How tight do I want my tolerances?
			// if count > numComponents {
			// 	panic("AAAA")
			// }
		}
	}

	return archIds
}

func getStorage[T any](e *archEngine) *componentSliceStorage[T] {
	var val T
	n := name(val)
	return getStorageByCompId[T](e, n)
}

// Note: This will panic if the wrong compId doesn't match the generic type
func getStorageByCompId[T any](e *archEngine, compId componentId) *componentSliceStorage[T] {
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

	index, ok := lookup.index[id]
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

	// Loop through all components and add them to individual component slices
	for i := range comp {
		comp[i].write(e, archId, index)
	}
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

	index, ok := lookup.index[id]
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

	index, ok := lookup.index[id]
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

// TODO - Think: Is it better to read everything then push it into the new archetypeId? Or better to migrate everything in place?
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
		ent := e.ReadEntity(archId, id)
		ent.Add(comp...)
		combinedComps := ent.Comps()
		// Note (maybe TODO): Right now we don't know if the new archtypeId exists for this, so we have to go through the normal code path to potentially create it (even though we currently have the new archtypeMask)
		newArchId := e.getArchetypeId(combinedComps...)

		// Case 2: Archetype changes
		// 1: Delete all components in old archetype
		e.TagForDeletion(archId, id)

		// 2: We need to write the entire list of combinedComps
		e.write(newArchId, id, combinedComps...)
		return newArchId
	}
}

func (e *archEngine) ReadEntity(archId archetypeId, id Id) *Entity {
	lookup := e.lookup[archId]
	if lookup == nil {
		panic("Archetype doesn't have lookup list")
	}

	index, ok := lookup.index[id]
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

	index, ok := lookup.index[id]
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

	index, ok := lookup.index[id]
	if !ok {
		panic("Archetype doesn't contain ID")
	}

	// This indicates that the index needs to be cleaned up and should be skipped in any list processing
	lookup.id[index] = InvalidEntity
	delete(lookup.index, id)

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
		lookup.index[lastId] = index
		for n := range e.compSliceStorage {
			if e.compSliceStorage[n] != nil {
				e.compSliceStorage[n].Delete(archId, index)
			}
		}
	}

	// Clear holes slice
	lookup.holes = lookup.holes[:0]
}
