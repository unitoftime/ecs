package ecs

import (
	"fmt"
	"sync"
)

func nameTyped[T any](comp T) componentId {
	compId := name(comp)
	registerComponentStorage[T](compId)
	return compId
}

type storageBuilder interface {
	build() storage
}
type storageBuilderImp[T any] struct {
}

func (s storageBuilderImp[T]) build() storage {
	return &componentSliceStorage[T]{
		slice: make(map[archetypeId]*componentSlice[T], DefaultAllocation),
	}
}

var componentStorageLookupMut sync.RWMutex
var componentStorageLookup = make(map[componentId]storageBuilder)

func registerComponentStorage[T any](compId componentId) {
	componentStorageLookupMut.Lock()
	_, ok := componentStorageLookup[compId]
	if !ok {
		componentStorageLookup[compId] = storageBuilderImp[T]{}
	}
	componentStorageLookupMut.Unlock()
}

func newComponentStorage(c componentId) storage {
	componentStorageLookupMut.RLock()
	s, ok := componentStorageLookup[c]
	if !ok {
		panic(fmt.Sprintf("tried to build component storage with unregistered componentId: %d", c))
	}

	componentStorageLookupMut.RUnlock()
	return s.build()
}

type componentId uint16

type Component interface {
	SetOther(Component) // TODO: ???
	Clone() Component   // TODO: ???
	write(*archEngine, archetypeId, int)
	id() componentId
}

// This type is used to box a component with all of its type info so that it implements the component interface. I would like to get rid of this and simplify the APIs
type Box[T any] struct {
	Comp   T
	compId componentId
}

// Createst the boxed component type
func C[T any](comp T) Box[T] {
	compId := nameTyped[T](comp)
	return Box[T]{
		Comp:   comp,
		compId: compId,
	}
}
func (c Box[T]) write(engine *archEngine, archId archetypeId, index int) {
	store := getStorageByCompId[T](engine, c.id())
	writeArch[T](engine, archId, index, store, c.Comp)
}
func (c Box[T]) writeVal(engine *archEngine, archId archetypeId, index int, val T) {
	store := getStorageByCompId[T](engine, c.id())
	writeArch[T](engine, archId, index, store, val)
}
func (c Box[T]) getPtr(engine *archEngine, archId archetypeId, index int) *T {
	store := getStorageByCompId[T](engine, c.id())
	slice := store.slice[archId]
	return &slice.comp[index]
}

func (c Box[T]) id() componentId {
	if c.compId == invalidComponentId {
		c.compId = name(c.Comp)
	}
	return c.compId
}

func (c Box[T]) SetOther(other Component) {
	o := other.(Box[T])
	o.Comp = c.Comp
}

func (c Box[T]) Get() T {
	return c.Comp
}

func (c Box[T]) Clone() Component {
	return c
}

func (c Box[T]) UnbundleVal(bun *Bundler, val T) {
	c.Comp = val
	c.Unbundle(bun)
}

func (c Box[T]) Unbundle(bun *Bundler) {
	compId := c.compId
	val := c.Comp
	bun.archMask.addComponent(compId)
	bun.Set[compId] = true
	if bun.Components[compId] == nil {
		// Note: We need a pointer so that we dont do an allocation every time we set it
		c2 := c // Note: make a copy, so the bundle doesn't contain a pointer to the original
		bun.Components[compId] = &c2
	} else {
		rwComp := bun.Components[compId].(*Box[T])
		rwComp.Comp = val
	}

	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
}
