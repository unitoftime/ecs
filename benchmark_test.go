package ecs

import "testing"

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

// Note: Removed bc it uses internal code path: allocateMove -> causing it to potentially break hooks
// func BenchmarkAllocateQuery(b *testing.B) {
// 	world := NewWorld()

// 	query := Query4[position, velocity, acceleration, radius](world)

// 	b.ResetTimer()

// 	mask := buildArchMask(
// 		C(position{}),
// 		C(velocity{}),
// 		C(acceleration{}),
// 		C(radius{}),
// 	)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			world.allocateMove(id, mask)

// 			p, v, a, r := query.Read(id)
// 			*p = position{1, 2, 3}
// 			*v = velocity{4, 5, 6}
// 			*a = acceleration{7, 8, 9}
// 			*r = radius{10}
// 		}
// 	}
// }

// func BenchmarkAllocateQueryNoQuery(b *testing.B) {
// 	world := NewWorld()

// 	b.ResetTimer()

// 	mask := buildArchMask(
// 		C(position{}),
// 		C(velocity{}),
// 		C(acceleration{}),
// 		C(radius{}),
// 	)
// 	// archId := world.engine.getArchetypeId(mask)

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			newLoc := world.allocateMove(id, mask)

// 			// // Note: Slightly faster option. Actually, I'm not so sure
// 			// p := positionId.getPtr(world.engine, archId, index)
// 			// v := velocityId.getPtr(world.engine, archId, index)
// 			// a := accelerationId.getPtr(world.engine, archId, index)
// 			// r := radiusId.getPtr(world.engine, archId, index)
// 			// *p = position{1, 2, 3}
// 			// *v = velocity{4, 5, 6}
// 			// *a = acceleration{7, 8, 9}
// 			// *r = radius{10}

// 			// positionId.writeVal(world.engine, archId, index, position{1, 2, 3})
// 			// velocityId.writeVal(world.engine, archId, index, velocity{1, 2, 3})
// 			// accelerationId.writeVal(world.engine, archId, index, acceleration{1, 2, 3})
// 			// radiusId.writeVal(world.engine, archId, index, radius{10})

// 			wd := W{
// 				engine: world.engine,
// 				archId: newLoc.archId,
// 				index:  int(newLoc.index),
// 			}
// 			positionId.WriteVal(wd, position{1, 2, 3})
// 			velocityId.WriteVal(wd, velocity{1, 2, 3})
// 			accelerationId.WriteVal(wd, acceleration{1, 2, 3})
// 			radiusId.WriteVal(wd, radius{10})

// 		}
// 	}
// }

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

// func BenchmarkAllocateBundler2(b *testing.B) {
// 	world := NewWorld()

// 	bun := &Bundler{}

// 	b.ResetTimer()

// 	pos := C(position{1, 2, 3})
// 	vel := C(velocity{4, 5, 6})
// 	acc := C(acceleration{7, 8, 9})
// 	rad := C(radius{10})

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			bun.Clear()

// 			pos.Comp = position{1, 2, 3}

// 			bun.Add(&pos)
// 			bun.Add(&vel)
// 			bun.Add(&acc)
// 			bun.Add(&rad)

// 			id := world.NewId()
// 			bun.Write(world, id)
// 		}
// 	}
// }

// type outerBundle struct {
// 	myBundle myBundle
// }

// func (m outerBundle) Unbundle(bun *Bundler) {
// 	m.myBundle.Unbundle(bun)
// }

//--------------------------------------------------------------------------------

type myBundle struct {
	pos position
	vel velocity
	acc acceleration
	rad radius
}

// Note: This was interesting, but ended up being pretty slow
// func (m myBundle) BundleSeq() iter.Seq[Component] {
// 	return func(yield func(Component) bool) {
// 		if !yield(positionId.With(m.pos)) { return }
// 		if !yield(velocityId.With(m.vel)) { return }
// 		if !yield(accelerationId.With(m.acc)) { return }
// 		if !yield(radiusId.With(m.rad)) { return }
// 	}
// }

// func (m myBundle) Unbundle(bun *Bundler) {
// 	positionId.With(m.pos).Unbundle(bun)
// 	velocityId.With(m.vel).Unbundle(bun)
// 	accelerationId.With(m.acc).Unbundle(bun)
// 	radiusId.With(m.rad).Unbundle(bun)

// 	// positionId.UnbundleVal(bun, m.pos)
// 	// velocityId.UnbundleVal(bun, m.vel)
// 	// accelerationId.UnbundleVal(bun, m.acc)
// 	// radiusId.UnbundleVal(bun, m.rad)
// }

func (m myBundle) CompWrite(wd W) {
	m.pos.CompWrite(wd)
	m.vel.CompWrite(wd)
	m.acc.CompWrite(wd)
	m.rad.CompWrite(wd)
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
			unbundle(m, bun)
			id := world.NewId()
			bun.Write(world, id)
		}
	}
}

// func BenchmarkAllocateBundle4(b *testing.B) {
// 	world := NewWorld()

// 	var myBundle2 = NewBundle4[position, velocity, acceleration, radius]()

// 	bun := &Bundler{}

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			bun.Clear()

// 			myBundle2.Unbundle(bun,
// 				position{1, 2, 3},
// 				velocity{1, 2, 3},
// 				acceleration{1, 2, 3},
// 				radius{1},
// 			)

// 			id := world.NewId()
// 			bun.Write(world, id)
// 		}
// 	}
// }

// func BenchmarkAllocateBundle4Direct(b *testing.B) {
// 	world := NewWorld()

// 	var myBundle2 = NewBundle4[position, velocity, acceleration, radius]()

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			id := world.NewId()
// 			myBundle2.Write(world, id,
// 				position{1, 2, 3},
// 				velocity{1, 2, 3},
// 				acceleration{1, 2, 3},
// 				radius{1},
// 			)
// 		}
// 	}
// }

func BenchmarkAllocateNonBundle4Direct(b *testing.B) {
	world := NewWorld()

	b.ResetTimer()

	comps := []Component{
		position{1, 2, 3},
		velocity{1, 2, 3},
		acceleration{1, 2, 3},
		radius{1},
	}

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			id := world.NewId()
			world.Write(id, comps...)
		}
	}
}

// func BenchmarkAllocateCommands(b *testing.B) {
// 	world := NewWorld()
// 	cmd := NewCommandQueue(world)

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			cmd.Spawn(myBundle{
// 				pos: position{1, 2, 3},
// 				vel: velocity{1, 2, 3},
// 				acc: acceleration{1, 2, 3},
// 				rad: radius{1},
// 			})
// 		}
// 		cmd.Execute()
// 	}
// }

// func BenchmarkAllocateCommands2(b *testing.B) {
// 	world := NewWorld()
// 	cmd := NewCommandQueue(world)

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			CmdSpawn(cmd, myBundle{
// 				pos: position{1, 2, 3},
// 				vel: velocity{1, 2, 3},
// 				acc: acceleration{1, 2, 3},
// 				rad: radius{1},
// 			})
// 		}
// 		cmd.Execute()
// 	}
// }

func BenchmarkAllocateCommands3(b *testing.B) {
	world := NewWorld()
	cmd := NewCommandQueue(world)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			cmd.SpawnEmpty().Insert(
				myBundle{
					pos: position{1, 2, 3},
					vel: velocity{1, 2, 3},
					acc: acceleration{1, 2, 3},
					rad: radius{1},
				})
		}
		// 	myBundle{
		// 		pos: position{1, 2, 3},
		// 		vel: velocity{1, 2, 3},
		// 		acc: acceleration{1, 2, 3},
		// 		rad: radius{1},
		// 	}.Unbundle(entCmd.cmd.bundler)
		// }
		cmd.Execute()
	}
}

// func BenchmarkAllocateCommands4(b *testing.B) {
// 	world := NewWorld()
// 	cmd := NewCommandQueue()

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			cmd.SpawnEmpty().
// 				Insert(positionId.With(position{1, 2, 3})).
// 				Insert(velocityId.With(velocity{1, 2, 3})).
// 				Insert(accelerationId.With(acceleration{1, 2, 3})).
// 				Insert(radiusId.With(radius{1}))
// 		}
// 		cmd.Execute(world)
// 	}
// }

// func BenchmarkAllocateCommands5(b *testing.B) {
// 	world := NewWorld()
// 	cmd := NewCommandQueue(world)

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < addEntSize; i++ {
// 			cmd.SpawnEmpty().Add(
// 				myBundle{
// 					pos: position{1, 2, 3},
// 					vel: velocity{1, 2, 3},
// 					acc: acceleration{1, 2, 3},
// 					rad: radius{1},
// 				}.BundleSeq())
// 		}
// 		cmd.Execute()
// 	}
// }

func BenchmarkAllocateCommands6(b *testing.B) {
	world := NewWorld()
	cmd := NewCommandQueue(world)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < addEntSize; i++ {
			cmd.SpawnEmpty().
				Insert(position{1, 2, 3}).
				Insert(velocity{1, 2, 3}).
				Insert(acceleration{1, 2, 3}).
				Insert(radius{1})
		}
		cmd.Execute()
	}
}
