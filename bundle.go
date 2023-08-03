package ecs

type Bundle[T any] struct {
	compId componentId
	storage *componentSliceStorage[T]
	// world *ecs.World //Needed?
}

// Createst the boxed component type
func NewBundle[T any](world *World) Bundle[T] {
	var t T
	compId := name(t)
	return Bundle[T]{
		compId: compId,
		storage: getStorageByCompId[T](world.engine, compId),
	}
}

func (c Bundle[T]) New(comp T) Box[T] {
	return Box[T]{
		Comp: comp,
		compId: c.compId,
		// storage: c.storage,
	}
}

func (b Bundle[T]) write(engine *archEngine, archId archetypeId, index int, comp T) {
	writeArch[T](engine, archId, index, b.storage, comp)
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

