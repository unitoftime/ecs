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

// Refactor 1: Pushing componentId outward
// goos: linux
// goarch: amd64
// pkg: github.com/unitoftime/ecs
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkAddEntityWrite-12         	    1198	   1133353 ns/op	  916486 B/op	   12064 allocs/op
// BenchmarkAddEntity-12              	     841	   1441826 ns/op	 1127953 B/op	   16087 allocs/op
// BenchmarkAddEntityCached-12        	    1276	    903884 ns/op	  696180 B/op	    5060 allocs/op
// BenchmarkAddEntityCommands-12      	     913	   1382452 ns/op	 1007375 B/op	   15084 allocs/op
// BenchmarkAddEntityViaBundles-12    	    1388	    924249 ns/op	  771132 B/op	    8065 allocs/op


// Refactor 2: replacing archId trie with a bitmap
// BenchmarkAddEntityWrite-12         	    1519	    865592 ns/op	  780596 B/op	   10074 allocs/op
// BenchmarkAddEntity-12              	    1044	   1192381 ns/op	 1057146 B/op	   14073 allocs/op
// BenchmarkAddEntityCached-12        	    1797	    734056 ns/op	  640552 B/op	    3085 allocs/op
// BenchmarkAddEntityCommands-12      	    1118	   1107654 ns/op	  881888 B/op	   13068 allocs/op
// BenchmarkAddEntityViaBundles-12    	    1918	    746616 ns/op	  738067 B/op	    6080 allocs/op

// With slice as entity storage instead of map
// BenchmarkAddEntityWrite-12         	    1498	    845244 ns/op	  788350 B/op	   10072 allocs/op
// BenchmarkAddEntity-12              	    1453	    853206 ns/op	  805814 B/op	   10069 allocs/op
// BenchmarkAddEntityCached-12        	    2227	    608231 ns/op	  555755 B/op	    2069 allocs/op
// BenchmarkAddEntityCommands-12      	    1131	   1108961 ns/op	  875865 B/op	   13068 allocs/op
// BenchmarkAddEntityViaBundles-12    	    1909	    716650 ns/op	  740988 B/op	    6080 allocs/op

// only add arch to archset if we just created it
// BenchmarkAddEntityWrite-12         	    1580	    838312 ns/op	  759720 B/op	   10079 allocs/op
// BenchmarkAddEntity-12              	    1542	    841495 ns/op	  772783 B/op	   10076 allocs/op
// BenchmarkAddEntityCached-12        	    2550	    587830 ns/op	  567675 B/op	    2060 allocs/op
// BenchmarkAddEntityCommands-12      	    1174	   1096164 ns/op	  944671 B/op	   13065 allocs/op
// BenchmarkAddEntityViaBundles-12    	    2023	    694800 ns/op	  706566 B/op	    6076 allocs/op

// preallocated array for archset holder
// BenchmarkAddEntityWrite-12         	    1765	    767105 ns/op	  778658 B/op	   10087 allocs/op
// BenchmarkAddEntity-12              	    1735	    755672 ns/op	  788496 B/op	   10088 allocs/op
// BenchmarkAddEntityCached-12        	    3001	    527071 ns/op	  585524 B/op	    2073 allocs/op
// BenchmarkAddEntityCommands-12      	    1272	    989622 ns/op	  897973 B/op	   13060 allocs/op
// BenchmarkAddEntityViaBundles-12    	    2058	    611106 ns/op	  696776 B/op	    6074 allocs/op

// filterlist now uses array instead of maps
// BenchmarkAddEntityWrite-12         	    1838	    783168 ns/op	  844602 B/op	   10083 allocs/op
// BenchmarkAddEntity-12              	    1726	    752888 ns/op	  791532 B/op	   10089 allocs/op
// BenchmarkAddEntityCached-12        	    3126	    512635 ns/op	  563923 B/op	    2078 allocs/op
// BenchmarkAddEntityCommands-12      	    1263	    989989 ns/op	  901981 B/op	   13060 allocs/op
// BenchmarkAddEntityViaBundles-12    	    2281	    632377 ns/op	  744278 B/op	    6067 allocs/op

// replaced more things with arrays instead of maps
// BenchmarkAddEntityWrite-12         	    1960	    740365 ns/op	  805078 B/op	   10078 allocs/op
// BenchmarkAddEntity-12              	    1909	    735657 ns/op	  820935 B/op	   10080 allocs/op
// BenchmarkAddEntityCached-12        	    3496	    453399 ns/op	  583412 B/op	    2088 allocs/op
// BenchmarkAddEntityCommands-12      	    1384	    927160 ns/op	  868407 B/op	   13065 allocs/op
// BenchmarkAddEntityViaBundles-12    	    2748	    518987 ns/op	  655500 B/op	    6064 allocs/op

// BenchmarkAddEntityWrite-12          	    2101	    667549 ns/op	  758943 B/op	    8073 allocs/op
// BenchmarkAddEntity-12               	    2080	    672744 ns/op	  764670 B/op	    8073 allocs/op
// BenchmarkAddEntityCached-12         	    3726	    437371 ns/op	  617807 B/op	      82 allocs/op
// BenchmarkAddEntityCommands-12       	    1417	    894351 ns/op	  840387 B/op	   11067 allocs/op
// BenchmarkAddEntityViaBundles-12     	    2832	    505260 ns/op	  714083 B/op	    4067 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3770	    448083 ns/op	  610571 B/op	      81 allocs/op

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

	posBundle := NewBundle[position](world)
	velBundle := NewBundle[velocity](world)
	accBundle := NewBundle[acceleration](world)
	radBundle := NewBundle[radius](world)

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

func BenchmarkAddEntityViaBundles2(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	bundle := NewBundle4[position, velocity, acceleration, radius]()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			bundle.Write(world, id,
				position{1, 2, 3},
				velocity{4, 5, 6},
				acceleration{7, 8, 9},
				radius{10},
			)
		}
	}
}

// BenchmarkAddEntityViaBundles-12     	    2988	    496709 ns/op	  699920 B/op	    6072 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3024	    501376 ns/op	  666834 B/op	    6074 allocs/op

// BenchmarkAddEntityViaBundles-12     	    3012	    506969 ns/op	  695520 B/op	    6073 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3589	    447601 ns/op	  657292 B/op	    2085 allocs/op

// BenchmarkAddEntityViaBundles-12     	    2862	    504393 ns/op	  723965 B/op	    6068 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3552	    444953 ns/op	  663962 B/op	    2086 allocs/op

// Removed []componentId and just use component
// BenchmarkAddEntityViaBundles-12     	    3325	    463634 ns/op	  643600 B/op	    4086 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3982	    424222 ns/op	  578243 B/op	      77 allocs/op


// increased default allocation to 10k
// BenchmarkAddEntityViaBundles-12     	    3020	    479307 ns/op	  678231 B/op	    4074 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3673	    446787 ns/op	  626812 B/op	      83 allocs/op

// BenchmarkAddEntityViaBundles-12     	    2922	    484760 ns/op	  645464 B/op	    4070 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    3890	    453437 ns/op	  633719 B/op	      79 allocs/op

// Tests writing the same entity multiple times
func BenchmarkAddEntitySameCached(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	ent := NewEntity(
		C(position{1, 2, 3}),
		C(velocity{4, 5, 6}),
		C(acceleration{7, 8, 9}),
		C(radius{10}),
	)
	id := world.NewId()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			ent.Write(world, id)
		}
	}
}
