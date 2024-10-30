package ecs

import (
	"fmt"
	"reflect"
	"sync"
)

func nameTyped[T any](comp T) CompId {
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
var componentStorageLookup = make(map[CompId]storageBuilder)

func registerComponentStorage[T any](compId CompId) {
	componentStorageLookupMut.Lock()
	_, ok := componentStorageLookup[compId]
	if !ok {
		componentStorageLookup[compId] = storageBuilderImp[T]{}
	}
	componentStorageLookupMut.Unlock()
}

func newComponentStorage(c CompId) storage {
	componentStorageLookupMut.RLock()
	s, ok := componentStorageLookup[c]
	if !ok {
		panic(fmt.Sprintf("tried to build component storage with unregistered componentId: %d", c))
	}

	componentStorageLookupMut.RUnlock()
	return s.build()
}

//--------------------------------------------------------------------------------

var componentIdMutex sync.Mutex
var registeredComponents = make(map[reflect.Type]CompId, maxComponentId)
var invalidComponentId CompId = 0
var componentRegistryCounter CompId = 1

func name(t any) CompId {
	// Note: We have to lock here in case there are multiple worlds
	// TODO!! - This probably causes some performance penalty
	componentIdMutex.Lock()
	defer componentIdMutex.Unlock()

	typeof := reflect.TypeOf(t)
	compId, ok := registeredComponents[typeof]
	if !ok {
		compId = componentRegistryCounter
		registeredComponents[typeof] = compId
		componentRegistryCounter++
	}
	return compId
}

// // Possible solution: Runs faster than reflection (mostly useful for potentially removing/reducing ecs.C(...) overhead
// import (
// 	"sync"
// 	"unsafe"
// )

// type emptyInterface struct {
//     typ unsafe.Pointer
//     ptr unsafe.Pointer
// }

// var componentIdMutex sync.Mutex
// var registeredComponents = make(map[uintptr]componentId, maxComponentId)
// var invalidComponentId componentId = 0
// var componentRegistryCounter componentId = 1

// func name(t any) componentId {
// 	// Note: We have to lock here in case there are multiple worlds
// 	// TODO!! - This probably causes some performance penalty
// 	componentIdMutex.Lock()
// 	defer componentIdMutex.Unlock()

// 	iface := (*emptyInterface)(unsafe.Pointer(&t))
// 	typeptr := uintptr(iface.typ)
// 	compId, ok := registeredComponents[typeptr]
// 	if !ok {
// 		compId = componentRegistryCounter
// 		registeredComponents[typeptr] = compId
// 		componentRegistryCounter++
// 	}
// 	return compId
// }
