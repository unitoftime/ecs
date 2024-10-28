package ecs

import "testing"

func TestMainRewriteArch(t *testing.T) {
	world := NewWorld()

	id := world.NewId()

	// --- position ---
	{
		val, ok := Read[position](world, id)
		check(t, !ok)
		compare(t, position{}, val)
		world.Write(id, C(position{1, 1, 1}))

		val, ok = Read[position](world, id)
		check(t, ok)
		compare(t, position{1, 1, 1}, val)
	}

	// --- Velocity ---
	{
		val, ok := Read[velocity](world, id)
		check(t, !ok)
		compare(t, velocity{}, val)
		world.Write(id, C(velocity{2, 2, 2}))
		val, ok = Read[velocity](world, id)
		check(t, ok)
		compare(t, velocity{2, 2, 2}, val)
	}

	// --- Acceleration ---
	{
		val, ok := Read[acceleration](world, id)
		check(t, !ok)
		compare(t, acceleration{}, val)
		world.Write(id, C(acceleration{3, 3, 3}))
		val, ok = Read[acceleration](world, id)
		check(t, ok)
		compare(t, acceleration{3, 3, 3}, val)
	}

	// --- Radius ---
	{
		val, ok := Read[radius](world, id)
		check(t, !ok)
		compare(t, radius{}, val)
		world.Write(id, C(radius{4}))
		val, ok = Read[radius](world, id)
		check(t, ok)
		compare(t, radius{4}, val)
	}
}

func BenchmarkMainRewriteArch(b *testing.B) {
	world := NewWorld()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		id := world.NewId()
		world.Write(id, C(position{}))
		world.Write(id, C(velocity{}))
		world.Write(id, C(acceleration{}))
		world.Write(id, C(radius{}))
	}
}
