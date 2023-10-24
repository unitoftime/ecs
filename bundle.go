package ecs

type Bundle[T any] struct {
	compId  ComponentId
	storage *componentSliceStorage[T]
	// world *ecs.World //Needed?
}

// Createst the boxed component type
func NewBundle[T any](world *World) Bundle[T] {
	var t T
	compId := name(t)
	return Bundle[T]{
		compId:  compId,
		storage: getStorageByCompId[T](world.engine, compId),
	}
}

func (c Bundle[T]) New(comp T) Box[T] {
	return Box[T]{
		Comp:   comp,
		compId: c.compId,
		// storage: c.storage,
	}
}

func (b Bundle[T]) write(engine *archEngine, archId archetypeId, index int, comp T) {
	writeArch[T](engine, archId, index, b.storage, comp)
}

func (b Bundle[T]) id() ComponentId {
	return b.compId
}

type Bundle4[A, B, C, D any] struct {
	compId ComponentId
	boxA   *Box[A]
	boxB   *Box[B]
	boxC   *Box[C]
	boxD   *Box[D]
	comps  []Component
}

// Createst the boxed component type
func NewBundle4[A, B, C, D any]() Bundle4[A, B, C, D] {
	var a A
	var b B
	var c C
	var d D
	boxA := &Box[A]{a, name(a)}
	boxB := &Box[B]{b, name(b)}
	boxC := &Box[C]{c, name(c)}
	boxD := &Box[D]{d, name(d)}
	comps := []Component{
		boxA, boxB, boxC, boxD,
	}

	return Bundle4[A, B, C, D]{
		boxA:  boxA,
		boxB:  boxB,
		boxC:  boxC,
		boxD:  boxD,
		comps: comps,
	}
}

func (bun Bundle4[A, B, C, D]) Write(world *World, id Id, a A, b B, c C, d D) {
	// bun.boxA.Comp = a
	// bun.boxB.Comp = b
	// bun.boxC.Comp = c
	// bun.boxD.Comp = d
	// bun.comps[0] = bun.boxA
	// bun.comps[1] = bun.boxB
	// bun.comps[2] = bun.boxC
	// bun.comps[3] = bun.boxC

	// bun.comps[0].(*Box[A]).Comp = a
	// bun.comps[1].(*Box[B]).Comp = b
	// bun.comps[2].(*Box[C]).Comp = c
	// bun.comps[3].(*Box[D]).Comp = d

	bun.boxA.Comp = a
	bun.boxB.Comp = b
	bun.boxC.Comp = c
	bun.boxD.Comp = d

	Write(world, id,
		bun.comps...,
	)
}
