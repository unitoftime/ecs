package ecs

import (
	"testing"
)

// func TestCommandExecution(t *testing.T) {
// 	world := NewWorld()

// 	cmd := NewCommand(world)
// 	query := Query2[position, velocity](world)

// 	// Write position
// 	id := world.NewId()
// 	pos := position{1, 1, 1}
// 	WriteCmd(cmd, id, pos)
// 	cmd.Execute()

// 	// Check position and velocity
// 	posOut, velOut := query.Read(id)
// 	compare(t, *posOut, pos)
// 	compare(t, velOut, nil)

// 	// Write velocity
// 	vel := velocity{2, 2, 2}
// 	WriteCmd(cmd, id, vel)
// 	cmd.Execute()

// 	// Check position and velocity
// 	posOut, velOut = query.Read(id)
// 	compare(t, *posOut, pos)
// 	compare(t, *velOut, vel)

// 	compare(t, world.engine.count(position{}), 1)
// 	compare(t, world.engine.count(position{}, velocity{}), 1)
// 	compare(t, world.engine.count(position{}, velocity{}), 1)
// 	compare(t, world.engine.count(acceleration{}), 0)

// 	count := 0
// 	query.MapId(func(id Id, p *position, v *velocity) {
// 		count++
// 	})
// 	compare(t, count, 1)

// 	// count = 0
// 	// view := ViewAll2[position, velocity](world)
// 	// for {
// 	// 	_, _, _, ok := view.Iter()
// 	// 	if !ok { break }
// 	// 	count++
// 	// }
// 	// compare(t, count, 1)
// }

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

func BenchmarkAddEntitySingle(b *testing.B) {
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

func BenchmarkAddEntityMemCached(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	ent := NewEntity()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			ent.Clear()
			ent.Add(
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

// func BenchmarkAddEntityCommands(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	cmd := NewCommand(world)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			WriteCmd(cmd, id, position{1, 2, 3})
// 			WriteCmd(cmd, id, velocity{4, 5, 6})
// 			WriteCmd(cmd, id, acceleration{7, 8, 9})
// 			WriteCmd(cmd, id, radius{10})
// 			cmd.Execute()
// 		}
// 	}
// }

// func BenchmarkAddEntityViaBundles(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	posBundle := NewBundle[position](world)
// 	velBundle := NewBundle[velocity](world)
// 	accBundle := NewBundle[acceleration](world)
// 	radBundle := NewBundle[radius](world)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			Write(world, id,
// 				posBundle.New(position{1, 2, 3}),
// 				velBundle.New(velocity{4, 5, 6}),
// 				accBundle.New(acceleration{7, 8, 9}),
// 				radBundle.New(radius{10}),
// 			)
// 		}
// 	}
// }

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

// func BenchmarkAddEntityViaBundles3(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	command := NewCommand(world)
// 	BundleOne := NewBundle2[position, velocity]()
// 	BundleTwo := NewBundle2[acceleration, radius]()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			command.Spawn(
// 				BundleOne.With(
// 					position{1, 2, 3},
// 					velocity{4, 5, 6},
// 				),
// 				BundleTwo.With(
// 					acceleration{7, 8, 9},
// 					radius{10},
// 				),
// 			)

// 			// command.Execute()
// 		}
// 	}
// }

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

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			ent.Write(world, id)
		}
	}
}

func BenchmarkAddTry2EntitySameCachedThenWrite(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	query := Query4[position, velocity, acceleration, radius](world)
	ent := NewEntity(
		C(position{}),
		C(velocity{}),
		C(acceleration{}),
		C(radius{}),
	)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			ent.Write(world, id)

			pos, vel, acc, rad := query.Read(id)
			*pos = position{1, 2, 3}
			*vel = velocity{4, 5, 6}
			*acc = acceleration{7, 8, 9}
			*rad = radius{10}
		}
	}
}

// Swissmap1: inside component slice lookuplist
// BenchmarkAddEntityWrite-12          	    2535	    554449 ns/op	  654120 B/op	    4018 allocs/op
// BenchmarkAddEntity-12               	    2352	    572987 ns/op	  696081 B/op	    4020 allocs/op
// BenchmarkAddEntityCached-12         	    4146	    445121 ns/op	  554724 B/op	      22 allocs/op
// BenchmarkAddEntityCommands-12       	    1718	    778746 ns/op	  776313 B/op	    7027 allocs/op
// BenchmarkAddEntityViaBundles-12     	    2992	    473943 ns/op	  670642 B/op	    4026 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4311	    443715 ns/op	  533557 B/op	      22 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2792	    525603 ns/op	  730906 B/op	    6022 allocs/op
// BenchmarkAddEntitySameCached-12     	   15274	     78000 ns/op	       0 B/op	       0 allocs/op

// Swissmap 2: main world arch map + lookuplist
// BenchmarkAddEntityWrite-12          	    2691	    508476 ns/op	  629258 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2451	    517719 ns/op	  663984 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    4018	    446250 ns/op	  546891 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1696	    704253 ns/op	  749053 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3045	    436831 ns/op	  652619 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4462	    440346 ns/op	  595200 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2875	    499738 ns/op	  796589 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   12859	     92020 ns/op	       0 B/op	       0 allocs/op

// Using intmap
// BenchmarkAddEntityWrite-12          	    2724	    496733 ns/op	  619912 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2511	    544596 ns/op	  647454 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    3640	    368577 ns/op	  569115 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1692	    725601 ns/op	  767593 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3399	    427582 ns/op	  627948 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4722	    391206 ns/op	  558876 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2938	    498754 ns/op	  804297 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   15183	     75826 ns/op	       3 B/op	       0 allocs/op

// BenchmarkAddEntityWrite-12          	    2593	    486711 ns/op	  650183 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2510	    488575 ns/op	  657692 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    5149	    346940 ns/op	  532104 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1767	    731716 ns/op	  790303 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3566	    428045 ns/op	  737624 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    5036	    328989 ns/op	  534023 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    3046	    490770 ns/op	  836045 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   17461	     68760 ns/op	       0 B/op	       0 allocs/op

// Default map
// BenchmarkAddEntityWrite-12          	    2552	    574152 ns/op	  660288 B/op	    4037 allocs/op
// BenchmarkAddEntity-12               	    2239	    590225 ns/op	  645485 B/op	    4042 allocs/op
// BenchmarkAddEntityCached-12         	    4203	    434851 ns/op	  558904 B/op	      45 allocs/op
// BenchmarkAddEntityCommands-12       	    1645	    836026 ns/op	  832556 B/op	    7057 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3070	    491333 ns/op	  667152 B/op	    4056 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4210	    438253 ns/op	  557990 B/op	      45 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2757	    556359 ns/op	  747118 B/op	    6043 allocs/op
// BenchmarkAddEntitySameCached-12     	   16693	     72033 ns/op	       3 B/op	       0 allocs/op

// Swissmap
// BenchmarkAddEntityWrite-12          	    2650	    581867 ns/op	  637261 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2648	    540628 ns/op	  637654 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    4462	    508153 ns/op	  595204 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1676	    744876 ns/op	  750089 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3346	    455147 ns/op	  618570 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4434	    479402 ns/op	  598962 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2858	    522964 ns/op	  799992 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   12530	    106595 ns/op	       1 B/op	       0 allocs/op

// Adapted internal intmap with simplified hash function
// BenchmarkAddEntityWrite-12          	    3116	    349783 ns/op	  710300 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    3478	    348030 ns/op	  662060 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    6510	    246756 ns/op	  577626 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1892	    582171 ns/op	  839947 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    4527	    264847 ns/op	  706068 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    6100	    186840 ns/op	  600819 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    3715	    317135 ns/op	  824529 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   17976	     68139 ns/op	       3 B/op	       0 allocs/op

// With Reflection
// BenchmarkAddEntityWrite-12          	    2588	    486795 ns/op	  651222 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2587	    486373 ns/op	  651431 B/op	    4000 allocs/op
// BenchmarkAddEntityMemCached-12      	    2534	    537192 ns/op	  652525 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    5383	    323072 ns/op	  513768 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1773	    733919 ns/op	  788452 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3232	    429839 ns/op	  698811 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    5034	    334947 ns/op	  534235 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    3075	    490386 ns/op	  830276 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   17659	     66830 ns/op	       3 B/op	       0 allocs/op

// Reflection Elimination
// BenchmarkAddEntityWrite-12          	    2990	    487903 ns/op	  735514 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2934	    469310 ns/op	  678795 B/op	    4000 allocs/op
// BenchmarkAddEntityMemCached-12      	    2702	    499574 ns/op	  633361 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    5054	    318631 ns/op	  542102 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1791	    700002 ns/op	  782936 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3489	    416354 ns/op	  660322 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    5064	    331580 ns/op	  541033 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    3040	    485595 ns/op	  837256 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   17451	     68978 ns/op	       3 B/op	       0 allocs/op

//--------------------------------------------------------------------------------
// Rebaseline on new computer
//--------------------------------------------------------------------------------
// BenchmarkAddEntityWrite-12          	    2576	    534658 ns/op	  643713 B/op	    4000 allocs/op
// BenchmarkAddEntity-12               	    2626	    529261 ns/op	  643420 B/op	    4000 allocs/op
// BenchmarkAddEntityMemCached-12      	    2574	    553593 ns/op	  644119 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12         	    5134	    398114 ns/op	  533658 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12       	    1888	    740393 ns/op	  841207 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12     	    3339	    491029 ns/op	  684959 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12    	    4836	    409933 ns/op	  556111 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12    	    2828	    549254 ns/op	  812032 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12     	   16140	     73152 ns/op	       0 B/op	       0 allocs/op

// BenchmarkAddEntityWrite-12             	    2790	    515633 ns/op	  616915 B/op	    4000 allocs/op
// BenchmarkAddEntity-12                  	    2793	    513607 ns/op	  616374 B/op	    4000 allocs/op
// BenchmarkAddEntityMemCached-12         	    2691	    551973 ns/op	  635489 B/op	    4000 allocs/op
// BenchmarkAddEntityCached-12            	    5208	    396999 ns/op	  531032 B/op	       0 allocs/op
// BenchmarkAddEntityCommands-12          	    1905	    726353 ns/op	  835853 B/op	    7000 allocs/op
// BenchmarkAddEntityViaBundles-12        	    3054	    485749 ns/op	  722442 B/op	    4000 allocs/op
// BenchmarkAddEntityViaBundles2-12       	    5083	    408980 ns/op	  539012 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundles3-12       	    2956	    558824 ns/op	  854680 B/op	    6000 allocs/op
// BenchmarkAddEntitySameCached-12        	   17047	     70804 ns/op	       0 B/op	       0 allocs/op
// BenchmarkAddEntityViaBundlesTry2-12    	    3458	    419984 ns/op	  553238 B/op	       0 allocs/op

// BenchmarkAddEntityWrite-12          	    2523	    600336 ns/op
// BenchmarkAddEntity-12               	    2706	    586171 ns/op
// BenchmarkAddEntityMemCached-12      	    2595	    623440 ns/op
// BenchmarkAddEntityCached-12         	    4144	    474984 ns/op
// BenchmarkAddEntityCommands-12       	    1626	    795483 ns/op
// BenchmarkAddEntityViaBundles-12     	    3411	    545720 ns/op
// BenchmarkAddEntityViaBundles2-12    	    3888	    482885 ns/op
// BenchmarkAddEntityViaBundles3-12    	    2805	    571085 ns/op
// BenchmarkAddEntitySameCached-12     	   16396	     72816 ns/op

// BenchmarkAddEntityWrite-12                  	    2508	    580966 ns/op
// BenchmarkAddEntity-12                       	    2844	    581545 ns/op
// BenchmarkAddEntityMemCached-12              	    2426	    628340 ns/op
// BenchmarkAddEntityCached-12                 	    4381	    443793 ns/op
// BenchmarkAddEntityCommands-12               	    1676	    805492 ns/op
// BenchmarkAddEntityViaBundles-12             	    3052	    549700 ns/op
// BenchmarkAddEntityViaBundles2-12            	    4027	    465475 ns/op
// BenchmarkAddEntityViaBundles3-12            	    2733	    568369 ns/op
// BenchmarkAddEntitySameCached-12             	    4578	    474989 ns/op
// BenchmarkAddEntitySameCachedThenWrite-12    	    3069	    453504 ns/op

// type myBundle struct {
// 	pos position
// 	vel velocity
// 	acc acceleration
// 	rad radius
// }

// func BenchmarkAddEntityViaBundlesTry2(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	bundle := NewBundleTry2[position, velocity, acceleration, radius](world)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			bundle.Write(id,
// 				&position{1, 2, 3},
// 				&velocity{4, 5, 6},
// 				&acceleration{7, 8, 9},
// 				&radius{10},
// 			)
// 		}
// 	}
// }

func BenchmarkCompareBaselineEntityWrite(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	id := world.NewId()

	ent := NewEntity(
		C(position{1, 2, 3}),
		C(velocity{4, 5, 6}),
		C(acceleration{7, 8, 9}),
		C(radius{10}),
	)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			ent.Write(world, id)
		}
	}
}

func BenchmarkCompareBaselineQueryWrite(b *testing.B) {
	world := NewWorld()

	query := Query4[position, velocity, acceleration, radius](world)

	b.ResetTimer()

	id := world.NewId()

	pp := position{1, 2, 3}
	vv := velocity{4, 5, 6}
	aa := acceleration{7, 8, 9}
	rr := radius{10}

	Write(world, id, C(pp), C(vv), C(aa), C(rr))

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			p, v, a, r := query.Read(id)
			*p = pp
			*v = vv
			*a = aa
			*r = rr
		}
	}
}

func BenchmarkCompare1(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	id := world.NewId()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			ent := NewEntity(
				C(position{1, 2, 3}),
				C(velocity{4, 5, 6}),
				C(acceleration{7, 8, 9}),
				C(radius{10}),
			)

			ent.Write(world, id)
		}
	}
}

func BenchmarkCompare2(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	id := world.NewId()

	p := C(position{1, 2, 3})
	v := C(velocity{4, 5, 6})
	a := C(acceleration{7, 8, 9})
	c := C(radius{10})

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			Write(world, id, p, v, a, c)
		}
	}
}

func BenchmarkAllocateBaseline(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()

			ent := NewEntity(
				C(position{1, 2, 3}),
				C(velocity{4, 5, 6}),
				C(acceleration{7, 8, 9}),
				C(radius{10}),
			)

			ent.Write(world, id)
		}
	}
}

func BenchmarkAllocateQuery(b *testing.B) {
	world := NewWorld()

	query := Query4[position, velocity, acceleration, radius](world)

	b.ResetTimer()

	mask := buildArchMask(
		C(position{}),
		C(velocity{}),
		C(acceleration{}),
		C(radius{}),
	)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			world.allocate(id, mask)

			p, v, a, r := query.Read(id)
			*p = position{1, 2, 3}
			*v = velocity{4, 5, 6}
			*a = acceleration{7, 8, 9}
			*r = radius{10}
		}
	}
}

func BenchmarkAllocateQueryNoQuery(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	mask := buildArchMask(
		C(position{}),
		C(velocity{}),
		C(acceleration{}),
		C(radius{}),
	)
	archId := world.engine.getArchetypeId(mask)

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			index := world.allocate(id, mask)

			// // Note: Slightly faster option. Actually, I'm not so sure
			// p := positionId.getPtr(world.engine, archId, index)
			// v := velocityId.getPtr(world.engine, archId, index)
			// a := accelerationId.getPtr(world.engine, archId, index)
			// r := radiusId.getPtr(world.engine, archId, index)
			// *p = position{1, 2, 3}
			// *v = velocity{4, 5, 6}
			// *a = acceleration{7, 8, 9}
			// *r = radius{10}

			positionId.writeVal(world.engine, archId, index, position{1, 2, 3})
			velocityId.writeVal(world.engine, archId, index, velocity{1, 2, 3})
			accelerationId.writeVal(world.engine, archId, index, acceleration{1, 2, 3})
			radiusId.writeVal(world.engine, archId, index, radius{10})
		}
	}
}

// I think this is a good start. it basically makes it so you can just allocate archetypes and write them later rather than having to pass []component lists around everywhere
// Maybe something with Write1(), Write2(), Write3(), Generation? ... <- you'd still have to call 'name' and lookup the mask every frame

// func BenchmarkAllocateBundleTry2(b *testing.B) {
// 	world := NewWorld()

// 	bun := NewBundleTry2[position, velocity, acceleration, radius](world)

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			bun.Write(id,
// 				&position{1, 2, 3},
// 				&velocity{4, 5, 6},
// 				&acceleration{7, 8, 9},
// 				&radius{10},
// 			)
// 		}
// 	}
// }

// Note: Was slow
// func BenchmarkAllocateBundler(b *testing.B) {
// 	world := NewWorld()

// 	bun := &Bundler{}

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {

// 			bun.Clear()
// 			WriteComponent(bun, position{1, 2, 3})
// 			WriteComponent(bun, velocity{4, 5, 6})
// 			WriteComponent(bun, acceleration{7, 8, 9})
// 			WriteComponent(bun, radius{10})

// 			id := world.NewId()
// 			bun.Write(world, id)
// 		}
// 	}
// }

func BenchmarkAllocateBundler2(b *testing.B) {
	world := NewWorld()

	bun := &Bundler{}

	b.ResetTimer()

	pos := C(position{1, 2, 3})
	vel := C(velocity{4, 5, 6})
	acc := C(acceleration{7, 8, 9})
	rad := C(radius{10})

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			bun.Clear()

			pos.Comp = position{1, 2, 3}

			bun.Add(&pos)
			bun.Add(&vel)
			bun.Add(&acc)
			bun.Add(&rad)

			id := world.NewId()
			bun.Write(world, id)
		}
	}
}

var positionId = C(position{})
var velocityId = C(velocity{})
var accelerationId = C(acceleration{})
var radiusId = C(radius{})

type outerBundle struct {
	myBundle myBundle
}

func (m outerBundle) Unbundle(bun *Bundler) {
	m.myBundle.Unbundle(bun)
}

type myBundle struct {
	pos position
	vel velocity
	acc acceleration
	rad radius
}

func (m myBundle) Unbundle(bun *Bundler) {
	positionId.UnbundleVal(bun, m.pos)
	velocityId.UnbundleVal(bun, m.vel)
	accelerationId.UnbundleVal(bun, m.acc)
	radiusId.UnbundleVal(bun, m.rad)
}

func BenchmarkAllocateManual(b *testing.B) {
	world := NewWorld()

	bun := &Bundler{}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			m := myBundle{
				pos: position{1, 2, 3},
				vel: velocity{1, 2, 3},
				acc: acceleration{1, 2, 3},
				rad: radius{1},
			}

			bun.Clear()
			m.Unbundle(bun)
			id := world.NewId()
			bun.Write(world, id)
		}
	}
}

var myBundle2 = NewBundle4[position, velocity, acceleration, radius]()

func BenchmarkAllocateBundle4(b *testing.B) {
	world := NewWorld()

	bun := &Bundler{}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			bun.Clear()

			myBundle2.Unbundle(bun,
				position{1, 2, 3},
				velocity{1, 2, 3},
				acceleration{1, 2, 3},
				radius{1},
			)

			id := world.NewId()
			bun.Write(world, id)
		}
	}
}

type unbundler2 interface {
	Unbundle(*Bundler)
}

// type singleCmd interface {
// 	apply(*World)
// }

// type spawnCmd struct {
// 	bundler *Bundler
// }
// func (c spawnCmd) apply(world *World) {
// 	id := world.NewId()
// 	c.bundler.Write(world, id)
// }

type CmdType uint8

const (
	CmdTypeSpawn CmdType = iota
	CmdTypeWrite
	CmdTypeCustom
)

type singleCmd struct {
	Type    CmdType
	id      Id
	bundler *Bundler
}

func (c singleCmd) apply(world *World) {
	switch c.Type {
	case CmdTypeSpawn:
		id := world.NewId()
		c.bundler.Write(world, id)
	case CmdTypeWrite:
		c.bundler.Write(world, c.id)
	}
}

type cmdQueue struct {
	commands []singleCmd

	currentBundlerIndex int
	bundlers            []*Bundler
}

func newCmdQueue() *cmdQueue {
	return &cmdQueue{}
}
func (c *cmdQueue) NextBundler() *Bundler {
	if c.currentBundlerIndex >= len(c.bundlers) {
		bundler := &Bundler{}
		c.bundlers = append(c.bundlers, bundler)
		c.currentBundlerIndex = len(c.bundlers) - 1
		return bundler
	} else {
		bundler := c.bundlers[c.currentBundlerIndex]
		bundler.Clear()
		c.currentBundlerIndex++
		return bundler
	}
}

func CmdSpawn[T unbundler2](c *cmdQueue, ub T) {
	bundler := c.NextBundler()
	ub.Unbundle(bundler)
	// c.commands = append(c.commands, spawnCmd{
	// 	bundler: bundler,
	// })
	c.commands = append(c.commands, singleCmd{
		Type:    CmdTypeSpawn,
		bundler: bundler,
	})
}

func (c *cmdQueue) Spawn(ub unbundler2) {
	bundler := c.NextBundler()
	ub.Unbundle(bundler)
	// c.commands = append(c.commands, spawnCmd{
	// 	bundler: bundler,
	// })
	c.commands = append(c.commands, singleCmd{
		Type:    CmdTypeSpawn,
		bundler: bundler,
	})
}

func (c *cmdQueue) Execute(world *World) {
	for i := range c.commands {
		c.commands[i].apply(world)
	}
	c.commands = c.commands[:0]
	c.currentBundlerIndex = 0
}

func BenchmarkAllocateCommands(b *testing.B) {
	world := NewWorld()
	cmd := newCmdQueue()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			cmd.Spawn(myBundle{
				pos: position{1, 2, 3},
				vel: velocity{1, 2, 3},
				acc: acceleration{1, 2, 3},
				rad: radius{1},
			})
		}
		cmd.Execute(world)
	}
}

func BenchmarkAllocateCommands2(b *testing.B) {
	world := NewWorld()
	cmd := newCmdQueue()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			CmdSpawn(cmd, myBundle{
				// cmd.Spawn(myBundle{
				pos: position{1, 2, 3},
				vel: velocity{1, 2, 3},
				acc: acceleration{1, 2, 3},
				rad: radius{1},
			})
		}
		cmd.Execute(world)
	}
}
