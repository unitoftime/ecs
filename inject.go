package ecs

import (
	"reflect"
	"runtime"
	"time"
)

type Initializer interface {
	initialize(*World) any
}

func NewSystem1[A Initializer](world *World, lambda func(dt time.Duration, a A)) System {
	aRes := GetInjectable[A](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(lambda)).Pointer()).Name()
	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			lambda(dt, aRes)
		},
	}
}

func NewSystem2[A, B Initializer](world *World, lambda func(dt time.Duration, a A, b B)) System {
	aRes := GetInjectable[A](world)
	bRes := GetInjectable[B](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(lambda)).Pointer()).Name()

	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			lambda(dt, aRes, bRes)
		},
	}
}

func NewSystem3[A, B, C Initializer](world *World, lambda func(dt time.Duration, a A, b B, c C)) System {
	aRes := GetInjectable[A](world)
	bRes := GetInjectable[B](world)
	cRes := GetInjectable[C](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(lambda)).Pointer()).Name()

	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			lambda(dt, aRes, bRes, cRes)
		},
	}
}

func NewSystem4[A, B, C, D Initializer](world *World, lambda func(dt time.Duration, a A, b B, c C, d D)) System {
	aRes := GetInjectable[A](world)
	bRes := GetInjectable[B](world)
	cRes := GetInjectable[C](world)
	dRes := GetInjectable[D](world)

	systemName := runtime.FuncForPC(reflect.ValueOf(any(lambda)).Pointer()).Name()

	return System{
		Name: systemName,
		Func: func(dt time.Duration) {
			lambda(dt, aRes, bRes, cRes, dRes)
		},
	}
}
