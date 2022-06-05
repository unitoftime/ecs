package ecs

// func SpecialMap2[A, B any, F func(Id, *A, *B)](world *World, lambda F) {
// 	view := ViewAll2[A, B](world)
// 	for view.Ok() {
// 		id, pos, vel := view.IterChunkClean()

// 		// mapFuncPhyGen(id, pos, vel, physicsTick)

// 		genMap2(id, pos, vel, lambda)
// 		// for j := range id {
// 		// 	lambda(id[j], &pos[j], &vel[j])
// 		// }
// 	}
// }

// func SpecialMap2NonGen(world *World, lambda func(Id, *Position, *Velocity)) {
// 	view := ViewAll2[Position, Velocity](world)
// 	for view.Ok() {
// 		id, pos, vel := view.IterChunkClean()

// 		// mapFuncPhyGen(id, pos, vel, physicsTick)

// 		for j := range id {
// 			lambda(id[j], &pos[j], &vel[j])
// 		}
// 	}
// }


// func Map[A any](world *World, lambda func(id Id, a *A)) {
func Map[A any, F func(Id, *A)](world *World, lambda F) {
	var a A
	archIds := world.engine.Filter(a)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i, id := range lookup.id {
			if id == InvalidEntity { continue }
			lambda(lookup.id[i], &aSlice.comp[i])
		}
	}
}

// func Map2[A, B any](world *World, lambda func(id Id, a *A, b *B)) {
func Map2[A, B any, F func(Id,*A,*B)](world *World, lambda F) {
	// This one is faster 360 ms
	ExecuteSystem2(world, func(query *Query2[A, B, ]) {
		query.Map(func(ids []Id, a []A, b []B) {
			if len(ids) != len(a) || len(ids) != len(b) { panic("ERR") }
			for i := range ids {
				lambda(ids[i], &a[i], &b[i])
			}
		})
	})

	// This one is slower 335 ms
	// var a A
	// var b B

	// archIds := world.engine.Filter(a, b)

	// // storages := getAllStorages(world, a)
	// aStorage := GetStorage[A](world.engine)
	// bStorage := GetStorage[B](world.engine)

	// for _, archId := range archIds {
	// 	aSlice, ok := aStorage.slice[archId]
	// 	if !ok { continue }
	// 	bSlice, ok := bStorage.slice[archId]
	// 	if !ok { continue }

	// 	lookup, ok := world.engine.lookup[archId]
	// 	if !ok { panic("LookupList is missing!") }

	// 	ids := lookup.id
	// 	aComp := aSlice.comp
	// 	bComp := bSlice.comp
	// 	if len(ids) != len(aComp) || len(ids) != len(bComp) {
	// 		panic("ERROR - Bounds don't match")
	// 	}
	// 	for i := range ids {
	// 		id := ids[i]
	// 		if id == InvalidEntity { continue }
	// 		aVal := &aComp[i]
	// 		bVal := &bComp[i]
	// 		lambda(id, aVal, bVal)
	// 		// if lookup.id[i] == InvalidEntity { continue }
	// 		// lambda(lookup.id[i], &aSlice.comp[i], &bSlice.comp[i])
	// 	}
	// }
}

// This ia a Map2, but if the lambda returns false, we stop looping
func SmartMap2[A, B any](world *World, lambda func(id Id, a *A, b *B) bool) {
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

		for i, id := range lookup.id {
			if id == InvalidEntity { continue }

			success := lambda(id, &aSlice.comp[i], &bSlice.comp[i])
			if !success { return }
		}
	}
}

func Map3[A, B, C any](world *World, lambda func(id Id, a *A, b *B, c *C)) {
	var a A
	var b B
	var c C
	archIds := world.engine.Filter(a, b, c)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)
	cStorage := GetStorage[C](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }
		cSlice, ok := cStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i, id := range lookup.id {
			if id == InvalidEntity { continue }
			lambda(lookup.id[i], &aSlice.comp[i], &bSlice.comp[i], &cSlice.comp[i])
		}
	}
}

func Map4[A, B, C, D any](world *World, lambda func(id Id, a *A, b *B, c *C, d *D)) {
	var a A
	var b B
	var c C
	var d D
	archIds := world.engine.Filter(a, b, c, d)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)
	cStorage := GetStorage[C](world.engine)
	dStorage := GetStorage[D](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }
		cSlice, ok := cStorage.slice[archId]
		if !ok { continue }
		dSlice, ok := dStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i, id := range lookup.id {
			if id == InvalidEntity { continue }
			lambda(lookup.id[i], &aSlice.comp[i], &bSlice.comp[i], &cSlice.comp[i], &dSlice.comp[i])
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// type SliceReader[T any] interface {
// 	Read(int) T
// }

// type CompSliceStorageReader[T any] interface {
// 	GetSliceReader(ArchId) SliceReader[T]
// }

// func GetStorage2[T any](e *ArchEngine) ComponentSliceStorage[T] {
// 	var val T
// 	n := name(val)
// 	ss, ok := e.compSliceStorage[n]
// 	if !ok {
// 		panic("Arch engine doesn't have this storage (I should probably just instantiate it and replace this code with write")
// 	}
// 	return storage
// }

// func (ss ComponentStorageSlice[T]) GetSliceReader(archId ArchId) (SliceReader, bool) {
// 	return ss.slice[archId]
// }
/*
type ptr[T any] interface {
	*T
}

type get[T, U any] interface {
	get([]T, int) U
}

type RO[T any] struct {
}
func (r RO[T]) get(slice []T, index int) T {
	return slice[index]
}

type RW[T any] struct {
}
func (r RW[T]) get(slice []T, index int) *T {
	return &slice[index]
}

func RwMap2[GA get[A, AO], GB get[B, BO], A any, B any, AO, BO any](world *World, lambda func(id Id, a AO, b BO)) {
	var a A
	var b B
	archIds := world.engine.Filter(a, b)

	var getA GA
	var getB GB

	// aPtr := (reflect.ValueOf(a).Kind() == reflect.Ptr)
	// bPtr := (reflect.ValueOf(b).Kind() == reflect.Ptr)

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
			lambda(lookup.id[i], getA.get(aSlice.comp, i), getB.get(bSlice.comp, i))
		}
	}
}

func RwMap[GA get[A, AO], A any, AO any](world *World, lambda func(id Id, a AO)) {
	var a A
	archIds := world.engine.Filter(a)

	var getA GA

	aStorage := GetStorage[A](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		for i := range lookup.id {
			lambda(lookup.id[i], getA.get(aSlice.comp, i))
		}
	}
}
*/
// func getInternalSlice[A any](world *World, archId ArchId) []A {
// 	aStorage := GetStorage[A](world.engine)
// 	aSlice, ok := aStorage.slice[archId]
// 	if !ok { return nil }

// 	return aSlice.comp
// }

type View2[A, B any, F func(Id, *A, *B)] struct {
	world *World
	id [][]Id
	aSlice [][]A
	bSlice [][]B

	outerIter, innerIter int
}
func ViewAll2[A, B any, F func(Id, *A, *B)](world *World) *View2[A, B, F] {
	v := View2[A, B, F]{
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

func (v *View2[A, B, F]) Reset() {
	v.outerIter = 0
	v.innerIter = 0
}

func (v *View2[A, B, F]) Ok() bool {
	return v.outerIter < len(v.id)
}

func (v *View2[A, B, F]) Map(lambda F) {
	for i := range v.id {
		// id := v.id[i]
		// aSlice := v.aSlice[i]
		// bSlice := v.bSlice[i]
		// for j := range id {
		// 	lambda(id[j], &aSlice[j], &bSlice[j])
		// }

		genMap2(v.id[i], v.aSlice[i], v.bSlice[i], lambda)
		// for j := range v.id[i] {
		// 	lambda(v.id[i][j], &v.aSlice[i][j], &v.bSlice[i][j])
		// }
	}
}

func genMap2[A any, B any, F func(Id, *A, *B)](id []Id, aa []A, bb []B, f F) {
	for j := range id {
		f(id[j], &aa[j], &bb[j])
	}
}

// func (v *View2[A, B, F]) Iter() (Id, A, B, bool) {
// 	v.innerIter++
// 	if v.innerIter >= len(v.id[v.outerIter]) {
// 		v.innerIter = 0
// 		v.outerIter++
// 	}

// 	if v.outerIter >= len(v.id) {
// 		var id Id
// 		var a A
// 		var b B
// 		return id, a, b, false
// 	}

// 	return v.id[v.outerIter][v.innerIter], v.aSlice[v.outerIter][v.innerIter], v.bSlice[v.outerIter][v.innerIter], true
// }

// func (v *View2[A, B, F]) Iter2(id *Id, a *A, b *B) bool {
// 	inner := v.innerIter
// 	outer := v.outerIter

// 	v.innerIter++
// 	if v.innerIter >= len(v.id[v.outerIter]) {
// 		v.innerIter = 0
// 		v.outerIter++
// 	}

// 	if outer >= len(v.id) {
// 		return false
// 	}

// 	*id = v.id[outer][inner]
// 	*a = v.aSlice[outer][inner]
// 	*b = v.bSlice[outer][inner]
// 	return true
// }

// func (v *View2[A, B, F]) IterPointer(id **Id, a **A, b **B) bool {
// 	inner := v.innerIter
// 	outer := v.outerIter

// 	v.innerIter++
// 	if v.innerIter >= len(v.id[v.outerIter]) {
// 		v.innerIter = 0
// 		v.outerIter++
// 	}

// 	if outer >= len(v.id) {
// 		return false
// 	}

// 	*id = &v.id[outer][inner]
// 	*a = &v.aSlice[outer][inner]
// 	*b = &v.bSlice[outer][inner]
// 	return true
// }

// func (v *View2[A, B, F]) Iter3() (Id, *A, *B, bool) {
// 	v.innerIter++
// 	if v.innerIter >= len(v.id[v.outerIter]) {
// 		v.innerIter = 0
// 		v.outerIter++
// 	}

// 	if v.outerIter >= len(v.id) {
// 		return InvalidEntity, nil, nil, false
// 	}

// 	return v.id[v.outerIter][v.innerIter], &v.aSlice[v.outerIter][v.innerIter], &v.bSlice[v.outerIter][v.innerIter], true
// }

// func (v *View2[A, B, F]) Next() {
// 	return v.Iter4()
// }

// func (v *View2[A, B, F]) Iter4() (Id, *A, *B) {
// 	inner := v.innerIter
// 	outer := v.outerIter

// 	v.innerIter++
// 	if v.innerIter >= len(v.id[v.outerIter]) {
// 		v.innerIter = 0
// 		v.outerIter++
// 	}

// 	if outer >= len(v.id) {
// 		return InvalidEntity, nil, nil
// 	}

// 	return v.id[outer][inner], &v.aSlice[outer][inner], &v.bSlice[outer][inner]
// }

// func (v *View2[A, B, F]) Iterate() (*Id, **A, **B, *Iterator2[A, B, F]) {
func (v *View2[A, B, F]) Iterate() *Iterator2[A, B, F] {
	newView := *v
	iterator := &Iterator2[A, B, F]{
		view: &newView,
	}
	return iterator
	// return iterator.id, &iterator.a, &iterator.b, iterator
}

// TODO  You could probably make iterators fast if you removed all the bounds checking that happens, but you'd probably have to do pointer arithmetic on the slices (potentially unsafe)
type Iterator2[A, B any, F func(Id, *A, *B)] struct {
	view *View2[A, B, F]
	innerIter, outerIter int
}

func (i *Iterator2[A, B, F]) Ok() bool {
	return i.outerIter < len(i.view.id)
}

func (i *Iterator2[A, B, F]) Next() (Id, *A, *B) {
	inner := i.innerIter
	outer := i.outerIter

	i.innerIter++
	if i.innerIter >= len(i.view.id[i.outerIter]) {
		i.innerIter = 0
		i.outerIter++
	}

	if outer >= len(i.view.id) {
		return InvalidEntity, nil, nil
	}

	return i.view.id[outer][inner], &i.view.aSlice[outer][inner], &i.view.bSlice[outer][inner]

	// return i.view.Iter4()
	// i.view.IterPointer(&i.id, &i.a, &i.b)
}

// Iterates on archetype chunks, returns underlying arrays so modifications are automatically written back
// func (v *View2[A, B, F]) IterChunk() ([]Id, []A, []B, bool) {
// 	if v.outerIter >= len(v.id) {
// 		return nil, nil, nil, false
// 	}
// 	idx := v.outerIter
// 	v.outerIter++

// 	return v.id[idx], v.aSlice[idx], v.bSlice[idx], true
// }

func (v *View2[A, B, F]) IterChunkClean() ([]Id, []A, []B) {
	if v.outerIter >= len(v.id) {
		return nil, nil, nil
	}
	idx := v.outerIter
	v.outerIter++

	return v.id[idx], v.aSlice[idx], v.bSlice[idx]
}

func (v *View2[A, B, F]) GetAllSlices() ([][]Id, [][]A, [][]B) {
	return v.id, v.aSlice, v.bSlice
}

func ArchetypeMap[A any, B any, F func(Id, *A, *B)](id [][]Id, aa [][]A, bb [][]B, f F) {
	for i := range id {
		for j := range id[i] {
			f(id[i][j], &aa[i][j], &bb[i][j])
		}
	}
}

func SliceMap2[A any, B any, F func(Id, *A, *B)](id []Id, aa []A, bb []B, f F) {
	for i := range id {
		f(id[i], &aa[i], &bb[i])
	}
}

// func CleanMap2[A, B any, F func(Id, *A, *B)](world *World, lambda F) {
// 	view := ViewAll2[A, B](world)
// 	for view.Ok() {
// 		id, pos, vel := view.IterChunkClean()
// 		for i := range id {
// 			lambda(id[i], &pos[i], &vel[i])
// 		}
// 	}
// }


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

type View2F[A, B any, F func(Id, *A, *B)] struct {
	world *World
	lambda F

	id [][]Id
	aSlice [][]A
	bSlice [][]B

	outerIter, innerIter int
}
func ViewAll2F[A, B any, F func(Id, *A, *B)](world *World, lambda F) *View2F[A, B, F] {
	v := View2F[A, B, F]{
		world: world,
		lambda: lambda,

		id: make([][]Id, 0),
		aSlice: make([][]A, 0),
		bSlice: make([][]B, 0),
		innerIter: -1, // When we iterate the first thing we do is increment
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

func (v *View2F[A, B, F]) Map() {
	for i := range v.id {
		genMap2(v.id[i], v.aSlice[i], v.bSlice[i], v.lambda)
	}
}
