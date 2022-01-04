package ecs

import (
	"fmt"
	"testing"
	"math/rand"
)

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

func physicsTick(id Id, pos Position, vel Velocity) {
	dt := float32(0.001)
	pos.X += vel.X * dt
	pos.Y += vel.Y * dt
	pos.Z += vel.Z * dt
	// TODO - writeback?
}

func physicsTick2(id Id, pos *Position, vel Velocity) {
	dt := float32(0.001)
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

func BenchmarkPhysicsEcsInternalAccess(b *testing.B) {
	world := setupPhysics(1e6)
	// fmt.Println(len(pos), len(vel))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pos := getInternalSlice[Position](world, 2)
		vel := getInternalSlice[Velocity](world, 2)
		// Map2[Position, Velocity](world, physicsTick)
		mapFuncPhy(pos, vel, physicsTick)
	}
}

func BenchmarkPhysicsEcsViewMap(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		view.Map(physicsTick)
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
			physicsTick(id, pos, vel)
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
			physicsTick(id, pos, vel)
		}
	}
}

func BenchmarkPhysicsEcsViewIterChunk(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		view := ViewAll2[Position, Velocity](world)
		for {
			_, pos, vel, ok := view.IterChunk()
			if !ok { break }
			// fmt.Println(len(id))
			mapFuncPhy(pos, vel, physicsTick)
			// for i := range id {
			// 	physicsTick(id[i], pos[i], vel[i])
			// }
		}
	}
}

// func BenchmarkPhysicsEcsViewMapPtr(b *testing.B) {
// 	world := setupPhysics(1e6)
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		view := ViewAll2[*Position, Velocity](world)
// 		view.Map(physicsTick2)
// 	}
// }

func mapFuncPhy(pos []Position, vel []Velocity, f func(id Id, pos Position, vel Velocity)) {
	for j := range pos {
		f(Id(j), pos[j], vel[j])
	}
}

func mapFuncPhyGen[A any, B any](id []Id, aa []A, bb []B, f func(id Id, x A, y B)) {
	for j := range aa {
		f(id[j], aa[j], bb[j])
	}
}

func BenchmarkPhysicsGeneric(b *testing.B) {
	id := make([]Id, 1e6)
	aa := make([]Position, 1e6)
	bb := make([]Velocity, 1e6)
	fmt.Println(len(aa), len(bb))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhyGen(id, aa, bb, physicsTick)
	}
}

func BenchmarkPhysicsSlice(b *testing.B) {
	aa := make([]Position, 1e6)
	bb := make([]Velocity, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhy(aa, bb, physicsTick)
	}
}
