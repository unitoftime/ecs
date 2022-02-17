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

func (e *Entity) Add(comp Component) {
	// n := name(comp) // TODO - name is wrong here because we pass in a boxed component
	n := comp.Name()
	e.comp[n] = comp
}

// TODO - Hacky and probs slow
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
	return icomp.(CompBox[T]).comp, true
}

func WriteEntity(world *World, id Id, ent *Entity) {
	comps := ent.Comps()
	Write(world, id, comps...)
}


// type Entity map[string]Component

// func NewEntityy(components ...Component) *Entity {
// 	e := Entity(make(map[string]Component))
// 	for i := range components {
// 		e[comp[i].Name()] = comp[i]
// 	}
// }

/*
// TODO - revisit this abstraction
type Copier interface {
	Copy() interface{}
}

// TODO - revisit this abstraction
type Entity map[string]interface{}

func BlankEntity() Entity {
	return Entity(make(map[string]interface{}))
}

func (e Entity) Copy() Entity {
	copy := BlankEntity()
	for k,v := range e {
		c, ok := v.(Copier)
		if ok {
			// fmt.Println("Copying:", k)
			// If the component has a custom copy interface, then copy it
			copy[k] = c.Copy()
		} else {
			// Else just copy the top level struct
			copy[k] = v
		}
	}
	return copy
}

func (e Entity) Write(c interface{}) {
	name := name(c)
	e[name] = c
}

// func ReadEntity[T any](e Entity) (T, bool) { }

func (e Entity) Read(c interface{}) interface{} {
	name := name(c)
	comp, ok := e[name]
	if !ok { return nil }
	return comp
}

// TODO - untested
func GetEntity(world *World, id Id) Entity {
	comps := ReadAll(world, id)
	ent := BlankEntity()
	for i := range comps {
		ent.Write(comps[i])
	}
	return ent
}

func WriteEntity(world *World, id Id, ent Entity) {
	comps := make([]interface{}, 0)
	entCopy := ent.Copy()
	for k := range entCopy {
		comps = append(comps, ent[k])
	}

	Write(world, id, comps...)
}
*/
