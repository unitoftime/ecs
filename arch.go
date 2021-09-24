package ecs

import (
	"fmt"
	// "log"
)

type ArchMask uint64 // Restricts us to a max of 64 different components

type ArchComponent interface {
	ComponentSet(interface{})
	InternalRead(int, interface{})
	InternalWrite(int, interface{})
	InternalAppend(interface{})
	InternalPointer(int) interface{}
	InternalReadVal(int) interface{}
	Len() int
	Delete(int)
}

type ArchEngine struct {
	archCounter ArchId
	reg map[string]*ArchStorage
	archLookup map[ArchMask]ArchId
}

func NewArchEngine() *ArchEngine {
	return &ArchEngine{
		archCounter: 0,
		reg: make(map[string]*ArchStorage),
		archLookup: make(map[ArchMask]ArchId),
	}
}

func (e *ArchEngine) NewArchId() ArchId {
	archId := e.archCounter
	e.archCounter++
	return archId
}

func (e *ArchEngine) Print() {
	for k,v := range e.reg {
		fmt.Println(k, *v)
	}
}

func (e *ArchEngine) GetArchMask(comp ...interface{}) ArchMask {
	ret := ArchMask(0)
	for i := range comp {
		ret += componentRegistry.GetComponentMask(comp[i])
	}
	return ret
}

// Uses component mask to generate an archetype ID or creates that archetype
func (e *ArchEngine) GetArchId(comp ...interface{}) ArchId {
	mask := e.GetArchMask(comp...)
	archId, ok := e.archLookup[mask]
	if !ok {
		// Need to build this archetype because it doesn't exist
		archId = e.NewArchId()
		e.archLookup[mask] = archId

		// All archetypes get the LookupList component
		lookup := &LookupList{
			Lookup: make(map[Id]int),
			Ids: make([]Id, 0),
		}
		ArchWrite(e, archId, lookup)

		// Add all component Lists to this archetype
		for i := range comp {
			list := componentRegistry.GetArchStorageType(comp[i])
			ArchWrite(e, archId, list)
		}
	}

	return archId
}

func ArchGetStorage(e *ArchEngine, t interface{}) *ArchStorage {
	name := name(t)
	storage, ok := e.reg[name]
	if !ok {
		e.reg[name] = NewArchStorage()
		storage, _ = e.reg[name]
	}
	return storage
}

func ArchPrint(e *ArchEngine, id ArchId, val interface{}) {
	storage := ArchGetStorage(e, val)
	fmt.Println(storage)
}

func ArchRead(e *ArchEngine, id ArchId, val ArchComponent) bool {
	storage := ArchGetStorage(e, val)
	newVal, ok := storage.Read(id)
	if ok {
		val.ComponentSet(newVal)
	}
	return ok
}

func ArchWrite(e *ArchEngine, id ArchId, val interface{}) {
	storage := ArchGetStorage(e, val)
	storage.Write(id, val)
}

func ArchReadAll(e *ArchEngine, archId ArchId) []ArchComponent {
	ret := make([]ArchComponent, 0)
	for _, storage := range e.reg {
		comp, ok := storage.list[archId]
		if !ok { continue }

		ret = append(ret, comp.(ArchComponent))
	}
	return ret
}

func ArchEach(engine *ArchEngine, t interface{}, f func(id ArchId, a interface{})) {
	storage := ArchGetStorage(engine, t)
	for id, a := range storage.list {
		f(id, a)
	}
}

// func ArchFilterSingle(engine *ArchEngine, comp interface{}) []ArchId {
// 	archIds := make([]ArchId, 0)

// 	// Find all archetypes that have comp[0]
// 	switch comp.(type) {
// 	case d1:
// 		ArchEach(engine, d1List{}, func(id ArchId, a interface{}) {
// 			archIds = append(archIds, id)
// 		})
// 	case d2:
// 		ArchEach(engine, d2List{}, func(id ArchId, a interface{}) {
// 			archIds = append(archIds, id)
// 		})
// 	default:
// 		panic("Unknown component type!")
// 	}

// 	return archIds
// }

// Get the ArchIds with all of these components
func ArchFilter(engine *ArchEngine, comp ...interface{}) []ArchId {
	if len(comp) <= 0 { panic("Must have at least one component!") }

	archIds := make([]ArchId, 0)

	// Find all archetypes that have comp[0]
	{
		storageType := componentRegistry.GetArchStorageType(comp[0])
		storage := ArchGetStorage(engine, storageType)
		for id := range storage.list {
			archIds = append(archIds, id)
		}
	}

	// log.Println("Initial:", archIds)

	finalArchIds := make([]ArchId, 0)
	// Loop over archetypes and remove ones that don't have all components
	for _, archId := range archIds {
		hasAllComps := true
		for i := 1; i < len(comp); i++ {
			storageType := componentRegistry.GetArchStorageType(comp[i])
			hasAllComps = hasAllComps && ArchRead(engine, archId, storageType)

			if !hasAllComps { break } // We can exit early if we are missing just one comp
		}

		// Only add archetypes which have all comps
		if hasAllComps {
			finalArchIds = append(finalArchIds, archId)
		}
	}

	return finalArchIds
}

// Get the ArchId with all of these components and only these components
// TODO - use a bitmask to track these
// func ArchFilterOnly(engine *ArchEngine, comp ...interface{}) {
// 	if len(comp) <= 0 { panic("Must have at least one component!") }

// 	final := make(map[ArchId]int)

// 	// Loop over all of the current archetypes and ensure they have this new component too
// 	for i := 1; i < len(comp); i++ {
// 		ArchEach(engine, comp[0], func(id ArchId, a interface{}) {
// 			final[id] = final[id] + 1
// 		})
// 	}

// 	for k,v := range final {
// 		if v == len(comp) {
// 			return k
// 		}
// 	}
// 	panic("Failed to find archetype!")
// }

// For storing archetype component group interfaces
// TODO - use array backed map instead of map for arch storage? we never remove archIds
type ArchStorage struct {
	list map[ArchId]interface{}
}
func NewArchStorage() *ArchStorage {
	return &ArchStorage{
		list: make(map[ArchId]interface{}),
	}
}

func (s *ArchStorage) Read(id ArchId) (interface{}, bool) {
	val, ok := s.list[id]
	return val, ok
}
func (s *ArchStorage) Write(id ArchId, val interface{}) {
	s.list[id] = val
}
// func (s *ArchStorage) Delete(id ArchId) {
// 	delete(s.list, id)
// }
