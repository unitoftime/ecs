package ecs

type Entity struct {
	comp map[string]Component
}

func NewEntity(components ...Component) *Entity {
	c := make(map[string]Component)
	for i := range components {
		c[components[i].Name()] = components[i]
	}
	return &Entity{
		comp: c,
	}
}

func (e *Entity) Add(components ...Component) {
	for i := range components {
		e.comp[components[i].Name()] = components[i]
	}
}

// TODO Hacky - Could probably improve performance
func (e *Entity) Comps() []Component {
	ret := make([]Component, 0, len(e.comp))
	for _, v := range e.comp {
		ret = append(ret, v)
	}
	return ret
}

func ReadFromEntity[T any](ent *Entity) (T, bool) {
	var t T
	n := name(t)

	icomp, ok := ent.comp[n]
	if !ok {
		return t, false
	}
	return icomp.(CompBox[T]).Comp, true
}

func WriteEntity(world *World, id Id, ent *Entity) {
	comps := ent.Comps()
	Write(world, id, comps...)
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

