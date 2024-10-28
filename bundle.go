package ecs

// type Bundle[T any] struct {
// 	compId  componentId
// 	storage *componentSliceStorage[T]
// 	// world *ecs.World //Needed?
// }

// // Createst the boxed component type
// func NewBundle[T any](world *World) Bundle[T] {
// 	var t T
// 	compId := name(t)
// 	return Bundle[T]{
// 		compId:  compId,
// 		storage: getStorageByCompId[T](world.engine, compId),
// 	}
// }

// func (c Bundle[T]) New(comp T) Box[T] {
// 	return Box[T]{
// 		Comp:   comp,
// 		compId: c.compId,
// 		// storage: c.storage,
// 	}
// }

// func (b Bundle[T]) write(engine *archEngine, archId archetypeId, index int, comp T) {
// 	writeArch[T](engine, archId, index, b.storage, comp)
// }

// func (b Bundle[T]) id() componentId {
// 	return b.compId
// }

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
	bun.boxA.Comp = a
	bun.boxB.Comp = b
	bun.boxC.Comp = c
	bun.boxD.Comp = d

	Write(world, id,
		bun.comps...,
	)
}

func (bun Bundle4[A, B, C, D]) Unbundle(bundler *Bundler, a A, b B, c C, d D) {
	bun.boxA.UnbundleVal(bundler, a)
	bun.boxB.UnbundleVal(bundler, b)
	bun.boxC.UnbundleVal(bundler, c)
	bun.boxD.UnbundleVal(bundler, d)
}

// --------------------------------------------------------------------------------
// type BundleTry2[A, B, C, D any] struct {
// 	world *World
// 	query *View4[A, B, C, D]
// 	boxA  *Box[A]
// 	boxB  *Box[B]
// 	boxC  *Box[C]
// 	boxD  *Box[D]
// }

// // Createst the boxed component type
// func NewBundleTry2[A, B, C, D any](world *World) BundleTry2[A, B, C, D] {
// 	var a A
// 	var b B
// 	var c C
// 	var d D
// 	boxA := &Box[A]{a, nameTyped(a)}
// 	boxB := &Box[B]{b, nameTyped(b)}
// 	boxC := &Box[C]{c, nameTyped(c)}
// 	boxD := &Box[D]{d, nameTyped(d)}

// 	return BundleTry2[A, B, C, D]{
// 		world: world,
// 		query: Query4[A, B, C, D](world),
// 		boxA:  boxA,
// 		boxB:  boxB,
// 		boxC:  boxC,
// 		boxD:  boxD,
// 	}
// }

// // Step 1: Allocate arch Id
// // Step 2: Read pointers
// // Step 3: Write everything
// func (bun BundleTry2[A, B, C, D]) Write(id Id, a *A, b *B, c *C, d *D) {
// 	var archMask archetypeMask
// 	bun.addToArchMask(&archMask, a, b, c, d)
// 	archId := bun.world.engine.dcr.getArchetypeId(bun.world.engine, archMask)

// 	bun.world.Allocate(id, archId)

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

type Bundler struct {
	archMask            archetypeMask             // The current archetypeMask
	Set                 [maxComponentId]bool      // The list of components that are being bundled
	Components          [maxComponentId]Component // Component storage array for everything we've bundled
	maxComponentIdAdded componentId
}

func (b *Bundler) Clear() {
	b.archMask = blankArchMask
	b.Set = [maxComponentId]bool{}
	b.maxComponentIdAdded = 0
	// b.Components // Note: No need to clear because we only use set values
}

func (bun *Bundler) Add(comp Component) {
	compId := comp.id()
	bun.archMask.addComponent(compId)
	bun.Set[compId] = true
	if bun.Components[compId] == nil {
		bun.Components[compId] = comp.Clone() // Create an internal copy
	} else {
		comp.SetOther(bun.Components[compId])
	}

	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
}

// func WriteComponent[T any](bun *Bundler, comp T) {
// 	compId := nameTyped(comp)
// 	bun.archMask.addComponent(compId)
// 	bun.Set[compId] = true
// 	if bun.Components[compId] == nil {
// 		bun.Components[compId] = C(comp)
// 	} else {
// 		bun.Components[compId].Set(comp)
// 	}

// 	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
// }

func (b *Bundler) Write(world *World, id Id) {
	archId := world.engine.getArchetypeId(b.archMask)
	archMask := world.engine.dcr.revArchMask[archId]
	index := world.allocate(id, archMask)

	for i := componentId(0); i < b.maxComponentIdAdded; i++ {
		if !b.Set[i] {
			continue
		}

		b.Components[i].write(world.engine, archId, index)
	}
}
