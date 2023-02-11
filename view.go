package ecs

// import "fmt"

// Ideas:
// view := View2[Position, ecs.Maybe[Velocity]](world)
// 1. Track pointer to slice for each combination of filter parameters
// 2. User builds views with whatever they want or dont want
// 3. todo - need some way to have optionals, or to get whatever fields they want. I guess maybe they'll call map or iterate or something and when they call that they can specify which storages they want to pull out?

// type ArchHandle struct {
// 	archId ArchId
// 	lookup *Lookup
// 	compStorage []Storage
// }

// type without struct {
// 	fields []any
// }
// func Without(fields ...any) without {
// 	return without{fields}
// }
	// extractedFilters := make([]any, 0)
	// for _, f := range filters {
	// 	switch t := f.(type) {
	// 	case without:
	// 	}
	// }

// type buildQuery interface {
// 	build(world)
// }

// TODO - Filter types:
// Optional - Lets you view even if component is missing (func will return nil)
// With - Lets you add additional components that must be present
// Without - Lets you add additional components that must not be present

type filterList struct {
	filters []any
	cachedArchetypeGeneration int // Denotes the world's archetype generation that was used to create the list of archIds. If the world has a new generation, we should probably regenerate
	archIds []ArchId
}
func newFilterList(filters []any) filterList {
	return filterList{
		filters: filters,
		archIds: make([]ArchId, 0),
	}
}
func (f *filterList) regenerate(world *World) {
	if world.engine.generation() != f.cachedArchetypeGeneration {
		f.archIds = world.engine.FilterList(f.archIds, f.filters)
		f.cachedArchetypeGeneration = world.engine.generation()
	}
}

type View1[A any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
}

func Query1[A any](world *World, filters ...any) *View1[A] {
	storageA := getStorage[A](world.engine)

	var a A
	filters = append(filters, a)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View1[A]{
		world: world,
		filter: filter,
		storageA: storageA,
	}
	return v
}

func (v *View1[A]) Read(id Id) (*A) {
	if id == InvalidEntity { return nil }

	archId, ok := v.world.arch[id]
	if !ok { return nil }

	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil }

	aSlice, ok := v.storageA.slice[archId]
	if !ok { return nil }

	return &aSlice.comp[index]
}

func (v *View1[A]) MapId(lambda func(id Id, a *A)) {
	v.filter.regenerate(v.world)
	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		if len(ids) != len(aComp) {
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i])
		}
	}
}

// --------------------------------------------------------------------------------
// - View 2
// --------------------------------------------------------------------------------

type View2[A,B any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
}

func Query2[A,B any](world *World, filters ...any) *View2[A,B] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)

	var a A
	var b B
	filters = append(filters, a, b)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View2[A,B]{
		world: world,
		filter: filter,
		storageA: storageA,
		storageB: storageB,
	}
	return v
}

// Reads always try to read as many components as possible regardless of if the component exists or not
func (v *View2[A,B]) Read(id Id) (*A, *B) {
	if id == InvalidEntity { return nil, nil }

	archId, ok := v.world.arch[id]
	if !ok {
		return nil, nil
	}
	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil, nil }

	var retA *A
	aSlice, ok := v.storageA.slice[archId]
	if ok {
		retA = &aSlice.comp[index]
	}

	var retB *B
	bSlice, ok := v.storageB.slice[archId]
	if ok {
		retB = &bSlice.comp[index]
	}

	return retA, retB
	// return &aSlice.comp[index], &bSlice.comp[index]
}

func (v *View2[A,B]) MapId(lambda func(id Id, a *A, b *B)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) {
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i])
		}
	}
}

func (v *View2[A,B]) MapSlices(lambda func(id []Id, a []A, b []B)) {
	v.filter.regenerate(v.world)

	id := make([][]Id, 0)
	sliceListA := make([][]A, 0)
	sliceListB := make([][]B, 0)

	for _, archId := range v.filter.archIds {
		sliceA, ok := v.storageA.slice[archId]
		if !ok { continue }
		sliceB, ok := v.storageB.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		id = append(id, lookup.id)
		sliceListA = append(sliceListA, sliceA.comp)
		sliceListB = append(sliceListB, sliceB.comp)
	}

	for i := range id {
		lambda(id[i], sliceListA[i], sliceListB[i])
	}
}
// --------------------------------------------------------------------------------
// - View 3
// --------------------------------------------------------------------------------

type View3[A,B,C any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
}

func Query3[A,B,C any](world *World, filters ...any) *View3[A,B,C] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)

	var a A
	var b B
	var c C
	filters = append(filters, a, b, c)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View3[A,B,C]{
		world: world,
		filter: filter,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
	}
	return v
}

func (v *View3[A,B,C]) MapId(lambda func(id Id, a *A, b *B, c *C)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		if !ok { continue }
		cSlice, ok := v.storageC.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		cComp := cSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp) {
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i], &cComp[i])
		}
	}
}

// Reads always try to read as many components as possible regardless of if the component exists or not
func (v *View3[A,B,C]) Read(id Id) (*A, *B, *C) {
	if id == InvalidEntity { return nil, nil, nil }

	archId, ok := v.world.arch[id]
	if !ok {
		return nil, nil, nil
	}
	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil, nil, nil }

	var retA *A
	sliceA, ok := v.storageA.slice[archId]
	if ok {
		retA = &sliceA.comp[index]
	}
	var retB *B
	sliceB, ok := v.storageB.slice[archId]
	if ok {
		retB = &sliceB.comp[index]
	}
	var retC *C
	sliceC, ok := v.storageC.slice[archId]
	if ok {
		retC = &sliceC.comp[index]
	}

	return retA, retB, retC
}

func (v *View3[A,B,C]) MapSlices(lambda func(id []Id, a []A, b []B, c []C)) {
	v.filter.regenerate(v.world)

	id := make([][]Id, 0)
	sliceListA := make([][]A, 0)
	sliceListB := make([][]B, 0)
	sliceListC := make([][]C, 0)

	for _, archId := range v.filter.archIds {
		sliceA, ok := v.storageA.slice[archId]
		if !ok { continue }
		sliceB, ok := v.storageB.slice[archId]
		if !ok { continue }
		sliceC, ok := v.storageC.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		id = append(id, lookup.id)
		sliceListA = append(sliceListA, sliceA.comp)
		sliceListB = append(sliceListB, sliceB.comp)
		sliceListC = append(sliceListC, sliceC.comp)
	}

	for i := range id {
		lambda(id[i], sliceListA[i], sliceListB[i], sliceListC[i])
	}
}
// --------------------------------------------------------------------------------
// - View 4
// --------------------------------------------------------------------------------

type View4[A,B,C,D any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
	storageD componentSliceStorage[D]
}

func Query4[A,B,C,D any](world *World, filters ...any) *View4[A,B,C,D] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)
	storageD := getStorage[D](world.engine)

	var a A
	var b B
	var c C
	var d D
	filters = append(filters, a, b, c, d)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View4[A,B,C,D]{
		world: world,
		filter: filter,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
	}
	return v
}

func (v *View4[A,B,C,D]) MapId(lambda func(id Id, a *A, b *B, c *C, d *D)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		if !ok { continue }
		cSlice, ok := v.storageC.slice[archId]
		if !ok { continue }
		dSlice, ok := v.storageD.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		cComp := cSlice.comp
		dComp := dSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp) || len(ids) != len(dComp){
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i], &cComp[i], &dComp[i])
		}
	}
}

// Reads always try to read as many components as possible regardless of if the component exists or not
func (v *View4[A,B,C,D]) Read(id Id) (*A, *B, *C, *D) {
	if id == InvalidEntity { return nil, nil, nil, nil }

	archId, ok := v.world.arch[id]
	if !ok {
		return nil, nil, nil, nil
	}
	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil, nil, nil, nil }

	var retA *A
	sliceA, ok := v.storageA.slice[archId]
	if ok {
		retA = &sliceA.comp[index]
	}
	var retB *B
	sliceB, ok := v.storageB.slice[archId]
	if ok {
		retB = &sliceB.comp[index]
	}
	var retC *C
	sliceC, ok := v.storageC.slice[archId]
	if ok {
		retC = &sliceC.comp[index]
	}
	var retD *D
	sliceD, ok := v.storageD.slice[archId]
	if ok {
		retD = &sliceD.comp[index]
	}

	return retA, retB, retC, retD
}

// --------------------------------------------------------------------------------
// - View 5
// --------------------------------------------------------------------------------

type View5[A,B,C,D,E any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
	storageD componentSliceStorage[D]
	storageE componentSliceStorage[E]
}

func Query5[A,B,C,D,E any](world *World, filters ...any) *View5[A,B,C,D,E] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)
	storageD := getStorage[D](world.engine)
	storageE := getStorage[E](world.engine)

	var a A
	var b B
	var c C
	var d D
	var e E
	filters = append(filters, a, b, c, d, e)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View5[A,B,C,D,E]{
		world: world,
		filter: filter,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
		storageE: storageE,
	}
	return v
}

func (v *View5[A,B,C,D,E]) MapId(lambda func(id Id, a *A, b *B, c *C, d *D, e *E)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		if !ok { continue }
		cSlice, ok := v.storageC.slice[archId]
		if !ok { continue }
		dSlice, ok := v.storageD.slice[archId]
		if !ok { continue }
		eSlice, ok := v.storageE.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		cComp := cSlice.comp
		dComp := dSlice.comp
		eComp := eSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp) || len(ids) != len(dComp) || len(ids) != len(eComp) {
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i], &cComp[i], &dComp[i], &eComp[i])
		}
	}
}

// Reads always try to read as many components as possible regardless of if the component exists or not
func (v *View5[A,B,C,D,E]) Read(id Id) (*A, *B, *C, *D, *E) {
	if id == InvalidEntity { return nil, nil, nil, nil, nil }

	archId, ok := v.world.arch[id]
	if !ok {
		return nil, nil, nil, nil, nil
	}
	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil, nil, nil, nil, nil }

	var retA *A
	sliceA, ok := v.storageA.slice[archId]
	if ok {
		retA = &sliceA.comp[index]
	}
	var retB *B
	sliceB, ok := v.storageB.slice[archId]
	if ok {
		retB = &sliceB.comp[index]
	}
	var retC *C
	sliceC, ok := v.storageC.slice[archId]
	if ok {
		retC = &sliceC.comp[index]
	}
	var retD *D
	sliceD, ok := v.storageD.slice[archId]
	if ok {
		retD = &sliceD.comp[index]
	}
	var retE *E
	sliceE, ok := v.storageE.slice[archId]
	if ok {
		retE = &sliceE.comp[index]
	}

	return retA, retB, retC, retD, retE
}

// --------------------------------------------------------------------------------
// - View 6
// --------------------------------------------------------------------------------

type View6[A,B,C,D,E,F any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
	storageD componentSliceStorage[D]
	storageE componentSliceStorage[E]
	storageF componentSliceStorage[F]
}

func Query6[A,B,C,D,E,F any](world *World, filters ...any) *View6[A,B,C,D,E,F] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)
	storageD := getStorage[D](world.engine)
	storageE := getStorage[E](world.engine)
	storageF := getStorage[F](world.engine)

	var a A
	var b B
	var c C
	var d D
	var e E
	var f F
	filters = append(filters, a, b, c, d, e, f)
	filter := filterList{
		filters: filters,
	}

	filter.regenerate(world)

	v := &View6[A,B,C,D,E,F]{
		world: world,
		filter: filter,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
		storageE: storageE,
		storageF: storageF,
	}
	return v
}

func (v *View6[A,B,C,D,E,F]) MapId(lambda func(id Id, a *A, b *B, c *C, d *D, e *E, f *F)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		if !ok { continue }
		cSlice, ok := v.storageC.slice[archId]
		if !ok { continue }
		dSlice, ok := v.storageD.slice[archId]
		if !ok { continue }
		eSlice, ok := v.storageE.slice[archId]
		if !ok { continue }
		fSlice, ok := v.storageF.slice[archId]
		if !ok { continue }

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		cComp := cSlice.comp
		dComp := dSlice.comp
		eComp := eSlice.comp
		fComp := fSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp) || len(ids) != len(dComp) || len(ids) != len(eComp) || len(ids) != len(fComp) {
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i], &cComp[i], &dComp[i], &eComp[i], &fComp[i])
		}
	}
}

// Reads always try to read as many components as possible regardless of if the component exists or not
func (v *View6[A,B,C,D,E,F]) Read(id Id) (*A, *B, *C, *D, *E, *F) {
	if id == InvalidEntity { return nil, nil, nil, nil, nil, nil }

	archId, ok := v.world.arch[id]
	if !ok {
		return nil, nil, nil, nil, nil, nil
	}
	lookup, ok := v.world.engine.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.index[id]
	if !ok { return nil, nil, nil, nil, nil, nil }

	var retA *A
	sliceA, ok := v.storageA.slice[archId]
	if ok {
		retA = &sliceA.comp[index]
	}
	var retB *B
	sliceB, ok := v.storageB.slice[archId]
	if ok {
		retB = &sliceB.comp[index]
	}
	var retC *C
	sliceC, ok := v.storageC.slice[archId]
	if ok {
		retC = &sliceC.comp[index]
	}
	var retD *D
	sliceD, ok := v.storageD.slice[archId]
	if ok {
		retD = &sliceD.comp[index]
	}
	var retE *E
	sliceE, ok := v.storageE.slice[archId]
	if ok {
		retE = &sliceE.comp[index]
	}
	var retF *F
	sliceF, ok := v.storageF.slice[archId]
	if ok {
		retF = &sliceF.comp[index]
	}

	return retA, retB, retC, retD, retE, retF
}
