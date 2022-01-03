package ecs

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
