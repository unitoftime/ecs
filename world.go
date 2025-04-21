package ecs

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"reflect" // For resourceName
)

var (
	DefaultAllocation = 0
)

const (
	InvalidEntity Id = 0 // Represents the default entity Id, which is invalid
	firstEntity   Id = 1
	MaxEntity     Id = math.MaxUint32
)

// World is the main data-holder. You usually pass it to other functions to do things.
type World struct {
	idCounter    atomic.Uint64
	nextId       Id
	minId, maxId Id // This is the range of Ids returned by NewId
	arch         locMap
	engine       *archEngine
	resources    map[reflect.Type]any
	observers    *internalMap[EventId, list[Handler]] // TODO: SliceMap instead of map
	cmd *CommandQueue
}

// Creates a new world
func NewWorld() *World {
	world := &World{
		nextId: firstEntity + 1,
		minId:  firstEntity + 1,
		maxId:  MaxEntity,
		arch:   newLocMap(DefaultAllocation),
		engine: newArchEngine(),

		resources: make(map[reflect.Type]any),
		observers: newMap[EventId, list[Handler]](0),
	}

	world.cmd = GetInjectable[*CommandQueue](world)

	return world
}

func (w *World) print() {
	fmt.Printf("%+v\n", *w)

	w.engine.print()
}

// Sets an range of Ids that the world will use when creating new Ids. Potentially helpful when you have multiple worlds and don't want their Id space to collide.
// Deprecated: This API is tentative. It may be better to just have the user create Ids as they see fit
func (w *World) SetIdRange(min, max Id) {
	if min <= firstEntity {
		panic("max must be greater than 1")
	}
	if max <= firstEntity {
		panic("max must be greater than 1")
	}
	if min > max {
		panic("min must be less than max!")
	}

	w.minId = min
	w.maxId = max
}

// Creates a new Id which can then be used to create an entity. This is threadsafe
func (w *World) NewId() Id {
	for {
		val := w.idCounter.Load()
		if w.idCounter.CompareAndSwap(val, val+1) {
			return (Id(val) % (w.maxId - w.minId)) + w.minId
		}
	}

	// if w.nextId < w.minId {
	// 	w.nextId = w.minId
	// }

	// id := w.nextId

	// if w.nextId == w.maxId {
	// 	w.nextId = w.minId
	// } else {
	// 	w.nextId++
	// }
	// return id
}

func (world *World) Spawn(comp ...Component) Id {
	id := world.NewId()
	world.spawn(id, comp...)
	return id
}

func (world *World) spawn(id Id, comp ...Component) {
	// Id does not yet exist, we need to add it for the first time
	mask := buildArchMask(comp...)
	archId := world.engine.getArchetypeId(mask)

	// Write all components to that archetype
	index := world.engine.spawn(archId, id, comp...)
	world.arch.Put(id, entLoc{archId, uint32(index)})

	world.engine.runFinalizedHooks(id)
}

// Writes components to the entity specified at id. This API can potentially break if you call it inside of a loop. Specifically, if you cause the archetype of the entity to change by writing a new component, then the loop may act in mysterious ways.
// Deprecated: This API is tentative, I might replace it with something similar to bevy commands to alleviate the above concern
func Write(world *World, id Id, comp ...Component) {
	world.Write(id, comp...)
}

func (world *World) Write(id Id, comp ...Component) {
	if len(comp) <= 0 {
		return // Do nothing if there are no components
	}

	loc, ok := world.arch.Get(id)
	if ok {
		newLoc := world.engine.rewriteArch(loc, id, comp...)
		world.arch.Put(id, newLoc)
	} else {
		world.spawn(id, comp...)
	}

	world.engine.runFinalizedHooks(id)
}

func (w *World) writeBundler(id Id, b *Bundler) {
	newLoc := w.allocateMove(id, b.archMask)

	wd := W{
		engine: w.engine,
		archId: newLoc.archId,
		index:  int(newLoc.index),
	}

	for i := CompId(0); i <= b.maxComponentIdAdded; i++ {
		if !b.Set[i] {
			continue
		}

		b.Components[i].CompWrite(wd)
	}

	w.engine.runFinalizedHooks(id)
}

// func (world *World) GetArchetype(comp ...Component) archetypeId {
// 	mask := buildArchMask(comp...)
// 	return world.engine.getArchetypeId(mask)
// }

// // Note: This returns the index of the location allocated
// func (world *World) Allocate(id Id, archId archetypeId) int {
// 	return world.allocate(id, world.engine.dcr.revArchMask[archId])
// }

// Allocates an index for the id at the specified addMask location
// 1. If the id already exists, an archetype move will happen
// 2. If the id doesn't exist, then the addMask is the newMask and the entity will be allocated there
// Returns the index of the location allocated. May return -1 if invalid archMask supplied
func (world *World) allocateMove(id Id, addMask archetypeMask) entLoc {
	if addMask == blankArchMask {
		// Nothing to allocate, aka do nothing
		loc, _ := world.arch.Get(id)
		// TODO: Technically this is some kind of error if id isn't set
		return loc
	}

	loc, ok := world.arch.Get(id)
	if ok {
		// Calculate the new mask based on the bitwise or of the old and added masks
		lookup := world.engine.lookup[loc.archId]
		oldMask := lookup.mask
		newMask := oldMask.bitwiseOr(addMask)

		// If the new mask matches the old mask, then we don't need to move anything
		if oldMask == newMask {
			return loc
		}

		newLoc := world.engine.moveArchetype(loc, newMask, id)
		world.arch.Put(id, newLoc)

		world.engine.finalizeOnAdd = markComponentDiff(world.engine.finalizeOnAdd, addMask, oldMask)

		return newLoc
	} else {
		// Id does not yet exist, we need to add it for the first time
		archId := world.engine.getArchetypeId(addMask)
		// Write all components to that archetype
		newIndex := world.engine.allocate(archId, id)

		newLoc := entLoc{archId, uint32(newIndex)}
		world.arch.Put(id, newLoc)

		world.engine.finalizeOnAdd = markComponentMask(world.engine.finalizeOnAdd, addMask)

		return newLoc
	}
}

// May return -1 if invalid archMask supplied, or if the entity doesn't exist
func (world *World) deleteMask(id Id, deleteMask archetypeMask) {
	loc, ok := world.arch.Get(id)
	if !ok {
		return
	}

	// 1. calculate the destination mask
	lookup := world.engine.lookup[loc.archId]
	oldMask := lookup.mask
	newMask := oldMask.bitwiseClear(deleteMask)

	// If the new mask requires the removal of all components, then just delete the current entity
	if newMask == blankArchMask {
		Delete(world, id)
		return
	}

	// If  the new mask matches the old mask, then we don't need to move anything
	if oldMask == newMask {
		return
	}

	// 2. Move all components from source arch to dest arch
	newLoc := world.engine.moveArchetypeDown(loc, newMask, id)
	world.arch.Put(id, newLoc)
}

// This is safe for maps and loops
// 1. This deletes the high level id -> archId lookup
// 2. This creates a "hole" in the archetype list
// Returns true if the entity was deleted, else returns false if the entity does not exist (or was already deleted)

// Deletes the entire entity specified by the id
// This can be called inside maps and loops, it will delete the entity immediately.
// Returns true if the entity exists and was actually deleted, else returns false
func Delete(world *World, id Id) bool {
	archId, ok := world.arch.Get(id)
	if !ok {
		return false
	}

	world.arch.Delete(id)

	world.engine.TagForDeletion(archId, id)
	return true
}

// Deletes specific components from an entity in the world
// Skips all work if the entity doesn't exist
// Skips deleting components that the entity doesn't have
// If no components remain after the delete, the entity will be completely removed
func DeleteComponent(world *World, id Id, comp ...Component) {
	if len(comp) <= 0 {
		return
	}

	mask := buildArchMask(comp...)
	world.deleteMask(id, mask)
}

// Returns true if the entity exists in the world else it returns false
func (world *World) Exists(id Id) bool {
	return world.arch.Has(id)
}

// --------------------------------------------------------------------------------
// - Observers
// --------------------------------------------------------------------------------
func (w *World) Trigger(event Event, id Id) {
	handlerList, ok := w.observers.Get(event.EventId())
	if !ok {
		return
	}
	for _, handler := range handlerList.list {
		handler.Run(id, event)
	}
}

func (w *World) AddObserver(handler Handler) {
	handlerList, ok := w.observers.Get(handler.EventTrigger())
	if !ok {
		handlerList = newList[Handler]()
	}

	handlerList.Add(handler)
	w.observers.Put(handler.EventTrigger(), handlerList)
}

// You may only register one hook per component, else it will panic
func (w *World) SetHookOnAdd(comp Component, handler Handler) {
	current := w.engine.onAddHooks[comp.CompId()]
	if current != nil {
		panic("AddHook: You may only register one hook per component")
	}
	w.engine.onAddHooks[comp.CompId()] = handler
}

// --------------------------------------------------------------------------------
// - Resources
// --------------------------------------------------------------------------------
func resourceName(t any) reflect.Type {
	return reflect.TypeOf(t)
}

// TODO: Should I force people to do pointers?
func PutResource[T any](world *World, resource *T) {
	name := resourceName(resource)
	world.resources[name] = resource
}

func GetResource[T any](world *World) *T {
	var t T
	name := resourceName(&t)
	anyVal, ok := world.resources[name]
	if !ok {
		return nil
	}

	return anyVal.(*T)
}

// --------------------------------------------------------------------------------
// - Systems
// --------------------------------------------------------------------------------
func (w *World) StepSystemList(dt time.Duration, systems ...System) time.Duration {
	start := time.Now()
	for i := range systems {
		systems[i].step(dt)
		w.cmd.Execute()
	}
	return time.Since(start)
}
