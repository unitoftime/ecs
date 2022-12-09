package ecs

import (
	"fmt"
	"math"
)

const (
	InvalidEntity Id = 0
	UniqueEntity Id = 1
	MaxEntity Id = math.MaxUint32
)

type World struct {
	nextId Id
	minId, maxId Id // This is the range of Ids returned by NewId
	arch map[Id]ArchId
	engine *ArchEngine
}

func NewWorld() *World {
	return &World{
		nextId: UniqueEntity + 1,
		minId: UniqueEntity + 1,
		maxId: MaxEntity,
		arch: make(map[Id]ArchId),
		engine: NewArchEngine(),
	}
}

func (w *World) SetIdRange(min, max Id) {
	if min <= UniqueEntity {
		panic("max must be greater than 1")
	}
	if max <= UniqueEntity {
		panic("max must be greater than 1")
	}
	if min > max {
		panic("min must be less than max!")
	}

	w.minId = min
	w.maxId = max
}

func (w *World) NewId() Id {
	if w.nextId < w.minId {
		w.nextId = w.minId
	}

	id := w.nextId

	if w.nextId == w.maxId {
		w.nextId = w.minId
	} else {
		w.nextId++
	}
	return id
}

func (w *World) Count(anything ...any) int {
	return w.engine.Count(anything...)
}

func (w *World) Print(amount int) {
	fmt.Println("--- World ---")
	fmt.Printf("nextId: %d\n", w.nextId)

	// max := amount
	// for id, archId := range w.arch {
	// 	fmt.Printf("id(%d) -> archId(%d)\n", id, archId)
	// 	max--; if max <= 0 { break }
	// }

	// w.engine.Print(amount)
}

// A debug function for describing the current state of memory allocations in the ECS
func (w *World) DescribeMemory() {
	fmt.Println("--- World ---")
	fmt.Printf("nextId: %d\n", w.nextId)
	fmt.Printf("Active Ent Count: %d\n", len(w.arch))
	for archId, lookup := range w.engine.lookup {
		efficiency := 100 * (1.0 - float64(len(lookup.holes))/float64(len(lookup.id)))
		fmt.Printf("Lookup[%d] = {len(index)=%d, len(id)=%d, len(holes)=%d} | Efficiency=%.2f%%\n", archId, len(lookup.index), len(lookup.id), len(lookup.holes), efficiency)
	}
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

func ReadPtr[T any](world *World, id Id) *T {
	archId, ok := world.arch[id]
	if !ok {
		return nil
	}

	return ReadPtrArch[T](world.engine, archId, id)
}

// This is safe for maps and loops
// 1. This deletes the high level id -> archId lookup
// 2. This creates a "hole" in the archetype list
func Delete(world *World, id Id) bool {
	archId, ok := world.arch[id]
	if !ok { return false }

	delete(world.arch, id)

	world.engine.TagForDeletion(archId, id)
	// Note: This was the old, more direct way, but isn't loop safe
	// - world.engine.DeleteAll(archId, id)
	return true
}

func ReadEntity(world *World, id Id) *Entity {
	archId, ok := world.arch[id]
	if !ok { return nil }

	return world.engine.ReadEntity(archId, id)
}
