package ecs

import (
	"fmt"
	"reflect"
)

type Id uint32
type ArchId uint32

func name(t any) string {
	// n := reflect.TypeOf(t).String()
	// if n[0] == '*' {
	// 	return n[1:]
	// }

	n := reflect.TypeOf(t).Name()
	// fmt.Printf("\n%s", n)

	return n
}

type ComponentSlice[T any] struct {
	comp []T
}

// Note: This will panic if you write past the buffer by more than 1
func (s *ComponentSlice[T]) Write(index int, val T) {
	if index == len(s.comp) {
		// Case: index causes a single append (new element added)
		s.comp = append(s.comp, val)
	} else {
		// Case: index is inside the length
		// Edge: (Causes Panic): Index is greater than 1 plus length
		s.comp[index] = val
	}
}

type Lookup struct {
	index map[Id]int // A mapping from entity ids to array indices
	id []Id // An array of every id in the arch list (essentially a reverse mapping from index to Id)
	holes []int // List of indexes that have ben deleted
}

type Storage interface {
	ReadToEntity(*Entity, ArchId, int) bool
	Delete(ArchId, int)
	print(int)
	// GetComponentSlice(archId ArchId) SliceReader
}

type ComponentSliceStorage[T any] struct {
	slice map[ArchId]*ComponentSlice[T]
}

func (ss ComponentSliceStorage[T]) ReadToEntity(entity *Entity, archId ArchId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok { return false }
	entity.Add(C(cSlice.comp[index]))
	return true
}

// Delete is somewhat special because it deletes the index of the archId for the componentSlice
// but then plugs the hole by pushing the last element of the componentSlice into index
func (ss ComponentSliceStorage[T]) Delete(archId ArchId, index int) {
	cSlice, ok := ss.slice[archId]
	if !ok { return }

	lastVal := cSlice.comp[len(cSlice.comp)-1]
	cSlice.comp[index] = lastVal
	cSlice.comp = cSlice.comp[:len(cSlice.comp)-1]
}

func (s ComponentSliceStorage[T]) print(amount int) {
	for archId, compSlice := range s.slice {
		fmt.Printf("archId(%d) - %v\n", archId, *compSlice)
	}
}

// func (ss ComponentStorageSlice[T]) GetComponentSlice(archId ArchId) (SliceReader, bool) {
// 	return ss.slice[archId]
// }

// type SliceReader interface {
// 	Get(index int)
// }

// Provides generic storage for all archetypes
type ArchEngine struct {
	lookup map[ArchId]*Lookup

	// positions map[ArchId]ComponentSliceStorage[Position]
	// velocities map[ArchId]ComponentSliceStorage[Velocity]
	compSliceStorage map[string]Storage

	dcr *DCR
}

func NewArchEngine() *ArchEngine {
	return &ArchEngine{
		lookup: make(map[ArchId]*Lookup),
		compSliceStorage: make(map[string]Storage),
		dcr: NewDCR(),
	}
}

func (e *ArchEngine) Print(amount int) {
	fmt.Println("--- ArchEngine ---")
	max := amount
	for archId, lookup := range e.lookup {
		fmt.Printf("archId(%d) - lookup(%v)\n", archId, lookup)
		max--; if max <= 0 { break }
	}
	for name, storage := range e.compSliceStorage {
		fmt.Printf("name(%s) -\n", name)
		storage.print(amount)
		max--; if max <= 0 { break }
	}
	e.dcr.print()
}

func (e *ArchEngine) Count(anything ...any) int {
	archIds := e.Filter(anything...)

	total := 0
	for _, archId := range archIds {
		lookup, ok := e.lookup[archId]
		if !ok { panic(fmt.Sprintf("Couldnt find archId in ArchEngine lookup table: %d", archId)) }

		// Each id represents an entity that holds the requested component(s)
		// Each hole represents a deleted entity that used to hold the requested component(s)
		total = total + len(lookup.id) - len(lookup.holes)
	}
	return total
}

func (e *ArchEngine) GetArchId(comp ...Component) ArchId {
	return e.dcr.GetArchId(comp...)
}

// Returns the list of ArchIds that contain all components
// TODO - this can be optimized
func (e *ArchEngine) Filter(comp ...any) []ArchId {
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

func GetStorage[T any](e *ArchEngine) ComponentSliceStorage[T] {
	var val T
	n := name(val)
	ss, ok := e.compSliceStorage[n]
	if !ok {
		// TODO - have write call this spot
		ss = ComponentSliceStorage[T]{
			slice: make(map[ArchId]*ComponentSlice[T]),
		}
		e.compSliceStorage[n] = ss
	}
	storage := ss.(ComponentSliceStorage[T])

	return storage
}

func WriteArch[T any](e *ArchEngine, archId ArchId, id Id, val T) {
	lookup, ok := e.lookup[archId]
	if !ok {
		lookup = &Lookup{
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
	storage := GetStorage[T](e)

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		cSlice = &ComponentSlice[T]{
			comp: make([]T, 0),
		}
		storage.slice[archId] = cSlice
	}

	cSlice.Write(index, val)
}

func ReadArch[T any](e *ArchEngine, archId ArchId, id Id) (T, bool) {
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

	// fmt.Printf("ComponentSliceStorage[T] type: %s != %s", name(ss), name(ret))
	storage, ok := ss.(ComponentSliceStorage[T])
	if !ok { panic(fmt.Sprintf("Wrong ComponentSliceStorage[T] type: %s != %s", name(ss), name(ret))) }

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		return ret, false
	}

	return cSlice.comp[index], true
}

// TODO - Think: Is it better to read everything then push it into the new ArchId? Or better to migrate everything in place?
// Returns the ArchId of where the entity ends up
func (e *ArchEngine) RewriteArch(archId ArchId, id Id, comp ...Component) ArchId {
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

func (e *ArchEngine) ReadEntity(archId ArchId, id Id) *Entity {
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

// func (e *ArchEngine) DeleteAll(archId ArchId, id Id) {
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

// func (e *ArchEngine) trimHoles(archId ArchId) {
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
func (e *ArchEngine) TagForDeletion(archId ArchId, id Id) {
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

func (e *ArchEngine) CleanupHoles(archId ArchId) {
	lookup, ok := e.lookup[archId]
	if !ok { panic("Archetype doesn't have lookup list") }
	// fmt.Println("Cleaning Holes: ", len(lookup.holes))
	for _, index := range lookup.holes {
		// e.DeleteAll(archId, id)

		// Pop all holes off the end of the archetype
		for {
			lastIndex := len(lookup.id) - 1
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
