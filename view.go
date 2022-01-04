package ecs

import (
	// "fmt"
	// "reflect"
)

type View struct {
	world *World
	components []any
	// readonly []bool
}

func ViewAll(world *World, comp ...any) *View {
	return &View{
		world: world,
		components: comp,
	}
}

func getAllStorages(world *World, comp ...any) []Storage {
	storages := make([]Storage, 0)
	for i := range comp {
		s := world.archEngine.GetStorage(comp[i])
		storages = append(storages, s)
	}
	return storages
}

func Map[A any](world *World, lambda func(id Id, a A)) {
	var a A
	archIds := world.archEngine.Filter(a)
	storages := getAllStorages(world, a)

	for _, archId := range archIds {
		aList := GetStorageList[A](storages[0], archId)

		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		for i := range lookup.Ids {
			lambda(lookup.Ids[i], aList[i])
		}
	}
}

func Map2[A, B any](world *World, lambda func(id Id, a *A, b *B)) {
	var a A
	var b B
	archIds := world.archEngine.Filter(a, b)
	storages := getAllStorages(world, a, b)

	// aPtr := (reflect.ValueOf(a).Kind() == reflect.Ptr)
	// bPtr := (reflect.ValueOf(b).Kind() == reflect.Ptr)

	for _, archId := range archIds {
		aList := GetStorageList[A](storages[0], archId)
		bList := GetStorageList[B](storages[1], archId)

		// var aIter Iterator[A] = ValueIterator[A]{aList}
		// var bIter Iterator[B] = ValueIterator[B]{bList}


		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		for i := range lookup.Ids {
			lambda(lookup.Ids[i], &aList[i], &bList[i])

			// lambda(lookup.Ids[i], aIter.Get(i), bIter.Get(i))

			// if !aPtr && !bPtr {
			// 	lambda(lookup.Ids[i], aList[i], bList[i])
			// } else if aPtr && !bPtr {
			// 	lambda(lookup.Ids[i], &aList[i], bList[i])
			// } else if !aPtr && !bPtr {
			// 	lambda(lookup.Ids[i], aList[i], &bList[i])
			// } else if aPtr && bPtr {
			// 	lambda(lookup.Ids[i], &aList[i], &bList[i])
			// }
		}
	}
}
