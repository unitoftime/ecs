package ecs

import (
	"reflect"
	"sync"
)

var componentIdMutex sync.Mutex
var registeredComponents = make(map[reflect.Type]componentId, maxComponentId)
var invalidComponentId componentId = 0
var componentRegistryCounter componentId = 1

func name(t any) componentId {
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
