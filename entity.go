package ecs

// An Entity is essentially a map of components that is held external to a world. Useful for pulling full entities in and out of the world.
// Deprecated: This type and its corresponding methods are tentative and might be replaced by something else.
type Entity struct {
	comp map[componentId]Component
}

// Creates a new entity with the specified components
func NewEntity(components ...Component) *Entity {
	c := make(map[componentId]Component)
	for i := range components {
		c[components[i].id()] = components[i]
	}
	return &Entity{
		comp: c,
	}
}

// Adds a component to an entity
func (e *Entity) Add(components ...Component) {
	for i := range components {
		e.comp[components[i].id()] = components[i]
	}
}

// Returns a list of the components held by the entity
func (e *Entity) Comps() []Component {
	ret := make([]Component, 0, len(e.comp))
	for _, v := range e.comp {
		ret = append(ret, v)
	}
	return ret
}

// Reads a specific component from the entity, returns false if the component doesn't exist
func ReadFromEntity[T any](ent *Entity) (T, bool) {
	var t T
	n := name(t)

	icomp, ok := ent.comp[n]
	if !ok {
		return t, false
	}
	return icomp.(Box[T]).Comp, true
}

// Writes the entire entity to the world
func (ent *Entity) Write(world *World, id Id) {
	comps := ent.Comps()
	world.Write(id, comps...)
}

// Reads the entire entity out of the world and into an *Entity object. Returns nil if the entity doesn't exist
func ReadEntity(world *World, id Id) *Entity {
	archId, ok := world.arch[id]
	if !ok {
		return nil
	}

	return world.engine.ReadEntity(archId, id)
}

// Deletes a component on this entity
func (e *Entity) Delete(c Component) {
	delete(e.comp, c.id())
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
