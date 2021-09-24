package ecs

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
			if !ok {
				//Archetype didn't have this component, move the entity to a new archetype
				moveAndAdd(world, id, comp...)
				return
			} else {
				list.InternalWrite(index, comp[i])
				ArchWrite(world.archEngine, archId, list)
				return
			}
		}
	} else {
		// The Entity isn't added yet. Construct it based on components
		archId = world.archEngine.GetArchId(comp...)

		// Update the archetype's lookup with the new entity
		lookup := &LookupList{}
		ok := ArchRead(world.archEngine, archId, lookup)
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
	}
}

func ReadAll(world *World, id Id) []interface{} {
	archId, ok := world.archLookup[id]
	if !ok {
		return []interface{}{}
	}

	ret := make([]interface{}, 0)

	lookup := LookupList{}
	ok = ArchRead(world.archEngine, archId, &lookup)
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.Lookup[id]
	if !ok { panic("Entity ID doesn't exist in archetype ID!") }

	archComponents := ArchReadAll(world.archEngine, archId)
	for _, archComp := range archComponents {
		ret = append(ret, archComp.InternalRead2(index))
	}

	return ret
}

func Delete(world *World, id Id) {
	archId, ok := world.archLookup[id]
	if !ok { return }

	lookup := LookupList{}
	ok = ArchRead(world.archEngine, archId, &lookup)
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.Lookup[id]
	if !ok { panic("Entity ID doesn't exist in archetype ID!") }

	archComponents := ArchReadAll(world.archEngine, archId)
	for _, archComp := range archComponents {
		archComp.Delete(index)
	}

	delete(world.archLookup, id)
}

func DeleteComponents(world *World, id Id, comp ...interface{}) {
	moveAndRemove(world, id, comp...)
}

func moveAndAdd(world *World, id Id, comp ...interface{}) {
	oldComps := ReadAll(world, id)
	Delete(world, id)

	finalComps := world.overlay(oldComps, comp)

	Write(world, id, finalComps...)
}

func (w *World) overlay(original, overlay []interface{}) []interface{} {
	retMap := make(map[ArchMask]interface{})
	for i := range original {
		_, ok := original[i].(Id)
		if !ok {
			mask := w.archEngine.compReg.GetComponentMask(original[i])
			retMap[mask] = original[i]
		}
	}

	for i := range overlay {
		_, ok := overlay[i].(Id)
		if !ok {
			mask := w.archEngine.compReg.GetComponentMask(overlay[i])
			retMap[mask] = overlay[i]
		}
	}

	ret := make([]interface{}, 0)
	for k := range retMap {
		ret = append(ret, retMap[k])
	}

	return ret
}

func moveAndRemove(world *World, id Id, comp ...interface{}) {
	oldComps := ReadAll(world, id)
	Delete(world, id)

	finalComps := world.unoverlay(oldComps, comp)

	Write(world, id, finalComps...)
}

func (w *World) unoverlay(original, overlay []interface{}) []interface{} {
	retMap := make(map[ArchMask]interface{})
	for i := range original {
		_, ok := original[i].(Id)
		if !ok {
			mask := w.archEngine.compReg.GetComponentMask(original[i])
			retMap[mask] = original[i]
		}
	}

	for i := range overlay {
		_, ok := overlay[i].(Id)
		if !ok {
			mask := w.archEngine.compReg.GetComponentMask(overlay[i])
			delete(retMap, mask)
		}
	}

	ret := make([]interface{}, 0)
	for k := range retMap {
		ret = append(ret, retMap[k])
	}

	return ret
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
func (t *LookupList) ComponentSet(val interface{}) { *t = *val.(*LookupList) }
func (t *LookupList) InternalRead(index int, val interface{}) { *val.(*Id) = t.Ids[index] }
func (t *LookupList) InternalWrite(index int, val interface{}) { t.Ids[index] = *val.(*Id) }
func (t *LookupList) InternalAppend(val interface{}) { t.Ids = append(t.Ids, val.(Id)) }
func (t *LookupList) InternalPointer(index int) interface{} { return &t.Ids[index] }
func (t *LookupList) InternalRead2(index int) interface{} { return t.Ids[index] }
func (t *LookupList) Delete(index int) {
	oldId := t.Ids[index]
	lastVal := t.Ids[len(t.Ids)-1]
	t.Ids[index] = lastVal
	t.Ids = t.Ids[:len(t.Ids)-1]

	// Re-key the map
	t.Lookup[lastVal] = index
	delete(t.Lookup, oldId)
}
func (t *LookupList) Len() int { return len(t.Ids) }
