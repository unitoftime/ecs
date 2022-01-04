package ecs

import (
	"fmt"
	"reflect"
)

type Id uint32
type ArchId uint32

func name(t any) string {
	n := reflect.TypeOf(t).String()
	// if n[0] == '*' {
	// 	return n[1:]
	// }

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
	index map[Id]int
	id []Id
}

type Storage interface {
	ReadToEntity(*Entity, ArchId, int) bool
	Delete(ArchId, int)
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

func (e *ArchEngine) Print() {
	for k, v := range e.lookup {
		fmt.Println(k, "-", v)
	}
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
		panic("Arch engine doesn't have this storage (I should probably just instantiate it and replace this code with write")
	}
	storage, ok := ss.(ComponentSliceStorage[T])
	if !ok { panic("Wrong ComponentSliceStorage[T] type!") }
	return storage
}

func WriteArch[T any](e *ArchEngine, archId ArchId, id Id, val T) {
	lookup, ok := e.lookup[archId]
	if !ok {
		lookup = &Lookup{
			index: make(map[Id]int),
			id: make([]Id, 0),
		}
		e.lookup[archId] = lookup
	}

	index, ok := lookup.index[id]
	if !ok {
		// Because the Id hasn't been added to this arch, we need to append it to the end
		lookup.id = append(lookup.id, id)
		index = len(lookup.id) - 1
		lookup.index[id] = index
	}

	// Get the dynamic componentSliceStorage
	n := name(val)
	ss, ok := e.compSliceStorage[n]
	if !ok {
		ss = ComponentSliceStorage[T]{
			slice: make(map[ArchId]*ComponentSlice[T]),
		}
		e.compSliceStorage[n] = ss
	}
	storage, ok := ss.(ComponentSliceStorage[T])
	if !ok { panic("Wrong ComponentSliceStorage[T] type!") }

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
	storage, ok := ss.(ComponentSliceStorage[T])
	if !ok { panic("Wrong ComponentSliceStorage[T] type!") }

	// Get the underlying Archetype's componentSlice
	cSlice, ok := storage.slice[archId]
	if !ok {
		return ret, false
	}

	return cSlice.comp[index], true
}

// TODO - Think: Is it better to read everything then push it into the new ArchId? Or better to migrate everything in place?
func (e *ArchEngine) RewriteArch(archId ArchId, id Id, comp ...Component) {
	ent := e.ReadEntity(archId, id)

	currentComps := ent.Comps()
	newArchId := e.GetArchId(currentComps...)

	if archId == newArchId {
		// Case 1: Archetype stays the same
		for i := range comp {
			comp[i].Write(e, archId, id)
		}
	} else {
		// Case 2: Archetype changes
		// 1: Delete all components in old archetype
		e.DeleteAll(archId, id)

		// 2: Write current entity to world
		for _, c := range ent.comp {
			c.Write(e, newArchId, id)
		}
		// 3: Write new components to world
		for _, c := range comp {
			c.Write(e, newArchId, id)
		}

		// 4: Write the new lookupList???
	}
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

func (e *ArchEngine) DeleteAll(archId ArchId, id Id) {
	lookup, ok := e.lookup[archId]
	if !ok { panic("Archetype doesn't have lookup list") }

	index, ok := lookup.index[id]
	if !ok { panic("Archetype doesn't contain ID") }

	lastVal := lookup.id[len(lookup.id)-1]
	lookup.id[index] = lastVal
	lookup.id = lookup.id[:len(lookup.id)-1]

	for n := range e.compSliceStorage {
		e.compSliceStorage[n].Delete(archId, index)
	}
}

type Entity struct {
	comp map[string]Component
}

func NewEntity() *Entity {
	return &Entity{
		comp: make(map[string]Component),
	}
}

func (e *Entity) Add(comp Component) {
	// n := name(comp) // TODO - name is wrong here because we pass in a boxed component
	n := comp.Name()
	e.comp[n] = comp
}

// TODO - Hacky and probs slow
func (e *Entity) Comps() []Component {
	ret := make([]Component, 0, len(e.comp))
	for _, v := range e.comp {
		ret = append(ret, v)
	}
	return ret
}
