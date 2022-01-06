package ecs

import (
	"fmt"
)

type Component interface {
	Write(*ArchEngine, ArchId, Id)
	Name() string
}
// TODO -I could get rid of reflect if there ends up being some way to compile-time reflect on generics
type CompBox[T any] struct {
	comp T
}
func C[T any](comp T) CompBox[T] {
	return CompBox[T]{comp}
}
func (c CompBox[T]) Write(engine *ArchEngine, archId ArchId, id Id) {
	WriteArch[T](engine, archId, id, c.comp)
}
func (c CompBox[T]) Name() string {
	return name(c.comp)
}

func (c CompBox[T]) Get() T {
	return c.comp
}

const (
	InvalidEntity Id = 0
	UniqueEntity Id = 1
)

type World struct {
	idCounter Id
	arch map[Id]ArchId
	engine *ArchEngine
}

func NewWorld() *World {
	return &World{
		idCounter: UniqueEntity + 1,
		arch: make(map[Id]ArchId),
		engine: NewArchEngine(),
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
	w.engine.Print()
}

// TODO - Note: This function is not safe inside Maps or view iteraions
// TODO - make this loop-safe by:
// 1. Read the entire entity into an entity object
// 2. Call loop-safe delete method on that ID (which tags it somehow to indicate it needs to be cleaned up)
// 3. Modify the entity object by removing the requested components
// 4. Write the entity object to the destination archetype
// 4.a If the destination archetype is currently locked/flagged to indicate we are looping over it then wait for the lock release before writing the entity
// 4.b When creating Maps and Views we need to lock each archId that needs to be processed. Notably this guarantees that all "Writes" to this ArchId will be done AFTER the lambda has processed - Meaning that we won't execute the same entity twice.
// 4.b.i When creating a view I may need like a "Close" method or "end" or something otherwise I'm not sure how to unlock the archId for modification
// Question: Why not write directly to holes if possible?
func Write(world *World, id Id, comp ...Component) {
	archId, ok := world.arch[id]
	if ok {
		newArchId := world.engine.RewriteArch(archId, id, comp...)
		world.arch[id] = newArchId
	} else {
		// Id does not yet exist, we need to add it for the first time
		archId = world.engine.GetArchId(comp...)
		world.arch[id] = archId

		// Write all components to that archetype
		// TODO - Push this inward for efficiency?
		for i := range comp {
			comp[i].Write(world.engine, archId, id)
		}
	}
}

func Read[T any](world *World, id Id) (T, bool) {
	var ret T
	archId, ok := world.arch[id]
	if !ok {
		return ret, false
	}

	return ReadArch[T](world.engine, archId, id)
}

// This is safe for maps and loops
// 1. This deletes the high level id -> archId lookup
// 2. This creates a "hole" in the archetype list
func Delete(world *World, id Id) {
	archId, ok := world.arch[id]
	if !ok { return }

	delete(world.arch, id)

	world.engine.TagForDeletion(archId, id)
	// Note: This was the old, more direct way, but isn't loop safe
	// - world.engine.DeleteAll(archId, id)
}
