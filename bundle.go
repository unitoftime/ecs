package ecs

type Bundler struct {
	archMask archetypeMask // The current archetypeMask
	// TODO: Instead of set, you could just use the arch mask
	Set                 [maxComponentId]bool      // The list of components that are being bundled
	Components          [maxComponentId]Component // Component storage array for everything we've bundled
	maxComponentIdAdded CompId
}

func (b *Bundler) Clear() {
	b.archMask = blankArchMask
	b.Set = [maxComponentId]bool{}
	b.maxComponentIdAdded = 0
	// b.Components // Note: No need to clear because we only use set values
}

// func (bun *Bundler) Add(comp Component) {
// 	compId := comp.id()
// 	bun.archMask.addComponent(compId)
// 	bun.Set[compId] = true
// 	if bun.Components[compId] == nil {
// 		bun.Components[compId] = comp.Clone() // Create an internal copy
// 	} else {
// 		comp.SetOther(bun.Components[compId])
// 	}

// 	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
// }

// func (bun *Bundler) Has(comp Component) bool {
// 	return bun.Set[comp.CompId()]
// }

func readBundle[T Component](bun *Bundler) (T, bool) {
	var comp T
	compId := comp.CompId()

	if !bun.Set[compId] {
		return comp, false // Was not set
	}
	return bun.Components[compId].(*box[T]).val, true
}

// func (bun *Bundler) Read(comp Component) (Component, bool) {
// 	compId := comp.CompId()
// 	if !bun.Set[compId] {
// 		return comp, false // Was not set
// 	}
// 	return bun.Components[compId], true
// }

// func (bun *Bundler) Remove(comp Component) {
// 	compId := comp.id()
// 	bun.archMask.removeComponent(compId)
// 	bun.Set[compId] = true
// }

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
	newLoc := world.allocateMove(id, b.archMask)

	wd := W{
		engine: world.engine,
		archId: newLoc.archId,
		index:  int(newLoc.index),
	}

	for i := CompId(0); i <= b.maxComponentIdAdded; i++ {
		if !b.Set[i] {
			continue
		}

		b.Components[i].CompWrite(wd)
	}
}

//--------------------------------------------------------------------------------

// Note: This is slightly more optimized than passing a list in every time because no allocs are needed
// type Bundle4[A, B, C, D any] struct {
// 	compId CompId
// 	boxA   *box[A]
// 	boxB   *box[B]
// 	boxC   *box[C]
// 	boxD   *box[D]
// 	comps  []Component
// }

// // Createst the boxed component type
// func NewBundle4[A, B, C, D any]() Bundle4[A, B, C, D] {
// 	// var a A
// 	// var b B
// 	// var c C
// 	// var d D

// 	// boxA := &box[A]{a, name(a)}
// 	// boxB := &box[B]{b, name(b)}
// 	// boxC := &box[C]{c, name(c)}
// 	// boxD := &box[D]{d, name(d)}
// 	var a A
// 	var b B
// 	var c C
// 	var d D

// 	boxA := Comp(a)
// 	boxB := Comp(b)
// 	boxC := Comp(c)
// 	boxD := Comp(d)

// 	comps := []Component{
// 		&boxA, &boxB, &boxC, &boxD,
// 	}

// 	return Bundle4[A, B, C, D]{
// 		boxA:  &boxA,
// 		boxB:  &boxB,
// 		boxC:  &boxC,
// 		boxD:  &boxD,
// 		comps: comps,
// 	}
// }

// func (bun Bundle4[A, B, C, D]) Write(world *World, id Id, a A, b B, c C, d D) {
// 	bun.boxA.comp = a
// 	bun.boxB.comp = b
// 	bun.boxC.comp = c
// 	bun.boxD.comp = d

// 	Write(world, id,
// 		bun.comps...,
// 	)
// }

// func (bun Bundle4[A, B, C, D]) Unbundle(bundler *Bundler, a A, b B, c C, d D) {
// 	bun.boxA.UnbundleVal(bundler, a)
// 	bun.boxB.UnbundleVal(bundler, b)
// 	bun.boxC.UnbundleVal(bundler, c)
// 	bun.boxD.UnbundleVal(bundler, d)
// }

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
