package aecs

import (
	"fmt"
	"testing"
)

type d1 struct {
	value int
}

type d2 struct {
	value int
}

type d1List []d1
func (t *d1List) ComponentSet(val interface{}) { *t = *val.(*d1List) }
// func (t *d1List) ComponentGet(val interface{}) { val.(d1List) = *t  }
func (t *d1List) InternalRead(index int, val interface{}) { *val.(*d1) = (*t)[index]  }
func (t *d1List) InternalWrite(index int, val interface{}) { (*t)[index] = *val.(*d1) }
func (t *d1List) InternalAppend(val interface{}) { (*t) = append((*t), val.(d1)) }
func (t *d1List) InternalPointer(index int) interface{} { return &(*t)[index]  }
func (t *d1List) Len() int { return len(*t) }

type d2List []d2
func (t *d2List) ComponentSet(val interface{}) { *t = *val.(*d2List) }
func (t *d2List) InternalRead(index int, val interface{}) { *val.(*d2) = (*t)[index]  }
func (t *d2List) InternalWrite(index int, val interface{}) { (*t)[index] = *val.(*d2) }
func (t *d2List) InternalAppend(val interface{}) { (*t) = append((*t), val.(d2)) }
func (t *d2List) InternalPointer(index int) interface{} { return &(*t)[index] }
func (t *d2List) Len() int { return len(*t) }

type ComponentRegistry struct {
}
func (r *ComponentRegistry) GetArchStorageType(component interface{}) ArchComponent {
	switch component.(type) {
	case d1:
		return &d1List{}
	case *d1:
		return &d1List{}
	case d2:
		return &d2List{}
	case *d2:
		return &d2List{}
	default:
		panic(fmt.Sprintf("Unknown component type: %T", component))
	}
}
func (r *ComponentRegistry) GetComponentMask(component interface{}) ArchMask {
	switch component.(type) {
	case d1:
		return ArchMask(1 << 0)
	case d2:
		return ArchMask(1 << 1)
	default:
		panic(fmt.Sprintf("Unknown component type: %T", component))
	}
	return 0
}

func TestWorld(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	a1 := d1{1}
	b1 := d2{2}
	Write(world, id, a1, b1)
	a := d1{}
	b := d2{}

	ok := Read(world, id, &a)
	if !ok { t.Errorf("Component should be there") }
	if a != a1 { t.Errorf("Failure - Expected: %v - got: %v", a1, a) }

	ok = Read(world, id, &b)
	if !ok { t.Errorf("Component should be there") }
	if b != b1 { t.Errorf("Failure - Expected: %v - got: %v", b1, b) }
	world.Print()

	for i := 0; i < 5; i ++ {
		view := ViewAll(world, &d1{5}, &d2{6})
		t.Log("view:\n", view)
		view.Map(func(id Id, comp ...interface{}) {
			aa := comp[0].(*d1)
			bb := comp[1].(*d2)
			t.Log("HERE:", id, aa, bb, "\n")
			aa.value += 1
			bb.value += 1
		})
	}
	// TODO - make storage injected by client and implement a specific interface? What interface?
}

func TestArchEngine(t *testing.T) {
	arch := NewArchEngine()
	id := arch.NewArchId()
	ArchWrite(arch, id, d1List(make([]d1, 0)))
	ArchWrite(arch, id, d2List(make([]d2, 0)))

	ArchEach(arch, d1List{}, func(id ArchId, a interface{}) {
		// list := a.(d1List)
	})
}
