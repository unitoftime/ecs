package ecs

import (
	"fmt"
	"reflect"
)

// TODO - Replace with constraints
type Slice[Elem any] interface { ~[]Elem }

type Id uint32
type ArchId uint32

// TODO - Replace?
func name[T any](t T) string {
	n := reflect.TypeOf(t).String()
	if n[0] == '*' {
		return n[1:]
	}

	return n
}

const (
	InvalidEntity Id = 0
	UniqueEntity Id = 1
)

type World struct {
	idCounter Id
	archLookup map[Id]ArchId
	archEngine *ArchEngine
	tags map[string]map[Id]bool
}

func NewWorld() *World {
	return &World{
		idCounter: UniqueEntity + 1,
		archLookup: make(map[Id]ArchId),
		archEngine: NewArchEngine(),
		tags: make(map[string]map[Id]bool),
	}
}

func (w *World) NewId() Id {
	if w.idCounter <= UniqueEntity {
		w.idCounter = UniqueEntity + 1
	}
	id := w.idCounter
	w.idCounter++
	return id
}

func (w *World) Print() {
	fmt.Printf("%v\n", w)
	w.archEngine.Print()
}

// https://cs.opensource.google/go/go/+/refs/tags/go1.17.5:src/encoding/gob/type.go;l=807
// TODO - RegisterName
func Register[T any](world *World) {
	var t T
	n := name(t)
	// world.archEngine.dcr.componentStorageType[n] = NewArchStorage[[]T, T]()
	world.archEngine.reg[n] = NewArchStorage[[]T, T]()
}

func Read[T any](world *World, id Id) (T, bool) {
	var ret T

	archId, ok := world.archLookup[id]
	if !ok {
		// Entity ID does not exist if it doesn't exist in the bookkeeping
		return ret, false
	}

	// lookup, ok := ArchRead[LookupList](world.archEngine, archId)
	lookup, ok := world.archEngine.lookup[archId]
	index, ok := lookup.Lookup[id]
	if !ok { panic("World bookkeeping said entity was here, but lookupList said it isn't") }

	// list, ok := ArchRead[T](world.archEngine, archId)
	// if !ok { return ret, false }

	// var list []T
	// ok = world.archEngine.Read(archId, ret, &list)
	// if !ok { return ret, false }

	// ret = list[index]
	// return ret, true

	storage, ok := world.archEngine.Read2(archId, ret)
	if !ok { return ret, false } // Arch doesn't have this component
	val, ok := storage.InternalGet(archId, index)
	if !ok { panic("Entity should have componenet because its in arch, but didn't") }
	return val.(T), true
}

func Write(world *World, id Id, comp ...any) {
	// var t T
	archId, ok := world.archLookup[id]

	if ok {
		// The Entity is already constructed, Update the correct archetype
		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		index, ok := lookup.Lookup[id]
		if !ok { panic("World bookkeeping said entity was here, but lookupList said it isn't") }

		// TODO - push this loop into another func? shared with other part of write
		for _, c := range comp {
			_, ok := world.archEngine.Read2(archId, c)
			if !ok {
				// If we go in here then we will need to move the entity to a new archId
				// Read entire entity
				ent := ReadEntity(world, id)
				// Delete the entity
				Delete(world, id)
				// Loop over all components and add them to the ent
				for _, c := range comp {
					ent.Add(c)
				}

				// Get a destination ArchId
				// TODO - make a version of this function that uses an ent rather than comp slice?
				archId = world.archEngine.GetArchId(ent.Comps()...)
				lookup, ok := world.archEngine.lookup[archId]
				if !ok { panic("LookupList is missing!") }
				lookup.Ids = append(lookup.Ids, id)
				index := len(lookup.Ids) - 1
				lookup.Lookup[id] = index

				world.archEngine.lookup[archId] = lookup

				// Update the world's archetype lookup
				world.archLookup[id] = archId

				// Write the entity back
				for _, c := range ent.comp {
					storage, ok := world.archEngine.Read2(archId, c)
					if !ok { panic("Unable to read storage for this component type") }
					storage.InternalAppend(archId, c)
				}

				// Ensure we exit early now that we've written
				return
			}
		}

		// If we are here then the archId will not be changing. Just overwrite every component
		for _, c := range comp {
			storage, ok := world.archEngine.Read2(archId, c)
			if !ok { panic("Unable to read storage for this component type") }
			storage.InternalSet(archId, index, c)
		}

		// lookup := LookupList{}
		// ArchRead(world.archEngine, archId, &lookup)
		// index, ok := lookup.Lookup[id]
		// if !ok { panic("World bookkeeping said entity was here, but lookupList said it isn't") }

		// list := cList[T]{}
		// ok = ArchRead(world.archEngine, archId, &list)
		// if !ok {
		// 	//Archetype didn't have this component, move the entity to a new archetype
		// 	moveAndAdd[T](world, id, comp)
		// 	return
		// } else {
		// 	list.InternalWrite(index, comp)
		// 	ArchWrite(world.archEngine, archId, list)
		// 	return
		// }
	} else {
		// The Entity isn't added yet. Construct it based on components
		archId = world.archEngine.GetArchId(comp...)

		// Update the archetype's lookup with the new entity
		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		lookup.Ids = append(lookup.Ids, id)
		index := len(lookup.Ids) - 1
		lookup.Lookup[id] = index

		world.archEngine.lookup[archId] = lookup

		// Update the world's archetype lookup
		world.archLookup[id] = archId

		// Read and append to the component list
		// var list []T
		// ok = world.archEngine.Read(archId, t, &list)
		// if !ok { panic("Archetype didn't have this component!") }
		for _, c := range comp {
			storage, ok := world.archEngine.Read2(archId, c)
			if !ok { panic("Unable to read storage for this component type") }
			storage.InternalAppend(archId, c)
		}


		// list = append(list, comp)
		// if len(list) != len(lookup.Lookup) {
		// 	panic("lookupList length doesn't match component list length!")
		// }

		// // Write back the component list
		// // ArchWrite(world.archEngine, archId, list)
		// world.archEngine.Write(archId, t, list)
	}
}

func Delete(world *World, id Id) {
	archId, ok := world.archLookup[id]
	if !ok { return } // Exit early if id doesn't exist

	// Delete index from all relevant component storages
	world.archEngine.DeleteAll(archId, id)

	delete(world.archLookup, id)
}

// Delete a component from a entity
// func DeleteComponents(world *World, id Id, comp ...interface{}) {
// 	archId, ok := world.archLookup[id]
// 	if !ok { return } // Return if id doesn't exist in the system

// 	lookup, ok := world.archEngine.lookup[archId]
// 	if !ok { panic("LookupList is missing!") }
// 	index, ok := lookup.Lookup[id]
// 	if !ok { panic("Entity ID doesn't exist in archetype ID!") }

// 	archComponents := world.archEngine.ReadAll(archId)
// 	for _, archComp := range archComponents {
// 		archComp.Delete(index)
// 	}
// }

// Represents a standalone entity with all of its components
type Entity struct {
	comp map[string]any
}

func NewEntity() Entity {
	return Entity{
		comp: make(map[string]any),
	}
}

func (e *Entity) Add(comp any) {
	n := name(comp)
	e.comp[n] = comp
}

func (e *Entity) Comps() []any {
	ret := make([]any, 0, len(e.comp))
	for _, v := range e.comp {
		ret = append(ret, v)
	}
	return ret
}

// TODO - Return a boolean as well?
func ReadEntity(world *World, id Id) Entity {
	archId, ok := world.archLookup[id]
	if !ok {
		return Entity{}
	}

	ent := world.archEngine.ReadAll(archId, id)

	return ent
}
