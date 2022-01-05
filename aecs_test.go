package ecs

import (
	"fmt"
	"testing"
)

type Pos struct {
	X,Y,Z float64
}

type Vel struct {
	X,Y,Z float64
}

func TestArchEngine(t *testing.T) {
	engine := NewArchEngine()
	WriteArch(engine, ArchId(1), Id(1), Pos{1,1,1})
	pos, ok := ReadArch[Pos](engine, ArchId(1), Id(1))
	fmt.Println(pos, ok)
	fmt.Println(engine)
}

func TestWorld(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	Write(world, id, C(Pos{1,1,1}), C(Vel{2,2,2}))
	fmt.Println(Read[Pos](world, id))
	fmt.Println(Read[Vel](world, id))

	world.Print()

	// id2 := world.NewId()
	// Write(world, id2, C(Pos{3,3,3}))
	// fmt.Println(Read[Pos](world, id2))
	// fmt.Println(Read[Vel](world, id2))

	// Write(world, id2, C(Pos{4, 4, 4}))
	// fmt.Println(Read[Pos](world, id2))
	// fmt.Println(Read[Vel](world, id2))

	// Write(world, id2, C(Vel{5, 5, 5}))
	// fmt.Println(Read[Pos](world, id2))
	// fmt.Println(Read[Vel](world, id2))

	// view := ViewAll(world, C(Pos{}), C(Vel{}))
	// view.Map(func(id Id, pos any) {
	// 	pos := a.(Position)
	// })

	// view := ViewAll[Pos](world)
	// view.Map(func(id Id, pos Pos) {
		
	// })

	Map[Pos](world, func(id Id, pos *Pos) {
		fmt.Println("Map:", id, pos)
	})

	// Map2[d1, d2](world, func(id Id, a *d1, b *d2) {
	// 	t.Log("Map2:", id, a, b)
	// })

	// Register[d1](world)
	// Register[d2](world)

	// a, ok := Read[d1](world, id)
	// t.Log(a, ok)

	// Write(world, id, d1{111})
	// {
	// 	a, ok = Read[d1](world, id)
	// 	t.Log(a, ok)
	// 	ent := ReadEntity(world, id)
	// 	t.Log(ent)
	// }

	// Write(world, id, d1{222})
	// {
	// 	a, ok = Read[d1](world, id)
	// 	t.Log(a, ok)
	// 	ent := ReadEntity(world, id)
	// 	t.Log(ent)
	// }

	// Write(world, id, d2{333})
	// {
	// 	b, ok := Read[d2](world, id)
	// 	t.Log(b, ok)
	// 	ent := ReadEntity(world, id)
	// 	t.Log(ent)
	// }

	// Map[d1](world, func(id Id, a d1) {
	// 	t.Log("Map:", id, a)
	// })

	// Map2[d1, d2](world, func(id Id, a *d1, b *d2) {
	// 	t.Log("Map2:", id, a, b)
	// })

	// archEngine := NewArchEngine()
	// aId := archEngine.NewArchId()

	// ArchWrite(archEngine, aId, d1{19})

	// b, ok := ArchRead[d1](archEngine, aId)
	// t.Log(b, ok)

	// archStorage := NewArchStorage[[]d1]()
	// a := []d1{d1{1}, d1{2}}
	// archId := ArchId(0)
	// ArchWrite[[]d1](archStorage, archId, a)

	// b, ok := ArchRead[[]d1](archStorage, archId)
	// t.Log(b, ok)

	// world := NewWorld()
	// id := world.NewId()
	// Write[d1](world, id, d1{100})

	// ret, ok := Read[d1](world, id)
	// t.Log(ok, ret)
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

// type d1List []d1
// func (t *d1List) ComponentSet(val interface{}) { *t = *val.(*d1List) }
// // func (t *d1List) ComponentGet(val interface{}) { val.(d1List) = *t  }
// func (t *d1List) InternalRead(index int, val interface{}) { *val.(*d1) = (*t)[index] }
// func (t *d1List) InternalWrite(index int, val interface{}) { (*t)[index] = val.(d1) }
// func (t *d1List) InternalAppend(val interface{}) { (*t) = append((*t), val.(d1)) }
// func (t *d1List) InternalPointer(index int) interface{} { return &(*t)[index]  }
// func (t *d1List) InternalReadVal(index int) interface{} { return (*t)[index] }
// func (t *d1List) Delete(index int) {
// 	lastVal := (*t)[len(*t)-1]
// 	(*t)[index] = lastVal
// 	(*t) = (*t)[:len(*t)-1]
// }
// func (t *d1List) Len() int { return len(*t) }

// type d2List []d2
// func (t *d2List) ComponentSet(val interface{}) { *t = *val.(*d2List) }
// func (t *d2List) InternalRead(index int, val interface{}) { *val.(*d2) = (*t)[index] }
// func (t *d2List) InternalWrite(index int, val interface{}) { (*t)[index] = val.(d2) }
// func (t *d2List) InternalAppend(val interface{}) { (*t) = append((*t), val.(d2)) }
// func (t *d2List) InternalPointer(index int) interface{} { return &(*t)[index] }
// func (t *d2List) InternalReadVal(index int) interface{} { return (*t)[index] }
// func (t *d2List) Delete(index int) {
// 	lastVal := (*t)[len(*t)-1]
// 	(*t)[index] = lastVal
// 	(*t) = (*t)[:len(*t)-1]
// }
// func (t *d2List) Len() int { return len(*t) }

// type d3List []d3
// func (t *d3List) ComponentSet(val interface{}) { *t = *val.(*d3List) }
// func (t *d3List) InternalRead(index int, val interface{}) { *val.(*d3) = (*t)[index] }
// func (t *d3List) InternalWrite(index int, val interface{}) { (*t)[index] = val.(d3) }
// func (t *d3List) InternalAppend(val interface{}) { (*t) = append((*t), val.(d3)) }
// func (t *d3List) InternalPointer(index int) interface{} { return &(*t)[index] }
// func (t *d3List) InternalReadVal(index int) interface{} { return (*t)[index] }
// func (t *d3List) Delete(index int) {
// 	lastVal := (*t)[len(*t)-1]
// 	(*t)[index] = lastVal
// 	(*t) = (*t)[:len(*t)-1]
// }
// func (t *d3List) Len() int { return len(*t) }

// type ComponentRegistry2 struct {
// }
// func (r *ComponentRegistry2) GetArchStorageType(component interface{}) ArchComponent {
// 	switch component.(type) {
// 	case d1:
// 		return &cList[d1]{}
// 	case *d1:
// 		return &cList[d1]{}
// 	case d2:
// 		return &cList[d2]{}
// 	case *d2:
// 		return &cList[d2]{}
// 	case d3:
// 		return &cList[d3]{}
// 	case *d3:
// 		return &cList[d3]{}

// 	// case d1:
// 	// 	return &d1List{}
// 	// case *d1:
// 	// 	return &d1List{}
// 	// case d2:
// 	// 	return &d2List{}
// 	// case *d2:
// 	// 	return &d2List{}
// 	// case d3:
// 	// 	return &d3List{}
// 	// case *d3:
// 	// 	return &d3List{}
// 	default:
// 		panic(fmt.Sprintf("Unknown component type: %T", component))
// 	}
// }
// func (r *ComponentRegistry2) GetComponentMask(component interface{}) ArchMask {
// 	switch component.(type) {
// 	case d1:
// 		return ArchMask(1 << 0)
// 	case d2:
// 		return ArchMask(1 << 1)
// 	case d3:
// 		return ArchMask(1 << 2)
// 	default:
// 		panic(fmt.Sprintf("Unknown component type: %T", component))
// 	}
// 	return 0
// }

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
// 	id := arch.NewArchId()
// 	ArchWrite(arch, id, d1List(make([]d1, 100)))
// 	ArchWrite(arch, id, d2List(make([]d2, 200)))

// 	ArchEach(arch, cList[d1]{}, func(id ArchId, a interface{}) {
// 		val := a.(d1)
// 		t.Log(val)
// 		// list := a.(d1List)
// 	})
// }
