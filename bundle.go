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

func (bun *Bundler) Remove(compId CompId) {
	bun.archMask.removeComponent(compId)
	bun.Set[compId] = false
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
	world.writeBundler(id, b)
}
