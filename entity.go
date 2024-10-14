package ecs

// An Entity is essentially a map of components that is held external to a world. Useful for pulling full entities in and out of the world.
// Deprecated: This type and its corresponding methods are tentative and might be replaced by something else.
type Entity struct {
	// comp map[componentId]Component
	comp []Component
}

// Creates a new entity with the specified components
func NewEntity(components ...Component) *Entity {
	return &Entity{
		comp: components,
	}

	// c := make(map[componentId]Component)
	// for i := range components {
	// 	c[components[i].id()] = components[i]
	// }
	// return &Entity{
	// 	comp: c,
	// }
}

// Returns the index that contains the same componentId or returns -1
func (e *Entity) findIndex(compId componentId) int {
	for i := range e.comp {
		if compId == e.comp[i].id() {
			return i
		}
	}

	return -1
}

// Adds a component to an entity
func (e *Entity) Add(components ...Component) {
	for i := range components {
		idx := e.findIndex(components[i].id())
		if idx < 0 {
			e.comp = append(e.comp, components[i])
		} else {
			e.comp[idx] = components[i]
		}
	}

	// for i := range components {
	// 	e.comp[components[i].id()] = components[i]
	// }
}

// Merges e2 on top of e (meaning that we will overwrite e with values from e2)
func (e *Entity) Merge(e2 *Entity) {
	e.Add(e2.comp...)

	// for k, v := range e2.comp {
	// 	e.comp[k] = v
	// }
}

// Returns a list of the components held by the entity
func (e *Entity) Comps() []Component {
	return e.comp

	// ret := make([]Component, 0, len(e.comp))
	// for _, v := range e.comp {
	// 	ret = append(ret, v)
	// }
	// return ret
}

// Reads a specific component from the entity, returns false if the component doesn't exist
func ReadFromEntity[T any](ent *Entity) (T, bool) {
	var t T
	n := name(t)
	idx := ent.findIndex(n)
	if idx < 0 {
		return t, false
	}

	icomp := ent.comp[idx]
	return icomp.(Box[T]).Comp, true

	// var t T
	// n := name(t)

	// icomp, ok := ent.comp[n]
	// if !ok {
	// 	return t, false
	// }
	// return icomp.(Box[T]).Comp, true
}

// Writes the entire entity to the world
func (ent *Entity) Write(world *World, id Id) {
	world.Write(id, ent.comp...)

	// comps := ent.Comps()
	// world.Write(id, comps...)
}

// Reads the entire entity out of the world and into an *Entity object. Returns nil if the entity doesn't exist
func ReadEntity(world *World, id Id) *Entity {
	archId, ok := world.arch.Get(id)
	if !ok {
		return nil
	}

	return world.engine.ReadEntity(archId, id)
}

// Deletes a component on this entity
func (e *Entity) Delete(c Component) {
	compId := c.id()
	idx := e.findIndex(compId)
	if idx < 0 {
		return
	}

	// If index does exist, then cut it out
	e.comp[idx] = e.comp[len(e.comp)-1]
	e.comp = e.comp[:len(e.comp)-1]

	// delete(e.comp, c.id())
}

// Clears the map, but retains the space
func (e *Entity) Clear() {
	e.comp = e.comp[:0]

	// // Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	// for k := range e.comp {
	// 	delete(e.comp, k)
	// }
}

// TODO revisit this abstraction
// type Copier interface {
// 	Copy() interface{}
// }

// func (e Entity) Copy() Entity {
// 	copy := BlankEntity()
// 	for k,v := range e {
// 		c, ok := v.(Copier)
// 		if ok {
// 			// fmt.Println("Copying:", k)
// 			// If the component has a custom copy interface, then copy it
// 			copy[k] = c.Copy()
// 		} else {
// 			// Else just copy the top level struct
// 			copy[k] = v
// 		}
// 	}
// 	return copy
// }

// A RawEntity is like an Entity, but every component is actually a pointer to the underlying component. I mostly use this to build inspector UIs that can directly modify an entity
// Deprecated: This type and its corresponding methods are tentative and might be replaced by something else.
type RawEntity struct {
	comp map[componentId]any
}

// Creates a new entity with the specified components
func NewRawEntity(components ...any) *RawEntity {
	c := make(map[componentId]any)
	for i := range components {
		c[name(components[i])] = components[i]
	}
	return &RawEntity{
		comp: c,
	}
}

// Adds a component to an entity
func (e *RawEntity) Add(components ...any) {
	for i := range components {
		e.comp[name(components[i])] = components[i]
	}
}

// Merges e2 on top of e (meaning that we will overwrite e with values from e2)
func (e *RawEntity) Merge(e2 *RawEntity) {
	for k, v := range e2.comp {
		e.comp[k] = v
	}
}

// Returns a list of the components held by the entity
func (e *RawEntity) Comps() []any {
	ret := make([]any, 0, len(e.comp))
	for _, v := range e.comp {
		ret = append(ret, v)
	}
	return ret
}

// // Reads a specific component from the entity, returns false if the component doesn't exist
// func ReadFromRawEntity[T any](ent *RawEntity) (T, bool) {
// 	var t T
// 	n := name(t)

// 	icomp, ok := ent.comp[n]
// 	if !ok {
// 		return t, false
// 	}
// 	return icomp.(Box[T]).Comp, true
// }

// Writes the entire entity to the world
// func (ent *RawEntity) Write(world *World, id Id) {
// 	comps := ent.Comps()
// 	world.Write(id, comps...)
// }

// Reads the entire entity out of the world and into an *RawEntity object. Returns nil if the entity doesn't exist. RawEntity is lik
func ReadRawEntity(world *World, id Id) *RawEntity {
	archId, ok := world.arch.Get(id)
	if !ok {
		return nil
	}

	return world.engine.ReadRawEntity(archId, id)
}

// Deletes a component on this entity
func (e *RawEntity) Delete(c Component) {
	delete(e.comp, name(c))
}

// Clears the map, but retains the space
func (e *RawEntity) Clear() {
	// Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	for k := range e.comp {
		delete(e.comp, k)
	}
}
