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

	outerIter, innerIter int
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
