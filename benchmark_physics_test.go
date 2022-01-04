package ecs

import (
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
	Register[Position](world)
	Register[Velocity](world)

	scale := float32(100.0)
	for i := 0; i < size; i++ {
		id := world.NewId()

		Write(world, id,
			Position{
				scale * rand.Float32(),
				scale * rand.Float32(),
				scale * rand.Float32(),
			},
			Velocity{
				scale * rand.Float32(),
				scale * rand.Float32(),
				scale * rand.Float32(),
			})

	}
	return world
}

func physicsTick(id Id, pos *Position, vel *Velocity) {
	dt := float32(0.001)
	pos.X += vel.X * dt
	pos.Y += vel.Y * dt
	pos.Z += vel.Z * dt
	// TODO - writeback?
}

func BenchmarkPhysicsEcs(b *testing.B) {
	world := setupPhysics(1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Map2[Position, Velocity](world, physicsTick)
	}
}


func mapFuncPhy(pos []Position, vel []Velocity, f func(id Id, pos *Position, vel *Velocity)) {
	for j := range pos {
		f(Id(j), &pos[j], &vel[j])
	}
}

func mapFuncPhyGen[A any, B any](aa []A, bb []B, f func(id Id, x *A, y *B)) {
	for j := range aa {
		f(Id(j), &aa[j], &bb[j])
	}
}

func BenchmarkPhysicsGeneric(b *testing.B) {
	aa := make([]Position, 1e6)
	bb := make([]Velocity, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhyGen[Position, Velocity](aa, bb, physicsTick)
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
