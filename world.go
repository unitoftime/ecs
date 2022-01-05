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

// TODO - Variadic
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
