package ecs

import (
	"slices"
)

// TODO - Filter types:
// Optional - Lets you view even if component is missing (func will return nil)
// With - Lets you add additional components that must be present
// Without - Lets you add additional components that must not be present
type Filter interface {
	Filter([]componentId) []componentId
}

type without struct {
	mask archetypeMask
}

// Creates a filter to ensure that entities will not have the specified components
func Without(comps ...any) without {
	return without{
		mask: buildArchMaskFromAny(comps...),
	}
}
func (w without) Filter(list []componentId) []componentId {
	return list // Dont filter anything. We need to exclude later on
	// return append(list, w.comps...)
}

type with struct {
	comps []componentId
}

// Creates a filter to ensure that entities have the specified components
func With(comps ...any) with {
	ids := make([]componentId, len(comps))
	for i := range comps {
		ids[i] = name(comps[i])
	}
	return with{
		comps: ids,
	}
}

func (w with) Filter(list []componentId) []componentId {
	return append(list, w.comps...)
}

type optional struct {
	comps []componentId
}

// Creates a filter to make the query still iterate even if a specific component is missing, in which case you'll get nil if the component isn't there when accessed
func Optional(comps ...any) optional {
	ids := make([]componentId, len(comps))
	for i := range comps {
		ids[i] = name(comps[i])
	}

	return optional{
		comps: ids,
	}
}

func (f optional) Filter(list []componentId) []componentId {
	for i := 0; i < len(list); i++ {
		for j := range f.comps {
			if list[i] == f.comps[j] {
				// If we have a match, we want to remove it from the list.
				list[i] = list[len(list)-1]
				list = list[:len(list)-1]

				// Because we just moved the last element to index i, we need to go back to process that element
				i--
				break
			}
		}
	}
	return list
}

type filterList struct {
	comps                     []componentId
	withoutArchMask archetypeMask
	cachedArchetypeGeneration int // Denotes the world's archetype generation that was used to create the list of archIds. If the world has a new generation, we should probably regenerate
	archIds                   []archetypeId
}

func newFilterList(comps []componentId, filters ...Filter) filterList {
	var withoutArchMask archetypeMask
	for _, f := range filters {
		withoutFilter, isWithout := f.(without)
		if isWithout {
			withoutArchMask = withoutFilter.mask
		} else {
			comps = f.Filter(comps)
		}
	}

	return filterList{
		comps:   comps,
		withoutArchMask: withoutArchMask,
		archIds: make([]archetypeId, 0),
	}
}
func (f *filterList) regenerate(world *World) {
	if world.engine.getGeneration() != f.cachedArchetypeGeneration {
		f.archIds = world.engine.FilterList(f.archIds, f.comps)

		if f.withoutArchMask != blankArchMask {
			f.archIds = slices.DeleteFunc(f.archIds, func(archId archetypeId) bool {
				return world.engine.dcr.archIdOverlapsMask(archId, f.withoutArchMask)
			})
		}

		f.cachedArchetypeGeneration = world.engine.getGeneration()
	}
}

/* Note: replaced all this with code generation

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
type View1[A any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
}

// Creates a View for the specified world with the specified component filters.
func Query1[A any](world *World, filters ...Filter) *View1[A] {
	storageA := getStorage[A](world.engine)

	var a A
	comps := []componentId{
		name(a),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View1[A]{
		world: world,
		filter: filterList,
		storageA: storageA,
	}
	return v
}

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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

// Maps the lambda function across every entity which matched the specified filters.
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

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
type View2[A,B any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
}

// Creates a View for the specified world with the specified component filters.
func Query2[A,B any](world *World, filters ...Filter) *View2[A,B] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)

	var a A
	var b B
	comps := []componentId{
		name(a),
		name(b),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View2[A,B]{
		world: world,
		filter: filterList,
		storageA: storageA,
		storageB: storageB,
	}
	return v
}

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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

// Maps the lambda function across every entity which matched the specified filters.
func (v *View2[A,B]) MapId(lambda func(id Id, a *A, b *B)) {
	v.filter.regenerate(v.world)

	var sliceA *componentSlice[A]
	var sliceB *componentSlice[B]
	var compA []A
	var compB []B
	var a *A
	var b *B
	for _, archId := range v.filter.archIds {
		sliceA, _ = v.storageA.slice[archId]
		sliceB, _ = v.storageB.slice[archId]

		lookup, ok := v.world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		ids := lookup.id


		// TODO - this flattened version causes a mild performance hit. Switch to code generation and use Option 2 eventually
		if sliceA != nil {
			compA = sliceA.comp
		} else {
			compA = nil
		}
		if sliceB != nil {
			compB = sliceB.comp
		} else {
			compB = nil
		}

		a = nil
		b = nil
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			if compA != nil { a = &compA[i] }
			if compB != nil { b = &compB[i] }
			lambda(ids[i], a, b)
		}

	// 	// Option 2
	// 	if compA == nil && compB == nil {
	// 		return
	// 	} else if compA != nil && compB == nil {
	// 		if len(ids) != len(compA) {
	// 			panic("ERROR - Bounds don't match")
	// 		}
	// 		for i := range ids {
	// 			if ids[i] == InvalidEntity { continue }
	// 			lambda(ids[i], &compA[i], nil)
	// 		}
	// 	} else if compA == nil && compB != nil {
	// 		if len(ids) != len(compB) {
	// 			panic("ERROR - Bounds don't match")
	// 		}
	// 		for i := range ids {
	// 			if ids[i] == InvalidEntity { continue }
	// 			lambda(ids[i], nil, &compB[i])
	// 		}
	// 	} else if compA != nil && compB != nil {
	// 		if len(ids) != len(compA) || len(ids) != len(compB) {
	// 			panic("ERROR - Bounds don't match")
	// 		}
	// 		for i := range ids {
	// 			if ids[i] == InvalidEntity { continue }
	// 			lambda(ids[i], &compA[i], &compB[i])
	// 		}
	// 	}
	}

		// Original - doesn't handle optional
	// for _, archId := range v.filter.archIds {
	// 	aSlice, ok := v.storageA.slice[archId]
	// 	if !ok { continue }
	// 	bSlice, ok := v.storageB.slice[archId]
	// 	if !ok { continue }

	// 	lookup, ok := v.world.engine.lookup[archId]
	// 	if !ok { panic("LookupList is missing!") }

	// 	ids := lookup.id
	// 	aComp := aSlice.comp
	// 	bComp := bSlice.comp
	// 	if len(ids) != len(aComp) || len(ids) != len(bComp) {
	// 		panic("ERROR - Bounds don't match")
	// 	}
	// 	for i := range ids {
	// 		if ids[i] == InvalidEntity { continue }
	// 		lambda(ids[i], &aComp[i], &bComp[i])
	// 	}
	// }
}

// Deprecated: This API is a tentative alternative way to map
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

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
type View3[A,B,C any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
}

// Creates a View for the specified world with the specified component filters.
func Query3[A,B,C any](world *World, filters ...Filter) *View3[A,B,C] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)

	var a A
	var b B
	var c C

	comps := []componentId{
		name(a),
		name(b),
		name(c),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View3[A,B,C]{
		world: world,
		filter: filterList,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
	}
	return v
}

// Maps the lambda function across every entity which matched the specified filters.
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

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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

// Deprecated: This API is a tentative alternative way to map
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

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
type View4[A,B,C,D any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
	storageD componentSliceStorage[D]
}

// Creates a View for the specified world with the specified component filters.
func Query4[A,B,C,D any](world *World, filters ...Filter) *View4[A,B,C,D] {
	storageA := getStorage[A](world.engine)
	storageB := getStorage[B](world.engine)
	storageC := getStorage[C](world.engine)
	storageD := getStorage[D](world.engine)

	var a A
	var b B
	var c C
	var d D
	comps := []componentId{
		name(a),
		name(b),
		name(c),
		name(d),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View4[A,B,C,D]{
		world: world,
		filter: filterList,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
	}
	return v
}

// Maps the lambda function across every entity which matched the specified filters.
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

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
type View5[A,B,C,D,E any] struct {
	world *World
	filter filterList
	storageA componentSliceStorage[A]
	storageB componentSliceStorage[B]
	storageC componentSliceStorage[C]
	storageD componentSliceStorage[D]
	storageE componentSliceStorage[E]
}

// Creates a View for the specified world with the specified component filters.
func Query5[A,B,C,D,E any](world *World, filters ...Filter) *View5[A,B,C,D,E] {
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
	comps := []componentId{
		name(a),
		name(b),
		name(c),
		name(d),
		name(e),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View5[A,B,C,D,E]{
		world: world,
		filter: filterList,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
		storageE: storageE,
	}
	return v
}

// Maps the lambda function across every entity which matched the specified filters.
func (v *View5[A,B,C,D,E]) MapId(lambda func(id Id, a *A, b *B, c *C, d *D, e *E)) {
	v.filter.regenerate(v.world)

	for _, archId := range v.filter.archIds {
		aSlice, ok := v.storageA.slice[archId]
		// if !ok { continue }
		bSlice, ok := v.storageB.slice[archId]
		// if !ok { continue }
		cSlice, ok := v.storageC.slice[archId]
		// if !ok { continue }
		dSlice, ok := v.storageD.slice[archId]
		// if !ok { continue }
		eSlice, ok := v.storageE.slice[archId]
		// if !ok { continue }

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

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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

// Represents a view of data in a specific world. Provides access to the components specified in the generic block
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

// Creates a View for the specified world with the specified component filters.
func Query6[A,B,C,D,E,F any](world *World, filters ...Filter) *View6[A,B,C,D,E,F] {
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
	comps := []componentId{
		name(a),
		name(b),
		name(c),
		name(d),
		name(e),
		name(f),
	}
	filterList := newFilterList(comps, filters...)
	filterList.regenerate(world)

	v := &View6[A,B,C,D,E,F]{
		world: world,
		filter: filterList,
		storageA: storageA,
		storageB: storageB,
		storageC: storageC,
		storageD: storageD,
		storageE: storageE,
		storageF: storageF,
	}
	return v
}

// Maps the lambda function across every entity which matched the specified filters.
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

// Reads a pointer to the underlying component at the specified id.
// Read will return even if the specified id doesn't match the filter list
// Read will return the value if it exists, else returns nil.
// If you execute any ecs.Write(...) or ecs.Delete(...) this pointer may become invalid.
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
*/
