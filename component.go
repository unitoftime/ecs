package ecs

type CompId uint16

func NewComp[T any]() comp[T] {
	var t T
	return Comp(t)
}

func Comp[T any](t T) comp[T] {
	compId := nameTyped[T](t)
	return comp[T]{
		compId,
	}
}

type comp[T any] struct {
	compId CompId
}

func (c comp[T]) CompId() CompId {
	return c.compId
}
func (c comp[T]) newBox(val T) box[T] {
	return box[T]{
		val:  val,
		comp: c,
	}
}

type W struct {
	engine  *archEngine
	archId  archetypeId
	index   int
	bundler *Bundler
}

type Writer interface {
	CompWrite(W)
}

type Component interface {
	Writer
	CompId() CompId
}

// This type is used to box a component with all of its type info so that it implements the component interface. I would like to get rid of this and simplify the APIs
type box[T any] struct {
	val T
	comp[T]
}

// Creates the boxed component type
func C[T any](val T) box[T] {
	comp := Comp(val)
	return comp.newBox(val)
}

func (c box[T]) CompWrite(wd W) {
	c.WriteVal(wd, c.val)
}

// func (c Box[T]) getPtr(engine *archEngine, archId archetypeId, index int) *T {
// 	store := getStorageByCompId[T](engine, c.Id())
// 	slice := store.slice[archId]
// 	return &slice.comp[index]
// }

// func (c box[T]) With(val T) box[T] {
// 	c.val = val
// 	return c
// }

// func (c box[T]) Unbundle(bun *Bundler) {
// 	c.UnbundleVal(bun, c.comp)
// }

// func (c Box[T]) Unbundle(bun *Bundler) {
// 	compId := c.compId
// 	val := c.Comp
// 	bun.archMask.addComponent(compId)
// 	bun.Set[compId] = true
// 	if bun.Components[compId] == nil {
// 		// Note: We need a pointer so that we dont do an allocation every time we set it
// 		c2 := c // Note: make a copy, so the bundle doesn't contain a pointer to the original
// 		bun.Components[compId] = &c2
// 	} else {
// 		rwComp := bun.Components[compId].(*Box[T])
// 		rwComp.Comp = val
// 	}

// 	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
// }

func (c comp[T]) WriteVal(cw W, val T) {
	if cw.bundler != nil {
		c.UnbundleVal(cw.bundler, val)
	} else {
		store := getStorageByCompId[T](cw.engine, c.CompId())
		writeArch[T](cw.engine, cw.archId, cw.index, store, val)
	}
}

// func (c Box[T]) writeVal(engine *archEngine, archId archetypeId, index int, val T) {
// 	store := getStorageByCompId[T](engine, c.id())
// 	writeArch[T](engine, archId, index, store, val)
// }

func (c comp[T]) UnbundleVal(bun *Bundler, val T) {
	compId := c.compId
	bun.archMask.addComponent(compId)
	bun.Set[compId] = true
	if bun.Components[compId] == nil {
		// Note: We need a pointer so that we dont do an allocation every time we set it
		box := c.newBox(val)
		bun.Components[compId] = &box
	} else {
		rwComp := bun.Components[compId].(*box[T])
		rwComp.val = val
	}

	bun.maxComponentIdAdded = max(bun.maxComponentIdAdded, compId)
}
