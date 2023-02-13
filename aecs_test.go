package ecs

import (
	"fmt"
	"testing"
)

type Pos struct {
	X, Y, Z float64
}

type Vel struct {
	X, Y, Z float64
}

func TestArchEngine(t *testing.T) {
	engine := newArchEngine()
	writeArch(engine, archetypeId(1), Id(1), Pos{1, 1, 1})
	pos, ok := readArch[Pos](engine, archetypeId(1), Id(1))
	fmt.Println(pos, ok)
	fmt.Println(engine)
}

type d1 struct {
	value int
}

type d2 struct {
	value int
}

type d3 struct {
	value int
}

// func TestBuildEntitiesGeneric(t *testing.T) {
// 	SetRegistry(&ComponentRegistry2{})
// 	world := NewWorld()
// }

// func TestBuildEntities(t *testing.T) {
// 	SetRegistry(&ComponentRegistry2{})
// 	world := NewWorld()

// 	n := 1_000_000
// 	mod := 3

// 	for i := 0; i < n; i++ {
// 		id := world.NewId()
// 		switch int(id) % mod {
// 		case 0:
// 			Write(world, id, d1{int(id)})
// 		case 1:
// 			Write(world, id, d2{int(id)})
// 		case 2:
// 			Write(world, id, d1{int(id)}, d2{int(id)})
// 		}
// 	}

// 	{
// 		view := ViewAll(world, &d1{}, &d2{})
// 		view.Map(func(id Id, comp ...interface{}) {
// 			if int(id) % mod == 0 || int(id) % mod == 1 {
// 				t.Errorf("Failure - These entities should match the view")
// 			}

// 			a := comp[0].(*d1)
// 			b := comp[1].(*d2)
// 			if a.value != int(id) || b.value != int(id) {
// 				t.Errorf("Failure - d1 and/or d2 are set wrong")
// 			}
// 		})
// 	}

// 	{
// 		view := ViewAll(world, &d1{})
// 		view.Map(func(id Id, comp ...interface{}) {
// 			if int(id) % mod == 1 {
// 				t.Errorf("Failure - These entities should match the view")
// 			}

// 			a := comp[0].(*d1)
// 			if a.value != int(id) {
// 				t.Errorf("Failure - d1 is set wrong")
// 			}
// 		})
// 	}

// 	{
// 		view := ViewAll(world, &d2{})
// 		view.Map(func(id Id, comp ...interface{}) {
// 			if int(id) % mod == 0 {
// 				t.Errorf("Failure - These entities should match the view")
// 			}

// 			a := comp[0].(*d2)
// 			if a.value != int(id) {
// 				t.Errorf("Failure - d2 is set wrong")
// 			}
// 		})
// 	}
// }

// func TestWorld(t *testing.T) {
// 	SetRegistry(&ComponentRegistry2{})
// 	world := NewWorld()
// 	id := world.NewId()

// 	a1 := d1{1}
// 	b1 := d2{2}
// 	Write(world, id, a1, b1)
// 	a := d1{}
// 	b := d2{}

// 	ok := Read(world, id, &a)
// 	if !ok { t.Errorf("Component should be there") }
// 	if a != a1 { t.Errorf("Failure - Expected: %v - got: %v", a1, a) }

// 	ok = Read(world, id, &b)
// 	if !ok { t.Errorf("Component should be there") }
// 	if b != b1 { t.Errorf("Failure - Expected: %v - got: %v", b1, b) }
// 	world.Print()

// 	for i := 0; i < 5; i ++ {
// 		view := ViewAll(world, &d1{5}, &d2{6})
// 		t.Log("view:\n", view)
// 		view.Map(func(id Id, comp ...interface{}) {
// 			aa := comp[0].(*d1)
// 			bb := comp[1].(*d2)
// 			t.Log("HERE:", id, aa, bb, "\n")
// 			aa.value += 1
// 			bb.value += 1
// 		})
// 	}

// 	comps := ReadAll(world, id)
// 	fmt.Println(comps)
// 	Delete(world, id)
// 	comps2 := ReadAll(world, id)
// 	fmt.Println(comps2)

// 	Write(world, id, d1{10})
// 	fmt.Println(ReadAll(world, id))
// 	Write(world, id, d2{11})
// 	fmt.Println(ReadAll(world, id))
// 	DeleteComponents(world, id, d1{})
// 	fmt.Println(ReadAll(world, id))
// 	Write(world, id, d1{15})
// 	fmt.Println(ReadAll(world, id))
// }

// func TestArchEngine(t *testing.T) {
// 	arch := NewArchEngine()
// 	id := arch.NewarchetypeId()
// 	ArchWrite(arch, id, d1List(make([]d1, 100)))
// 	ArchWrite(arch, id, d2List(make([]d2, 200)))

// 	ArchEach(arch, cList[d1]{}, func(id archetypeId, a interface{}) {
// 		val := a.(d1)
// 		t.Log(val)
// 		// list := a.(d1List)
// 	})
// }
