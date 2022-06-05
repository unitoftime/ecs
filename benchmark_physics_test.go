package ecs

import (
	"testing"
	"math/rand"
)

// Before we applied monomorphization techniques described here: https://planetscale.com/blog/generics-can-make-your-go-code-slower
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkPhysicsEcsMap-12                      	     356	   3444304 ns/op	      96 B/op	       7 allocs/op
// BenchmarkPhysicsEcsCleanMap-12                 	     333	   3552686 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewMap-12                  	     333	   3655658 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIter-12                 	     258	   4610706 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIter2-12                	     224	   5318517 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIterChunk-12            	     591	   1986197 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIterChunkForLoop-12     	     602	   1965277 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIterChunkCleanest-12    	     597	   1966483 ns/op	     264 B/op	      11 allocs/op
// BenchmarkPhysicsEcsViewIterChunk2-12           	     280	   4356459 ns/op	     264 B/op	      11 allocs/op
// BenchmarkDataGeneric-12                        	     613	   1959678 ns/op	       0 B/op	       0 allocs/op
// BenchmarkDataGeneric2-12                       	     616	   1978393 ns/op	       0 B/op	       0 allocs/op
// BenchmarkDataGeneric3-12                       	     595	   1941133 ns/op	       0 B/op	       0 allocs/op
// BenchmarkData-12                               	     628	   1942784 ns/op	       0 B/op	       0 allocs/op
// BenchmarkData2-12                              	     620	   2013305 ns/op	       0 B/op	       0 allocs/op
// BenchmarkDataFastest-12                        	     606	   1938433 ns/op	       0 B/op	       0 allocs/op


// goos: linux
// goarch: amd64
// pkg: github.com/unitoftime/ecs
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkPhysicsEcsMap-12               	     357	   3333690 ns/op
// BenchmarkPhysicsEcsInternalAccess-12    	     609	   1954592 ns/op
// BenchmarkPhysicsEcsViewMap-12           	     328	   3658008 ns/op
// BenchmarkPhysicsEcsViewIter-12          	     262	   4567038 ns/op
// BenchmarkPhysicsEcsViewIter2-12         	     220	   5343216 ns/op
// BenchmarkPhysicsEcsViewIterChunk-12     	     594	   1925341 ns/op
// BenchmarkPhysicsEcsViewIterChunk2-12    	     270	   4303717 ns/op
// BenchmarkPhysicsGeneric-12                    428	   2792891 ns/op
// BenchmarkPhysicsSlice-12                	     628	   1952861 ns/op

type Position struct {
	X, Y, Z float32
}

type Velocity struct {
	X, Y, Z float32
}

func setupPhysics(size int) *World {
	world := NewWorld()
	// Register[Position](world)
	// Register[Velocity](world)

	scale := float32(100.0)
	for i := 0; i < size; i++ {
		id := world.NewId()

		Write(world, id,
			C(Position{
				scale * rand.Float32(),
				scale * rand.Float32(),
				scale * rand.Float32(),
			}),
			C(Velocity{
				scale * rand.Float32(),
				scale * rand.Float32(),
				scale * rand.Float32(),
			}))

	}
	return world
}

var dt = float32(0.001)

func physicsTick(id Id, pos *Position, vel *Velocity) {
	pos.X += vel.X * dt
	pos.Y += vel.Y * dt
	pos.Z += vel.Z * dt
	// TODO - writeback?
}

func physicsTick2(id Id, pos *Position, vel *Velocity) {

	pos.X += vel.X * dt
	pos.Y += vel.Y * dt
	pos.Z += vel.Z * dt
	// TODO - writeback?
}

func BenchmarkPhysicsEcsMap(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Map2[Position, Velocity](world, physicsTick)
	}
}

func BenchmarkPhysicsEcsCleanMap(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CleanMap2[Position, Velocity](world, physicsTick)
	}
}

// func BenchmarkPhysicsEcsInternalAccess(b *testing.B) {
// 	world := setupPhysics(1e6)
// 	// fmt.Println(len(pos), len(vel))
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		pos := getInternalSlice[Position](world, 2)
// 		vel := getInternalSlice[Velocity](world, 2)
// 		// Map2[Position, Velocity](world, physicsTick)
// 		mapFuncPhy(pos, vel, physicsTick)
// 	}
// }

func BenchmarkPhysicsEcsViewMap(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		view.Map(physicsTick)
	}
}

func BenchmarkPhysicsEcsViewMapFunctionalType(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2F[Position, Velocity](world, physicsTick)
		view.Map()
	}
}

func BenchmarkPhysicsEcsViewIter(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		for {
			id, pos, vel, ok := view.Iter()
			if !ok { break }
			physicsTick(id, &pos, &vel)
		}
	}
}

func BenchmarkPhysicsEcsViewIter2(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		var id Id
		var pos Position
		var vel Velocity
		for {
			ok := view.Iter2(&id, &pos, &vel)
			if !ok { break }
			physicsTick(id, &pos, &vel)
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunk(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		for {
			id, pos, vel, ok := view.IterChunk()
			if !ok { break }
			// fmt.Println(len(id))

			mapFuncPhy(id, pos, vel, physicsTick)

			// for i := range id {
			// 	physicsTick(id[i], pos[i], vel[i])
			// }
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunkForLoop(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		for {
			id, pos, vel, ok := view.IterChunk()
			if !ok { break }
			for j := range id {
				pos[j].X += vel[j].X * dt
				pos[j].Y += vel[j].Y * dt
				pos[j].Z += vel[j].Z * dt
				// physicsTick(id[j], &pos[j], &vel[j])
			}
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunkCleanest(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		view := ViewAll2[Position, Velocity](world)

		// view.Map(physicsTick)
		// id, pos, vel := view.IterChunkClean()
		// map2(id, pos, vel, physicsTick)

		for view.Ok() {
			id, pos, vel := view.IterChunkClean()

			// mapFuncPhyGen(id, pos, vel, physicsTick)

			for j := range id {
				physicsTick(id[j], &pos[j], &vel[j])
			}
		}
	}
}

func SpecialMap2NonGen2(world *World, lambda func(Id, *Position, *Velocity)) {
	view := ViewAll2[Position, Velocity](world)
	for view.Ok() {
		id, pos, vel := view.IterChunkClean()
		mapFuncPhyGen(id, pos, vel, physicsTick)
		// for j := range id {
		// 	lambda(id[j], &pos[j], &vel[j])
		// }
	}
}

func BenchmarkPhysicsEcsSpecialMapGetAllSlices(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		ids, positions, velocities := view.GetAllSlices()

		// ArchetypeMap(ids, positions, velocities, physicsTick)

		for ii := range ids {
			SliceMap2(ids[ii], positions[ii], velocities[ii], physicsTick)

			// mapFuncPhyGen(ids[ii], positions[ii], velocities[ii], physicsTick)

			// idList := ids[ii]
			// posList := positions[ii]
			// velList := velocities[ii]
			// for iii := range idList {
			// 	// physicsTick(ids[ii][iii], &positions[ii][iii], &velocities[ii][iii])
			// 	physicsTick(idList[iii], &posList[iii], &velList[iii])
			// }
		}
	}
}

func BenchmarkPhysicsEcsSpecialMap(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SpecialMap2NonGen2(world, physicsTick)
	}
}

func BenchmarkPhysicsEcsSpecialMapFast(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		for view.Ok() {
			id, pos, vel := view.IterChunkClean()
			mapFuncPhy(id, pos, vel, physicsTick)
			// for j := range id {
			// 	physicsTick(id[j], &pos[j], &vel[j])
			// }
		}
	}
}

func BenchmarkPhysicsEcsViewIteratorClean(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		view := ViewAll2[Position, Velocity](world)

		for {
			id, pos, vel, ok := view.Iter3()
			if !ok { break }
			physicsTick(id, pos, vel)
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunkCleanestest(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CleanMap2(world, physicsTick)
	}
}


func simpleMap(world *World, lambda func(id Id, pos Position, vel Velocity)) {
	view := ViewAll2[Position, Velocity](world)
	for {
		id, pos, vel, ok := view.IterChunk()
		if !ok { break }
		// fmt.Println(len(id))
		// mapFuncPhy(pos, vel, physicsTick)
		for i := range id {
			// physicsTick(id[i], pos[i], vel[i])
			lambda(id[i], pos[i], vel[i])
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunk2(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		simpleMap(world, func(id Id, pos Position, vel Velocity) {
			physicsTick(id, &pos, &vel)
		})
	}
}

// func BenchmarkPhysicsEcsViewMapPtr(b *testing.B) {
// 	world := setupPhysics(1e6)
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		RwMap2[RW[Position], RW[Velocity], Position, Velocity](world, physicsTick2)
// 	}
// }

// func BenchmarkPhysicsEcsViewMapPtr2(b *testing.B) {
// 	world := setupPhysics(1e6)
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		RwMap2[RO[Position], RO[Velocity], Position, Velocity](world, physicsTick)
// 	}
// }

func mapFuncPhy(id []Id, pos []Position, vel []Velocity, f func(id Id, pos *Position, vel *Velocity)) {
	for j := range id {
		f(id[j], &pos[j], &vel[j])
	}
}

// func mapFuncPhyGen[A any, B any](id []Id, aa []A, bb []B, f func(id Id, x *A, y *B)) {
func mapFuncPhyGen[A any, B any, F func(Id, *A, *B)](id []Id, aa []A, bb []B, f F) {
	// ids := id
	// aaa := aa
	// bbb := bb
	// for i := range ids {
	// 	f(ids[i], &aaa[i], &bbb[i])
	// }

	for j := range id {
		f(id[j], &aa[j], &bb[j])
	}
}

type Data struct {
	ids []Id
	pos []Position
	vel []Velocity
}

func NewData() *Data {
	return &Data{
		ids: make([]Id, 1e6),
		pos: make([]Position, 1e6),
		vel: make([]Velocity, 1e6),
	}
}

func BenchmarkDataGeneric(b *testing.B) {
	d := NewData()
	ids := d.ids
	aa := d.pos
	bb := d.vel
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhyGen(ids, aa, bb, physicsTick)
	}
}

func BenchmarkDataGeneric2(b *testing.B) {
	d := NewData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhyGen(d.ids, d.pos, d.vel, physicsTick)
	}
}

func BenchmarkDataGeneric3(b *testing.B) {
	ids := make([]Id, 1e6)
	aa := make([]Position, 1e6)
	bb := make([]Velocity, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhyGen(ids, aa, bb, physicsTick)
	}
}

func BenchmarkData(b *testing.B) {
	ids := make([]Id, 1e6)
	aa := make([]Position, 1e6)
	bb := make([]Velocity, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhy(ids, aa, bb, physicsTick)
	}
}

func mapPhysicsTick(id []Id, pos []Position, vel []Velocity, f func(id Id, pos *Position, vel *Velocity)) {
	for j := range id {
		f(id[j], &pos[j], &vel[j])
	}
}
func BenchmarkData2(b *testing.B) {
	d := NewData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapPhysicsTick(d.ids, d.pos, d.vel, physicsTick)
	}
}

func BenchmarkDataFastest(b *testing.B) {
	d := NewData()
	ids := d.ids
	pos := d.pos
	vel := d.vel
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := range d.ids {
			physicsTick(ids[j], &pos[j], &vel[j])
		}
	}
}

