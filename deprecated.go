package ecs

import "fmt"

// Reads a specific component of the entity specified at id.
// Returns true if the entity was found and had that component, else returns false.
// Deprecated: This API is tentative, I'm trying to improve the QueryN construct so that it can capture this usecase.
func Read[T any](world *World, id Id) (T, bool) {
	var ret T
	loc, ok := world.arch.Get(id)
	if !ok {
		return ret, false
	}

	return readArch[T](world.engine, loc, id)
}

// Reads a pointer to the component of the entity at the specified id.
// Returns true if the entity was found and had that component, else returns false.
// This pointer is short lived and can become invalid if any other entity changes in the world
// Deprecated: This API is tentative, I'm trying to improve the QueryN construct so that it can capture this usecase.
func ReadPtr[T any](world *World, id Id) *T {
	loc, ok := world.arch.Get(id)
	if !ok {
		return nil
	}

	return readPtrArch[T](world.engine, loc, id)
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
