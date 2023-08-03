package ecs

type Bundle[T any] struct {
	compId componentId
}

// Createst the boxed component type
func NewBundle[T any]() Bundle[T] {
	var t T
	return Bundle[T]{
		compId: name(t),
	}
}

func (c Bundle[T]) New(comp T) Box[T] {
	return Box[T]{
		Comp: comp,
		compId: c.compId,
	}
}

func (b Bundle[T]) id() componentId {
	return b.compId
}

// type Bundle4[A, B, C, D any] struct {
// 	compId componentId
// 	world *World
// }

// // Createst the boxed component type
// func NewBundle4[A, B, C, D any]() Bundle4[A, B, C, D] {
// 	var a A
// 	var b B
// 	var c C
// 	var d D
// 	return Bundle4[T]{
// 		compIdA: name(a),
// 		compIdB: name(b),
// 		compIdC: name(c),
// 		compIdD: name(d),
// 	}
// }

// func (c Bundle4[A,B,C,D]) Write(id Id, a A, b B, c C, d D) {
// 	Write(c.world, id,
		
// 	)
// }

