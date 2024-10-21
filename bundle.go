package ecs

type Bundle[T any] struct {
	compId  componentId
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

func (b Bundle[T]) id() componentId {
	return b.compId
}

type Bundle4[A, B, C, D any] struct {
	compId componentId
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

// // --------------------------------------------------------------------------------
// type BundleTry2[A, B, C, D any] struct {
// 	query *View4[A, B, C, D]
// 	compId componentId
// 	boxA   *Box[A]
// 	boxB   *Box[B]
// 	boxC   *Box[C]
// 	boxD   *Box[D]
// 	comps  []Component
// }

// // Createst the boxed component type
// func NewBundleTry2[A, B, C, D any](world *World) BundleTry2[A, B, C, D] {
// 	var a A
// 	var b B
// 	var c C
// 	var d D
// 	boxA := &Box[A]{a, name(a)}
// 	boxB := &Box[B]{b, name(b)}
// 	boxC := &Box[C]{c, name(c)}
// 	boxD := &Box[D]{d, name(d)}
// 	comps := []Component{
// 		boxA, boxB, boxC, boxD,
// 	}

// 	return BundleTry2[A, B, C, D]{
// 		query: Query4[A, B, C, D](world),
// 		boxA:  boxA,
// 		boxB:  boxB,
// 		boxC:  boxC,
// 		boxD:  boxD,
// 		comps: comps,
// 	}
// }

// // Step 1: Allocate arch Id
// // Step 2: Read pointers
// // Step 3: Write everything
// func (bun BundleTry2[A, B, C, D]) Write(id Id, a *A, b *B, c *C, d *D) {
// 	var archMask archetypeMask
// 	bun.addToArchMask(&archMask, a, b, c, d)

// 	world.allocate(id, archMask)

// 	aDst, bDst, cDst, dDst := bun.query.Read(id)
// 	if a != nil {
// 		*aDst = *a
// 	}
// 	if b != nil {
// 		*bDst = *b
// 	}
// 	if c != nil {
// 		*cDst = *c
// 	}
// 	if d != nil {
// 		*dDst = *d
// 	}

// 	// bun.boxA.Comp = *a
// 	// bun.boxB.Comp = *b
// 	// bun.boxC.Comp = *c
// 	// bun.boxD.Comp = *d

// 	// Write(world, id,
// 	// 	bun.comps...,
// 	// )
// }

// func (bun BundleTry2[A, B, C, D]) addToArchMask(archMask *archetypeMask, a *A, b *B, c *C, d *D) {
// 	if a != nil {
// 		archMask.addComponent(bun.boxA.compId)
// 	}
// 	if b != nil {
// 		archMask.addComponent(bun.boxB.compId)
// 	}
// 	if c != nil {
// 		archMask.addComponent(bun.boxC.compId)
// 	}
// 	if d != nil {
// 		archMask.addComponent(bun.boxD.compId)
// 	}
// }
