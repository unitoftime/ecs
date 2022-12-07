package ecs

func ExecuteSystem2[A, B any](world *World, f func(q *Query2[A, B])) {
	query := NewQuery2[A, B](world)

	f(query)
}

type Query2[A, B any] struct {
	world *World
	id [][]Id
	aSlice [][]A
	bSlice [][]B
}
func NewQuery2[A, B any](world *World) *Query2[A, B] {
	v := Query2[A, B]{
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


func (q *Query2[A, B]) Map(f func(ids []Id, a []A, b []B)) {
	// Panics are for bounds check eliminations
	ids := q.id
	aa := q.aSlice
	bb := q.bSlice
	if len(ids) != len(aa) || len(ids) != len(bb) { panic("ERR") }
	for i := range q.id {
		f(ids[i], aa[i], bb[i])
	}
}

func (q *Query2[A, B]) Map2D(f func([]Id, []A, []B, []Id, []A, []B)) {
	// Old code:
	// for i := range q.ids {
	// 	for j := range q.ids {
	// 		f(q.ids[i], q.a[i], q.b[i], q.ids[j], q.a[j], q.b[j])
	// 	}
	// }

	// Panics are for bounds check eliminations
	ids := q.id
	aa := q.aSlice
	bb := q.bSlice
	if len(ids) != len(aa) || len(ids) != len(bb) { panic("ERR") }
	for i := range ids {
		iii1 := ids[i]
		aaa1 := aa[i]
		bbb1 := bb[i]

		if len(iii1) != len(aaa1) || len(iii1) != len(aaa1) { panic("ERR") }
		for j := range ids {
			// f(q.ids[i], q.a[i], q.b[i], q.ids[j], q.a[j], q.b[j])

			iii2 := ids[j]
			aaa2 := aa[j]
			bbb2 := bb[j]
			if len(iii2) != len(aaa2) || len(iii2) != len(aaa2) { panic("ERR") }

			f(iii1, aaa1, bbb1, iii2, aaa2, bbb2)
		}
	}
}

func (q *Query2[A, B]) Iterate() *UnsafeIterator2[A, B] {
	iterator := &UnsafeIterator2[A, B]{
		query: q,
		outerLen: len(q.id),
		id: q.id[0],
		a: q.aSlice[0],
		b: q.bSlice[0],
	}
	return iterator
}

// --------------------------------------------------------------------------------

type Query3[A, B, C any] struct {
	world *World
	id [][]Id
	aSlice [][]A
	bSlice [][]B
	cSlice [][]C
}
func NewQuery3[A, B, C any](world *World) *Query3[A, B, C] {
	v := Query3[A, B, C]{
		world: world,
		id: make([][]Id, 0),
		aSlice: make([][]A, 0),
		bSlice: make([][]B, 0),
		cSlice: make([][]C, 0),
	}
	var a A
	var b B
	var c C
	archIds := v.world.engine.Filter(a, b, c)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](v.world.engine)
	bStorage := GetStorage[B](v.world.engine)
	cStorage := GetStorage[C](v.world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }
		cSlice, ok := cStorage.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		v.id = append(v.id, lookup.id)
		v.aSlice = append(v.aSlice, aSlice.comp)
		v.bSlice = append(v.bSlice, bSlice.comp)
		v.cSlice = append(v.cSlice, cSlice.comp)
	}
	return &v
}


func (q *Query3[A, B, C]) Map(f func(ids []Id, a []A, b []B, c []C)) {
	// Panics are for bounds check eliminations
	ids := q.id
	aa := q.aSlice
	bb := q.bSlice
	cc := q.cSlice
	if len(ids) != len(aa) || len(ids) != len(bb) || len(ids) != len(cc){ panic("ERR") }
	for i := range q.id {
		f(ids[i], aa[i], bb[i], cc[i])
	}
}

// --------------------------------------------------------------------------------


type UnsafeIterator2[A, B any] struct {
	query *Query2[A, B]
	outerLen int
	innerIter, outerIter int

	id []Id
	a []A
	b []B
}

func (i *UnsafeIterator2[A, B]) Ok() bool {
	return i.outerIter < len(i.query.id)
}


func (i *UnsafeIterator2[A, B]) Next() (Id, *A, *B) {
	// // Note this is broken
	// func (i *UnsafeIterator2[A, B]) outerTransition() (Id, *A, *B) {
	// 	i.innerIter = 0
	// 	i.outerIter++

	// 	if i.outerIter < i.outerLen {
	// 		i.id = i.query.id[i.outerIter]
	// 		i.a = i.query.aSlice[i.outerIter]
	// 		i.b = i.query.bSlice[i.outerIter]
	// 	}
	// }
	// if i.outerIter >= i.outerLen {
	// 	// Case where our outer iterator has finished (no more to iterate)
	// 	return InvalidEntity, nil, nil
	// } else if i.innerIter >= len(i.id) {
	// 	// Case where our inner iterator has finished (no more in this group, go to the next one)
	// 	return i.outerTransition()
	// }

	// inner := i.innerIter
	// i.innerIter++
	// return i.id[inner], &i.a[inner], &i.b[inner]


	if i.outerIter >= i.outerLen {
		return InvalidEntity, nil, nil
	}

	inner := i.innerIter

	i.innerIter++
	if i.innerIter >= len(i.id) {
		i.innerIter = 0
		i.outerIter++

		if i.outerIter < i.outerLen {
			i.id = i.query.id[i.outerIter]
			i.a = i.query.aSlice[i.outerIter]
			i.b = i.query.bSlice[i.outerIter]
		}
	}

	return i.id[inner], &i.a[inner], &i.b[inner]
}
