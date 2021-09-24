package aecs

import (
	"log"
	"fmt"
	"reflect"
)

type Id uint32
type ArchId uint32

func name(t interface{}) string {
	name := reflect.TypeOf(t).String()
	if name[0] == '*' {
		return name[1:]
	}

	return name
}

const (
	InvalidEntity Id = 0
	UniqueEntity Id = 1
)

type World struct {
	idCounter Id
	archLookup map[Id]ArchId
	archEngine *ArchEngine
}

func NewWorld() *World {
	return &World{
		idCounter: UniqueEntity + 1,
		archLookup: make(map[Id]ArchId),
		archEngine: NewArchEngine(),
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

func Read(world *World, id Id, comp ...interface{}) bool {
	archId, ok := world.archLookup[id]
	if !ok {
		// Entity ID does not exist if it doesn't exist in the bookkeeping
		return false
	}

	lookup := LookupList{}
	ArchRead(world.archEngine, archId, &lookup)
	index, ok := lookup.Lookup[id]
	if !ok { panic("World bookkeeping said entity was here, but lookupList said it isn't") }

	for i := range comp {
		list := world.archEngine.compReg.GetArchStorageType(comp[i])
		ok := ArchRead(world.archEngine, archId, list)
		if !ok { panic("This archetype does not have this component!") } // TODO - return false?
		list.InternalRead(index, comp[i])
	}

	return true
}

func Write(world *World, id Id, comp ...interface{}) {
	archId, ok := world.archLookup[id]
	if ok {
		// The Entity is already constructed, Update the correct archetype
		lookup := LookupList{}
		ArchRead(world.archEngine, archId, &lookup)
		index, ok := lookup.Lookup[id]
		if !ok { panic("World bookkeeping said entity was here, but lookupList said it isn't") }

		for i := range comp {
			list := world.archEngine.compReg.GetArchStorageType(comp[i])
			ok := ArchRead(world.archEngine, archId, list)
			if !ok { panic("Archetype didn't have this component, need to move the entity to a new archetype!") }
			list.InternalWrite(index, comp[i])
			ArchWrite(world.archEngine, archId, list)
		}
	} else {
		// The Entity isn't added yet. Construct it based on components
		archId = world.archEngine.GetArchId(comp...)

		// Update the archetype's lookup with the new entity
		lookup := LookupList{}
		ok := ArchRead(world.archEngine, archId, &lookup)
		if !ok { panic("LookupList is missing!") }
		lookup.Ids = append(lookup.Ids, id)
		index := len(lookup.Ids) - 1
		lookup.Lookup[id] = index
		ArchWrite(world.archEngine, archId, lookup)

		// Update the world's archetype lookup
		world.archLookup[id] = archId

		for i := range comp {
			// Attempt 2
			list := world.archEngine.compReg.GetArchStorageType(comp[i])
			ok := ArchRead(world.archEngine, archId, list)
			if !ok { panic("Archetype didn't have this component!") }
			list.InternalAppend(comp[i])
			if list.Len() != lookup.Len() {
				panic("lookupList length doesn't match component list length!")
			}
			ArchWrite(world.archEngine, archId, list)
		}
		// n := name(comp)
		// world.arch[n].Write(id, )
	}
}

type View struct {
	world *World
	components []interface{}
}

// Returns a view that iterates over all archetypes that contain the designated components
func ViewAll(world *World, comp ...interface{}) View {
	return View{
		world: world,
		components: comp,
	}
}

func (v *View) Map(lambda func(id Id, comp ...interface{})) {
	archIds := ArchFilter(v.world.archEngine, v.components...)

	log.Println("archIds:", archIds)

	compLists := make([]ArchComponent, 0)
	for i := range v.components {
		list := v.world.archEngine.compReg.GetArchStorageType(v.components[i])
		compLists = append(compLists, list)
	}
	log.Println(compLists)

	lookup := LookupList{}
	for _, archId := range archIds {
		// Read Lookup List (which every archetype has)
		ok := ArchRead(v.world.archEngine, archId, &lookup)
		if !ok { panic("LookupList is missing!") }

		// Lookup all component lists for the archetype
		for i := range compLists {
			list := v.world.archEngine.compReg.GetArchStorageType(v.components[i])
			ok := ArchRead(v.world.archEngine, archId, list)
			if !ok { panic("Couldn't find component list for archetype!") }
			compLists[i] = list
		}

		// Execute lambda function with all component lists
		lambdaComps := v.components
		for i := range lookup.Ids {
			for j := range compLists {
				lambdaComps[j] = compLists[j].InternalPointer(i)
			}

			// Execute the function
			lambda(lookup.Ids[i], lambdaComps...)
		}
	}
}

// Archetype lookuplist component
type LookupList struct {
	Lookup map[Id]int
	Ids []Id
}
func (t *LookupList) ComponentSet(val interface{}) { *t = val.(LookupList) }
func (t *LookupList) InternalRead(index int, val interface{}) { *val.(*Id) = t.Ids[index]  }
func (t *LookupList) InternalWrite(index int, val interface{}) { t.Ids[index] = *val.(*Id) }
func (t *LookupList) InternalAppend(val interface{}) { t.Ids = append(t.Ids, val.(Id)) }
func (t *LookupList) InternalPointer(index int) interface{} { return &t.Ids[index] }
func (t *LookupList) Len() int { return len(t.Ids) }
