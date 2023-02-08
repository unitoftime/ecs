package ecs

func MapFunc[A any, F func(Id, *A)](world *World) func(F) {
	var a A
	aStorage := GetStorage[A](world.engine)

	return func(lambda F) {
		archIds := world.engine.Filter(a)

		for _, archId := range archIds {
			aSlice, ok := aStorage.slice[archId]
			if !ok { continue }

			lookup, ok := world.engine.lookup[archId]
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
}

func MapFunc2[A, B any, F func(Id, *A, *B)](world *World) func(F) {
	// This one is slower 335 ms
	var a A
	var b B
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)

	return func(lambda F) {
		archIds := world.engine.Filter(a, b)

		for _, archId := range archIds {
			aSlice, ok := aStorage.slice[archId]
			if !ok { continue }
			bSlice, ok := bStorage.slice[archId]
			if !ok { continue }

			lookup, ok := world.engine.lookup[archId]
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
}

func MapFunc3[A, B, C any, F func(id Id, a *A, b *B, c *C)](world *World) func(F) {
	var a A
	var b B
	var c C
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)
	cStorage := GetStorage[C](world.engine)

	return func(lambda F) {
		archIds := world.engine.Filter(a, b, c)

		for _, archId := range archIds {
			aSlice, ok := aStorage.slice[archId]
			if !ok { continue }
			bSlice, ok := bStorage.slice[archId]
			if !ok { continue }
			cSlice, ok := cStorage.slice[archId]
			if !ok { continue }

			lookup, ok := world.engine.lookup[archId]
			if !ok { panic("LookupList is missing!") }

			ids := lookup.id
			aComp := aSlice.comp
			bComp := bSlice.comp
			cComp := cSlice.comp
			if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp){
				panic("ERROR - Bounds don't match")
			}
			for i := range ids {
				if ids[i] == InvalidEntity { continue }
				lambda(ids[i], &aComp[i], &bComp[i], &cComp[i])
			}
		}
	}
}

func MapFunc4[A, B, C, D any](world *World, lambda func(id Id, a *A, b *B, c *C, d *D)) {
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
		// for i, id := range lookup.id {
		// 	if id == InvalidEntity { continue }
		// 	lambda(lookup.id[i], &aSlice.comp[i], &bSlice.comp[i], &cSlice.comp[i], &dSlice.comp[i])
		// }
	}
}

func MapFunc5[A, B, C, D, E any](world *World, lambda func(id Id, a *A, b *B, c *C, d *D, e *E)) {
	var a A
	var b B
	var c C
	var d D
	var e E
	archIds := world.engine.Filter(a, b, c, d, e)

	// storages := getAllStorages(world, a)
	aStorage := GetStorage[A](world.engine)
	bStorage := GetStorage[B](world.engine)
	cStorage := GetStorage[C](world.engine)
	dStorage := GetStorage[D](world.engine)
	eStorage := GetStorage[E](world.engine)

	for _, archId := range archIds {
		aSlice, ok := aStorage.slice[archId]
		if !ok { continue }
		bSlice, ok := bStorage.slice[archId]
		if !ok { continue }
		cSlice, ok := cStorage.slice[archId]
		if !ok { continue }
		dSlice, ok := dStorage.slice[archId]
		if !ok { continue }
		eSlice, ok := eStorage.slice[archId]
		if !ok { continue }

		lookup, ok := world.engine.lookup[archId]
		if !ok { panic("LookupList is missing!") }

		ids := lookup.id
		aComp := aSlice.comp
		bComp := bSlice.comp
		cComp := cSlice.comp
		dComp := dSlice.comp
		eComp := eSlice.comp
		if len(ids) != len(aComp) || len(ids) != len(bComp) || len(ids) != len(cComp) || len(ids) != len(dComp) || len(ids) != len(eComp){
			panic("ERROR - Bounds don't match")
		}
		for i := range ids {
			if ids[i] == InvalidEntity { continue }
			lambda(ids[i], &aComp[i], &bComp[i], &cComp[i], &dComp[i], &eComp[i])
		}
		// for i, id := range lookup.id {
		// 	if id == InvalidEntity { continue }
		// 	lambda(lookup.id[i], &aSlice.comp[i], &bSlice.comp[i], &cSlice.comp[i], &dSlice.comp[i])
		// }
	}
}
