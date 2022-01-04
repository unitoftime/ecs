package ecs

func Map[A any](world *World, lambda func(id Id, a A)) {
	var a A
	archIds := world.engine.Filter(a)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i := range lookup.id {
			lambda(lookup.id[i], aSlice.comp[i])
		}
	}
}

func Map2[A, B any](world *World, lambda func(id Id, a A, b B)) {
	var a A
	var b B
	archIds := world.engine.Filter(a, b)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i := range lookup.id {
			lambda(lookup.id[i], aSlice.comp[i], bSlice.comp[i])
		}
	}
}

func getInternalSlice[A any](world *World, archId ArchId) []A {
	aStorage := GetStorage[A](world.engine)
	aSlice, ok := aStorage.slice[archId]
	if !ok { return nil }

	return aSlice.comp
}

type View2[A, B any] struct {
	world *World
	id [][]Id
	aSlice [][]A
	bSlice [][]B

	outerIter, innerIter int
}
func ViewAll2[A, B any](world *World) *View2[A, B] {
	v := View2[A, B]{
		world: world,
		id: make([][]Id, 0),
		aSlice: make([][]A, 0),
		bSlice: make([][]B, 0),
	}
	var a A
	var b B
	archIds := v.world.engine.Filter(a, b)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](v.world.engine)
	bStorage := GetStorage[B](v.world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		v.id = append(v.id, lookup.id)
		v.aSlice = append(v.aSlice, aSlice.comp)
		v.bSlice = append(v.bSlice, bSlice.comp)
	}
	return &v
}

// func mapFunc2[A any, B any](id []Id, aa []A, bb []B, f func(id Id, a A, b B)){
// 	for j := range aa {
// 		f(id[j], aa[j], bb[j])
// 	}
// }

func (v *View2[A, B]) Map(lambda func(id Id, a A, b B)) {
	for i := range v.id {
		id := v.id[i]
		aSlice := v.aSlice[i]
		bSlice := v.bSlice[i]
		for j := range id {
			lambda(id[j], aSlice[j], bSlice[j])
		}
	}
}

func (v *View2[A, B]) Iter() (Id, A, B, bool) {
	v.innerIter++
	if v.innerIter >= len(v.id[v.outerIter]) {
		v.innerIter = 0
		v.outerIter++
	}

	if v.outerIter >= len(v.id) {
		var id Id
		var a A
		var b B
		return id, a, b, false
	}

	return v.id[v.outerIter][v.innerIter], v.aSlice[v.outerIter][v.innerIter], v.bSlice[v.outerIter][v.innerIter], true
}

func (v *View2[A, B]) Iter2(id *Id, a *A, b *B) bool {
	v.innerIter++
	if v.innerIter >= len(v.id[v.outerIter]) {
		v.innerIter = 0
		v.outerIter++
	}

	if v.outerIter >= len(v.id) {
		return false
	}

	*id = v.id[v.outerIter][v.innerIter]
	*a = v.aSlice[v.outerIter][v.innerIter]
	*b = v.bSlice[v.outerIter][v.innerIter]
	return true
}

// Iterates on archetype chunks, returns underlying arrays so modifications are automatically written back
func (v *View2[A, B]) IterChunk() ([]Id, []A, []B, bool) {
	if v.outerIter >= len(v.id) {
		return nil, nil, nil, false
	}
	idx := v.outerIter
	v.outerIter++

	return v.id[idx], v.aSlice[idx], v.bSlice[idx], true
}


// type View struct {
// 	world *World // TODO - Can I get away with just engine?
// 	id [][]Id
// 	componentSlices [][]any // component index -> outerIter -> innerIter

// 	outerIter, innerIter int
// }

// func ViewAll(world *World, comp ...Component) *View {
// 	v := View2[A, B]{
// 		world: world,
// 		id: make([][]Id, 0),
// 		componentSlices: make([][]any),
// 	}
// 	var a A
// 	var b B
// 	archIds := v.world.engine.Filter(a, b)

// 	// storages := getAllStorages(world, a)
// 	aStorage := GetStorage[A](v.world.engine)
// 	bStorage := GetStorage[B](v.world.engine)

// 	for _, archId := range archIds {
// 		aSlice, ok := aStorage.slice[archId]
// 		if !ok { continue }
// 		bSlice, ok := bStorage.slice[archId]
// 		if !ok { continue }

// 		lookup, ok := v.world.engine.lookup[archId]
// 		if !ok { panic("LookupList is missing!") }

// 		v.id = append(v.id, lookup.id)
// 		v.aSlice = append(v.aSlice, aSlice.comp)
// 		v.bSlice = append(v.bSlice, bSlice.comp)
// 	}
// 	return &v
// }



// import (
// 	// "fmt"
// 	// "reflect"
// )

// type View struct {
// 	world *World
// 	components []any
// 	// readonly []bool
// }

// func ViewAll(world *World, comp ...any) *View {
// 	return &View{
// 		world: world,
// 		components: comp,
// 	}
// }

// func getAllStorages(world *World, comp ...any) []Storage {
// 	storages := make([]Storage, 0)
// 	for i := range comp {
// 		s := world.archEngine.GetStorage(comp[i])
// 		storages = append(storages, s)
// 	}
// 	return storages
// }

// func Map[A any](world *World, lambda func(id Id, a A)) {
// 	var a A
// 	archIds := world.archEngine.Filter(a)
// 	storages := getAllStorages(world, a)

// 	for _, archId := range archIds {
// 		aList := GetStorageList[A](storages[0], archId)

// 		lookup, ok := world.archEngine.lookup[archId]
// 		if !ok { panic("LookupList is missing!") }
// 		for i := range lookup.Ids {
// 			lambda(lookup.Ids[i], aList[i])
// 		}
// 	}
// }

// func Map2[A, B any](world *World, lambda func(id Id, a *A, b *B)) {
// 	var a A
// 	var b B
// 	archIds := world.archEngine.Filter(a, b)
// 	storages := getAllStorages(world, a, b)

// 	// aPtr := (reflect.ValueOf(a).Kind() == reflect.Ptr)
// 	// bPtr := (reflect.ValueOf(b).Kind() == reflect.Ptr)

// 	for _, archId := range archIds {
// 		aList := GetStorageList[A](storages[0], archId)
// 		bList := GetStorageList[B](storages[1], archId)

// 		// var aIter Iterator[A] = ValueIterator[A]{aList}
// 		// var bIter Iterator[B] = ValueIterator[B]{bList}


// 		lookup, ok := world.archEngine.lookup[archId]
// 		if !ok { panic("LookupList is missing!") }
// 		for i := range lookup.Ids {
// 			lambda(lookup.Ids[i], &aList[i], &bList[i])

// 			// lambda(lookup.Ids[i], aIter.Get(i), bIter.Get(i))

// 			// if !aPtr && !bPtr {
// 			// 	lambda(lookup.Ids[i], aList[i], bList[i])
// 			// } else if aPtr && !bPtr {
// 			// 	lambda(lookup.Ids[i], &aList[i], bList[i])
// 			// } else if !aPtr && !bPtr {
// 			// 	lambda(lookup.Ids[i], aList[i], &bList[i])
// 			// } else if aPtr && bPtr {
// 			// 	lambda(lookup.Ids[i], &aList[i], &bList[i])
// 			// }
// 		}
// 	}
// }
