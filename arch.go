package ecs

import (
	"fmt"
	"reflect"
)

type Id uint32
type ArchId uint32

var registeredComponents = make(map[reflect.Type]CompId)
var invalidComponentId CompId = 0
var componentRegistryCounter CompId = 1

func name(t any) CompId {
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

type lookupList struct {
	index map[Id]int // A mapping from entity ids to array indices
	id []Id // An array of every id in the arch list (essentially a reverse mapping from index to Id)
	holes []int // List of indexes that have ben deleted
}

type storage interface {
	ReadToEntity(*Entity, ArchId, int) bool
	Delete(ArchId, int)
	print(int)
}

type componentSliceStorage[T any] struct {
	slice map[ArchId]*componentSlice[T]
}

func (ss componentSliceStorage[T]) ReadToEntity(entity *Entity, archId ArchId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok { return false }
	entity.Add(C(cSlice.comp[index]))
	return true
}

// Delete is somewhat special because it deletes the index of the archId for the componentSlice
// but then plugs the hole by pushing the last element of the componentSlice into index
func (ss componentSliceStorage[T]) Delete(archId ArchId, index int) {
	cSlice, ok := ss.slice[archId]
	if !ok { return }

	lastVal := cSlice.comp[len(cSlice.comp)-1]
	cSlice.comp[index] = lastVal
	cSlice.comp = cSlice.comp[:len(cSlice.comp)-1]
}

func (s componentSliceStorage[T]) print(amount int) {
	for archId, compSlice := range s.slice {
		fmt.Printf("archId(%d) - %v\n", archId, *compSlice)
	}
}

// Provides generic storage for all archetypes
type archEngine struct {
	lookup map[ArchId]*lookupList

	compSliceStorage map[CompId]storage

	dcr *componentRegistry
}

func newArchEngine() *archEngine {
	return &archEngine{
		lookup: make(map[ArchId]*lookupList),
		compSliceStorage: make(map[CompId]storage),
		dcr: newComponentRegistry(),
	}
}

func (e *archEngine) generation() int {
	return e.dcr.generation
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
	archIds := e.Filter(anything...)

	total := 0
	for _, archId := range archIds {
		lookup, ok := e.lookup[archId]
		if !ok { panic(fmt.Sprintf("Couldnt find archId in archEngine lookup table: %d", archId)) }

		// Each id represents an entity that holds the requested component(s)
		// Each hole represents a deleted entity that used to hold the requested component(s)
		total = total + len(lookup.id) - len(lookup.holes)
	}
	return total
}

func (e *archEngine) GetArchId(comp ...Component) ArchId {
	return e.dcr.GetArchId(comp...)
}

// TODO - using this makes things not thread safe
// TODO - map might be slower than just having an array. I could probably do a big bitmask and then just do a logical OR
var filterLists = make([]map[ArchId]bool, 0)
func (e *archEngine) FilterList(archIds []ArchId, comp []any) []ArchId {
	filterLists = filterLists[:0]

	for i := range comp {
		n := name(comp[i])
		filterLists = append(filterLists, e.dcr.archSet[n])
	}

	archIds = archIds[:0]
	for archId := range filterLists[0] {
		missing := false
		for i := range filterLists {
			_, exists := filterLists[i][archId]
			if !exists {
				missing = true
				break // at least one set was missing
			}
		}
		if !missing {
			archIds = append(archIds, archId)
		}
	}

	return archIds
}

// TODO!!! - dump this for FilterList
// Returns the list of ArchIds that contain all components
// TODO - this can be optimized
// var filterLists = make([]map[ArchId]bool, 0)
// // var returnedArchIds = make([][]ArchId, 1024) // TODO!!!! - this means that at max you can nest 1024 map functions
// // var currentIndexForReturnedArchIds = 0
// var returnedArchIds = make([]ArchId, 1024) // TODO!!! - this means you cant nest map functions
func (e *archEngine) Filter(comp ...any) []ArchId {
	// filterLists = filterLists[:0]

	// for i := range comp {
	// 	n := name(comp[i])
	// 	filterLists = append(filterLists, e.dcr.archSet[n])
	// }

	// // archIds := make([]ArchId, 0)
	// archIds := returnedArchIds[:0]
	// for archId := range filterLists[0] {
	// 	missing := false
	// 	for i := range filterLists {
	// 		_, exists := filterLists[i][archId]
	// 		if !exists {
	// 			missing = true
	// 			break // at least one set was missing
	// 		}
	// 	}
	// 	if !missing {
	// 		archIds = append(archIds, archId)
	// 	}
	// }

	// return archIds

	lists := make([]map[ArchId]bool, 0)
	for i := range comp {
		n := name(comp[i])
		lists = append(lists, e.dcr.archSet[n])
	}

	archIds := make([]ArchId, 0)
	for archId := range lists[0] {
		missing := false
		for i := range lists {
			_, exists := lists[i][archId]
			if !exists {
				missing = true
				break // at least one set was missing
			}
		}
		if !missing {
			archIds = append(archIds, archId)
		}
	}

	return archIds
}

func getStorage[T any](e *archEngine) componentSliceStorage[T] {
	var val T
	n := name(val)
	// n := nameGen[T]()
	ss, ok := e.compSliceStorage[n]
	if !ok {
		// TODO - have write call this spot
		ss = componentSliceStorage[T]{
			slice: make(map[ArchId]*componentSlice[T]),
		}
		e.compSliceStorage[n] = ss
	}
	storage := ss.(componentSliceStorage[T])

	return storage
}

func writeArch[T any](e *archEngine, archId ArchId, id Id, val T) {
	lookup, ok := e.lookup[archId]
	if !ok {
		lookup = &lookupList{
			index: make(map[Id]int),
			id: make([]Id, 0),
			holes: make([]int, 0),
		}
		e.lookup[archId] = lookup
	}

	// Check if we want to cleanup holes
	if len(lookup.holes) >= 1024 { // TODO - Hardcoded number, maybe make it percentage based on holes per total entities
		e.CleanupHoles(archId)
	}

	index, ok := lookup.index[id]
	if !ok {
		// Because the Id hasn't been added to this arch, we need to append it to the end
		lookup.id = append(lookup.id, id)
		index = len(lookup.id) - 1
		lookup.index[id] = index
	}

	// Get the componentSliceStorage
	storage := getStorage[T](e)

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		cSlice = &componentSlice[T]{
			comp: make([]T, 0),
		}
		storage.slice[archId] = cSlice
	}

	cSlice.Write(index, val)
}

func readArch[T any](e *archEngine, archId ArchId, id Id) (T, bool) {
	var ret T
	lookup, ok := e.lookup[archId]
	if !ok {
		return ret, false
	}

	index, ok := lookup.index[id]
	if !ok {
		return ret, false
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss, ok := e.compSliceStorage[n]
	if !ok {
		return ret, false
	}

	// fmt.Printf("componentSliceStorage[T] type: %s != %s", name(ss), name(ret))
	storage, ok := ss.(componentSliceStorage[T])
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

func readPtrArch[T any](e *archEngine, archId ArchId, id Id) *T {
	var ret T
	lookup, ok := e.lookup[archId]
	if !ok {
		return nil
	}

	index, ok := lookup.index[id]
	if !ok {
		return nil
	}

	// Get the dynamic componentSliceStorage
	n := name(ret)
	ss, ok := e.compSliceStorage[n]
	if !ok {
		return nil
	}

	// fmt.Printf("componentSliceStorage[T] type: %s != %s", name(ss), name(ret))
	storage, ok := ss.(componentSliceStorage[T])
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

// TODO - Think: Is it better to read everything then push it into the new ArchId? Or better to migrate everything in place?
// Returns the ArchId of where the entity ends up
func (e *archEngine) rewriteArch(archId ArchId, id Id, comp ...Component) ArchId {
	// fmt.Println("RewriteArch")
	ent := e.ReadEntity(archId, id)

	// currentComps := ent.Comps()
	// fmt.Println("Current", currentComps)

	ent.Add(comp...)
	combinedComps := ent.Comps()
	newArchId := e.GetArchId(combinedComps...)

	// fmt.Println("archId == newArchId", archId, newArchId)
	if archId == newArchId {
		// Case 1: Archetype stays the same
		for i := range comp {
			comp[i].Write(e, archId, id)
		}
	} else {
		// Case 2: Archetype changes
		// 1: Delete all components in old archetype
		// e.DeleteAll(archId, id)
		e.TagForDeletion(archId, id)

		// 2: Write current entity to world
		for _, c := range ent.comp {
			c.Write(e, newArchId, id)
		}
		// 3: Write new components to world
		for _, c := range comp {
			c.Write(e, newArchId, id)
		}

		// 4: TODO - Write the new lookupList???
	}
	return newArchId
}

func (e *archEngine) ReadEntity(archId ArchId, id Id) *Entity {
	lookup, ok := e.lookup[archId]
	if !ok { panic("Archetype doesn't have lookup list") }

	index, ok := lookup.index[id]
	if !ok { panic("Archetype doesn't contain ID") }

	ent := NewEntity()
	for n := range e.compSliceStorage {
		e.compSliceStorage[n].ReadToEntity(ent, archId, index)
	}
	return ent
}

// func (e *archEngine) DeleteAll(archId ArchId, id Id) {
// 	// Trim all holes off the end of the lookup list
// 	e.trimHoles(archId)

// 	lookup, ok := e.lookup[archId]
// 	if !ok { panic("Archetype doesn't have lookup list") }

// 	index, ok := lookup.index[id]
// 	if !ok { panic("Archetype doesn't contain ID") }

// 	if index == (len(lookup.id) - 1) {
// 		// Edge Case: If index is already the last element, just slice the end
// 		lookup.id = lookup.id[:len(lookup.id)-1]
// 		// delete(lookup.index, id)
// 		for n := range e.compSliceStorage {
// 			e.compSliceStorage[n].Delete(archId, index)
// 		}

// 		return
// 	}

// 	// Swap last element with hole
// 	lastId := lookup.id[len(lookup.id)-1]
// 	fmt.Println("DeleteAll:", archId, id, index, lastId)
// 	lookup.id[index] = lastId
// 	lookup.id = lookup.id[:len(lookup.id)-1]

// 	lookup.index[lastId] = index
// 	// delete(lookup.index, id)

// 	for n := range e.compSliceStorage {
// 		e.compSliceStorage[n].Delete(archId, index)
// 	}
// }

// func (e *archEngine) trimHoles(archId ArchId) {
// 	lookup, ok := e.lookup[archId]
// 	if !ok { panic("Archetype doesn't have lookup list") }

// 	// Trim the end until there are no holes there
// 	for {
// 		lastId := lookup.id[len(lookup.id)-1]
// 		if lastId == InvalidEntity {
// 			// If it's a hole, then slice it off and try again
// 			lookup.id = lookup.id[:len(lookup.id)-1]
// 			// delete(lookup.index, lastId) // No need to do this because lastId has already been deleted
// 			for n := range e.compSliceStorage {
// 				e.compSliceStorage[n].Delete(archId, len(lookup.id)-1)
// 			}
// 			continue
// 		}

// 		// If it wasn't a hole then proceed
// 		break
// 	}
// }

// This creates a "hole" in the archetype at the specified Id
// Once we get enough holes, we can re-pack the entire slice
// TODO - How many holes before we repack? How many holes to pack at a time?
func (e *archEngine) TagForDeletion(archId ArchId, id Id) {
	lookup, ok := e.lookup[archId]
	if !ok { panic("Archetype doesn't have lookup list") }

	index, ok := lookup.index[id]
	if !ok { panic("Archetype doesn't contain ID") }

	// This indicates that the index needs to be cleaned up and should be skipped in any list processing
	lookup.id[index] = InvalidEntity
	delete(lookup.index, id)

	// This is used to track the current list of indices that need to be cleaned
	lookup.holes = append(lookup.holes, index)
}

func (e *archEngine) CleanupHoles(archId ArchId) {
	lookup, ok := e.lookup[archId]
	if !ok { panic("Archetype doesn't have lookup list") }
	// fmt.Println("Cleaning Holes: ", len(lookup.holes))
	for _, index := range lookup.holes {
		// e.DeleteAll(archId, id)

		// Pop all holes off the end of the archetype
		for {
			lastIndex := len(lookup.id) - 1
			if lastIndex < 0 { break } // Break if the index we are trying to pop off is -1
			lastId := lookup.id[lastIndex]
			if lastId == InvalidEntity {
				// If the last id is a hole, then slice it off
				lookup.id = lookup.id[:lastIndex]
				for n := range e.compSliceStorage {
					e.compSliceStorage[n].Delete(archId, lastIndex)
				}

				continue // Try again
			}

			break
		}

		// Check bounds because we may have popped past our original index
		if index >= len(lookup.id) { continue }

		// Swap lastIndex (which is not a hole) with index (which is a hole)
		lastIndex := len(lookup.id) - 1
		lastId := lookup.id[lastIndex]
		if lastId == InvalidEntity { panic("Bug: This shouldn't happen")}

		lookup.id[index] = lastId
		lookup.id = lookup.id[:lastIndex]
		lookup.index[lastId] = index
		for n := range e.compSliceStorage {
			e.compSliceStorage[n].Delete(archId, index)
		}
	}

	// Clear holes slice
	lookup.holes = lookup.holes[:0]
}
