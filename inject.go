package ecs

import (
	"reflect"
	"runtime"
	"time"
)

func GetInjectable[T any](world *World) T {
	var t T
	name := resourceName(t)

	// 1. If already created, just use this variable
	anyVal, ok := world.resources[name]
	if ok {
		return anyVal.(T)
	}

	// 2. If supports initialization, then make a new one and return it
	tAny := any(t)
	initializer, ok := tAny.(Initializer)
	if ok {
		anyVal = initializer.initialize(world)
		world.resources[name] = anyVal
		return anyVal.(T)
	}

	// 3. Fallback: Just return the default value for whatever it is
	world.resources[name] = t
	return t
}

type Initializer interface {
	initialize(*World) any
}

type SystemBuilder interface {
	Build(world *World) System
}

type System1[A any] struct {
	lambda func(dt time.Duration, a A)
}

func (s System1[A]) Build(world *World) System {
	aRes := GetInjectable[A](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(s.lambda)).Pointer()).Name()
	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			s.lambda(dt, aRes)
		},
	}
}
func NewSystem1[A any](lambda func(dt time.Duration, a A)) System1[A] {
	return System1[A]{
		lambda: lambda,
	}
}

type System2[A, B any] struct {
	lambda func(dt time.Duration, a A, b B)
}

func (s System2[A, B]) Build(world *World) System {
	aRes := GetInjectable[A](world)
	bRes := GetInjectable[B](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(s.lambda)).Pointer()).Name()

	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			s.lambda(dt, aRes, bRes)
		},
	}
}
func NewSystem2[A, B any](lambda func(dt time.Duration, a A, b B)) System2[A, B] {
	return System2[A, B]{
		lambda: lambda,
	}
}

type System3[A, B, C any] struct {
	lambda func(dt time.Duration, a A, b B, c C)
}

func (s System3[A, B, C]) Build(world *World) System {
	aRes := GetInjectable[A](world)
	bRes := GetInjectable[B](world)
	cRes := GetInjectable[C](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(s.lambda)).Pointer()).Name()

	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			s.lambda(dt, aRes, bRes, cRes)
		},
	}
}
func NewSystem3[A, B, C any](lambda func(dt time.Duration, a A, b B, c C)) System3[A, B, C] {
	return System3[A, B, C]{
		lambda: lambda,
	}
}

// func NewSystem4[A, B, C, D any](world *World, lambda func(dt time.Duration, a A, b B, c C, d D)) System {
// 	aRes := GetInjectable[A](world)
// 	bRes := GetInjectable[B](world)
// 	cRes := GetInjectable[C](world)
// 	dRes := GetInjectable[D](world)

// 	systemName := runtime.FuncForPC(reflect.ValueOf(any(lambda)).Pointer()).Name()

// 	return System{
// 		Name: systemName,
// 		Func: func(dt time.Duration) {
// 			lambda(dt, aRes, bRes, cRes, dRes)
// 		},
// 	}
// }
