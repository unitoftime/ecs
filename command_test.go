package ecs

import (
	"testing"
)

func TestCommandExecution(t *testing.T) {
	world := NewWorld()

	cmd := NewCommand(world)
	query := Query2[position, velocity](world)

	// Write position
	id := world.NewId()
	pos := position{1, 1, 1}
	WriteCmd(cmd, id, pos)
	cmd.Execute()

	// Check position and velocity
	posOut, velOut := query.Read(id)
	compare(t, *posOut, pos)
	compare(t, velOut, nil)

	// Write velocity
	vel := velocity{2, 2, 2}
	WriteCmd(cmd, id, vel)
	cmd.Execute()

	// Check position and velocity
	posOut, velOut = query.Read(id)
	compare(t, *posOut, pos)
	compare(t, *velOut, vel)

	compare(t, world.engine.count(position{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(acceleration{}), 0)

	count := 0
	query.MapId(func(id Id, p *position, v *velocity) {
		count++
	})
	compare(t, count, 1)

	// count = 0
	// view := ViewAll2[position, velocity](world)
	// for {
	// 	_, _, _, ok := view.Iter()
	// 	if !ok { break }
	// 	count++
	// }
	// compare(t, count, 1)
}

// Note To self: Before I changed how archetype ids were generated
// goos: linux
// goarch: amd64
// pkg: github.com/unitoftime/ecs
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkAddEntityWrite-12         	    1206	   1180224 ns/op	  920468 B/op	   14063 allocs/op
// BenchmarkAddEntity-12              	     838	   1530854 ns/op	 1137792 B/op	   18087 allocs/op
// BenchmarkAddEntityCached-12        	    1189	   1059531 ns/op	  800969 B/op	    7064 allocs/op
// BenchmarkAddEntityCommands-12      	     910	   1417318 ns/op	 1017217 B/op	   17085 allocs/op
// BenchmarkAddEntityViaBundles-12    	    1220	    991152 ns/op	  833882 B/op	   10063 allocs/op

var addEntSize = 1000
func BenchmarkAddEntityWrite(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()

			Write(world, id,
				C(position{1, 2, 3}),
				C(velocity{4, 5, 6}),
				C(acceleration{7, 8, 9}),
				C(radius{10}),
			)
		}
	}
}

func BenchmarkAddEntity(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			ent := NewEntity(
				C(position{1, 2, 3}),
				C(velocity{4, 5, 6}),
				C(acceleration{7, 8, 9}),
				C(radius{10}),
			)

			id := world.NewId()
			ent.Write(world, id)
		}
	}
}

func BenchmarkAddEntityCached(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	ent := NewEntity(
		C(position{1, 2, 3}),
		C(velocity{4, 5, 6}),
		C(acceleration{7, 8, 9}),
		C(radius{10}),
	)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			ent.Write(world, id)
		}
	}
}

func BenchmarkAddEntityCommands(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	cmd := NewCommand(world)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			WriteCmd(cmd, id, position{1, 2, 3})
			WriteCmd(cmd, id, velocity{4, 5, 6})
			WriteCmd(cmd, id, acceleration{7, 8, 9})
			WriteCmd(cmd, id, radius{10})
			cmd.Execute()
		}
	}
}

func BenchmarkAddEntityViaBundles(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	posBundle := NewBundle[position]()
	velBundle := NewBundle[velocity]()
	accBundle := NewBundle[acceleration]()
	radBundle := NewBundle[radius]()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			Write(world, id,
				posBundle.New(position{1, 2, 3}),
				velBundle.New(velocity{4, 5, 6}),
				accBundle.New(acceleration{7, 8, 9}),
				radBundle.New(radius{10}),
			)
		}
	}
}

// func BenchmarkAddEntityViaBundles2(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	bundle := NewBundle4[postion, velocity, acceleration, radius](world)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			bundle.Write(id,
// 				position{1, 2, 3},
// 				velocity{4, 5, 6},
// 				acceleration{7, 8, 9},
// 				radius{10},
// 			)
// 		}
// 	}
// }
