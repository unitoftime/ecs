package ecs

// import (
// 	"fmt"
// )

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

	// fmt.Println("Map", archIds)
	for _, archId := range archIds {
		// fmt.Println("ArchId", archId)
		aList := storages[0].(*ArchStorage[[]A, A]).list[archId]

		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		for i := range lookup.Ids {
			// fmt.Println("Index", i)
			lambda(lookup.Ids[i], aList[i])
		}
	}
}

func Map2[A, B any](world *World, lambda func(id Id, a A, b B)) {
	var a A
	var b B
	archIds := world.archEngine.Filter(a, b)
	// archIds := []ArchId{ArchId(2)}
	storages := getAllStorages(world, a, b)

	for _, archId := range archIds {
		aList := storages[0].(*ArchStorage[[]A, A]).list[archId]
		bList := storages[1].(*ArchStorage[[]B, B]).list[archId]

		lookup, ok := world.archEngine.lookup[archId]
		if !ok { panic("LookupList is missing!") }
		for i := range lookup.Ids {
			lambda(lookup.Ids[i], aList[i], bList[i])
		}
	}
}
