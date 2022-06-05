package main

import (
	"fmt"
	"log"
	"time"
	"math/rand"

	"runtime"
	"runtime/pprof"
	"flag"
	"os"

	"github.com/unitoftime/ecs"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

type Vec2 struct {
	X, Y float64
}
type Position Vec2
type Velocity Vec2
type Collider struct {
	Radius float64
	Count int32
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		go func() {
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
		}()
		defer pprof.StopCPUProfile()
	}

	program := os.Args[1]
	size := 10000

	switch program {
	case "physics":
		benchPhysics(size, 0)
	case "physicsOpt":
		benchPhysicsOptimized2(size, 0)
	case "physicsDelete":
		benchPhysicsOptimized2(size, 100)
	default:
		fmt.Printf("Invalid Program name %s\n", program)
		fmt.Println("Available Options")
		fmt.Println("physics - Runs a physics simulation")
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func moveCircles(query *ecs.Query2[Position, Velocity], fixedTime float64, maxPosition float64, loopCounter *int) {
	query.Map(func(ids []ecs.Id, pos []Position, vel []Velocity) {
		if len(ids) != len(pos) || len(ids) != len(vel) { panic("ERR") }
		for i := range ids {
			pos[i].X += vel[i].X * fixedTime
			pos[i].Y += vel[i].Y * fixedTime

			// Bump into the bounding rect
			if pos[i].X <= 0 || pos[i].X >= maxPosition {
				vel[i].X = -vel[i].X
			}
			if pos[i].Y <= 0 || pos[i].Y >= maxPosition {
				vel[i].Y = -vel[i].Y
			}
			*loopCounter++
		}
	})
}

func checkCollisions(query *ecs.Query2[Position, Collider], collisionLimit int32, deathCount *int, loopCounter *int) {

	// Alternative?
	// archetypes.Map2D(func(
	// 	aId []ecs.Id, aPos []Position, aCol []Collider,
	// 	bId []ecs.Id, bPos []Position, bCol []Collider) {

	query.Map(func(aId []ecs.Id, aPos []Position, aCol []Collider) {
		query.Map(func(bId []ecs.Id, bPos []Position, bCol []Collider) {
			if len(aId) != len(aPos) || len(aId) != len(aCol) { panic("ERR") }
			if len(bId) != len(bPos) || len(bId) != len(bCol) { panic("ERR") }
			for i := range aId {
				for j := range bId {
					if aId[i] == bId[j] { continue } // Skip if entity is the same

					dx := aPos[i].X - bPos[j].X
					dy := aPos[i].Y - bPos[j].Y
					distSq := (dx * dx) + (dy * dy)

					dr := aCol[i].Radius + bCol[j].Radius
					drSq := dr * dr

					if drSq > distSq {
						aCol[i].Count++
					}

					// Kill and spawn one
					// TODO move to outer loop?
					if collisionLimit > 0 && aCol[i].Count > collisionLimit {
						*deathCount++
						// success := ecs.Delete(world, aId[i])
						// if success {
						// 	deathCount++
						// 	break
						// }
					}

					*loopCounter++
				}
			}
		})
	})
}


func benchPhysicsOptimized2(size int, collisionLimit int32) {
	iterations := 1000

	world := ecs.NewWorld()
	maxSpeed := 10.0
	maxPosition := 100.0
	maxCollider := 1.0

	for i := 0; i < size; i++ {
		id := world.NewId()
		ent := ecs.NewEntity(
			ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
			ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
			ecs.C(Collider{
				Radius: maxCollider * rand.Float64(),
				Count: 0,
			}),
		)
		ecs.WriteEntity(world, id, ent)
	}

	loopCounter := 0
	fixedTime := (15 * time.Millisecond).Seconds()

	start := time.Now()
	dt := time.Since(start)
	for iterCount := 0; iterCount < iterations; iterCount++ {
		start = time.Now()

		ecs.ExecuteSystem2(world, func(query *ecs.Query2[Position, Velocity]) {
			moveCircles(query, fixedTime, maxPosition, &loopCounter)
		})

		deathCount := 0
		ecs.ExecuteSystem2(world, func(query *ecs.Query2[Position, Collider]) {
			checkCollisions(query, collisionLimit, &deathCount, &loopCounter)
		})


		// ExecuteSystem2(world, func(archetypes Archetypes[Position, Velocity]) {
		// 	moveCircles(archetypes, fixedTime, maxPosition, &loopCounter)
		// })

		// deathCount := 0
		// ExecuteSystem2(world, func(arch Archetypes[Position, Collider]) {
		// 	checkCollisions(arch, collisionLimit, &deathCount, &loopCounter)
		// })

		// Spawn new entities, one per each entity we deleted
		for i := 0; i < deathCount; i++ {
			id := world.NewId()
			ent := ecs.NewEntity(
				ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
				ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
				ecs.C(Collider{
					Radius: maxCollider * rand.Float64(),
					Count: 0,
				}),
			)
			ecs.WriteEntity(world, id, ent)
		}

		// world.Print(0)

		dt = time.Since(start)
		fmt.Println(iterCount, dt, loopCounter)
		loopCounter = 0
	}

	ecs.Map(world, func(id ecs.Id, collider *Collider) {
		fmt.Println(id, collider.Count)
	})
}

	/*
974 1031
975 625
976 787
977 208
978 1601
979 1243
980 167
981 108
982 1040
983 500
984 637
985 1011
986 830
987 1247
988 901
989 1597
990 418
991 767
992 951
993 1252
994 948
995 194
996 290
997 181
998 1276
999 858
1000 789
1001 638
*/


func benchPhysics(size int, collisionLimit int32) {
	iterations := 1000

	loopCounter := 0
	world := ecs.NewWorld()
	maxSpeed := 10.0
	maxPosition := 100.0
	maxCollider := 1.0

	for i := 0; i < size; i++ {
		id := world.NewId()
		ent := ecs.NewEntity(
			ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
			ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
			ecs.C(Collider{
				Radius: maxCollider * rand.Float64(),
				Count: 0,
			}),
		)
		ecs.WriteEntity(world, id, ent)
	}

	start := time.Now()
	dt := time.Since(start)
	fixedTime := (15 * time.Millisecond).Seconds()
	for i := 0; i < iterations; i++ {
		start = time.Now()

		// Update positions
		ecs.Map2(world, func(id ecs.Id, position *Position, velocity *Velocity) {
			position.X += velocity.X * fixedTime
			position.Y += velocity.Y * fixedTime

			// Bump into the bounding rect
			if position.X <= 0 || position.X >= maxPosition {
				velocity.X = -velocity.X
			}
			if position.Y <= 0 || position.Y >= maxPosition {
				velocity.Y = -velocity.Y
			}
			loopCounter++
		})

		// Check collisions, increment the count if a collision happens
		deathCount := 0
		ecs.Map2(world, func(aId ecs.Id, aPos *Position, aCol *Collider) {
			ecs.Map2(world, func(bId ecs.Id, bPos *Position, bCol *Collider) {
				if aId == bId { return } // Skip if entity is the same

				dx := aPos.X - bPos.X
				dy := aPos.Y - bPos.Y
				distSq := (dx * dx) + (dy * dy)

				dr := aCol.Radius + bCol.Radius
				drSq := dr * dr

				if drSq > distSq {
					aCol.Count++
				}

				// Kill and spawn one
				// TODO move to outer loop?
				if collisionLimit > 0 && aCol.Count > collisionLimit {
					success := ecs.Delete(world, aId)
					if success {
						deathCount++
						return
					}
				}
				loopCounter++
			})
		})

		// Spawn new entities, one per each entity we deleted
		for i := 0; i < deathCount; i++ {
			id := world.NewId()
			ent := ecs.NewEntity(
				ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
				ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
				ecs.C(Collider{
					Radius: maxCollider * rand.Float64(),
					Count: 0,
				}),
			)
			ecs.WriteEntity(world, id, ent)
		}

		// world.Print(0)

		dt = time.Since(start)
		fmt.Println(i, dt, deathCount, loopCounter)
		loopCounter = 0
	}

	ecs.Map(world, func(id ecs.Id, collider *Collider) {
		fmt.Println(id, collider.Count)
	})
}

/*
func benchPhysicsOptimized(size int, collisionLimit int32) {
	iterations := 1000

	world := ecs.NewWorld()
	maxSpeed := 10.0
	maxPosition := 100.0
	maxCollider := 1.0

	for i := 0; i < size; i++ {
		id := world.NewId()
		ent := ecs.NewEntity(
			ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
			ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
			ecs.C(Collider{
				Radius: maxCollider * rand.Float64(),
				Count: 0,
			}),
		)
		ecs.WriteEntity(world, id, ent)
	}

	loopCounter := 0
	fixedTime := (15 * time.Millisecond).Seconds()

	start := time.Now()
	dt := time.Since(start)
	for iterCount := 0; iterCount < iterations; iterCount++ {
		start = time.Now()

		{
			// view := ecs.ViewAll2[Position, Velocity](world)
			// for iter := view.Iterate(); iter.Ok(); {
			// 	_, pos, vel := iter.Next()
			// 	// fmt.Println("0", iter)
			// 	pos.X += vel.X * fixedTime
			// 	pos.Y += vel.Y * fixedTime

			// 	// Bump into the bounding rect
			// 	if pos.X <= 0 || pos.X >= maxPosition {
			// 		vel.X = -vel.X
			// 	}
			// 	if pos.Y <= 0 || pos.Y >= maxPosition {
			// 		vel.Y = -vel.Y
			// 	}
			// 	loopCounter++
			// }


			view := ecs.ViewAll2[Position, Velocity](world)
			for view.Ok() {
				_, pos, vel := view.IterChunkClean()
				if len(pos) != len(vel) { panic("ERR") }
				for j := range pos {
					pos[j].X += vel[j].X * fixedTime
					pos[j].Y += vel[j].Y * fixedTime

					// Bump into the bounding rect
					if pos[j].X <= 0 || pos[j].X >= maxPosition {
						vel[j].X = -vel[j].X
					}
					if pos[j].Y <= 0 || pos[j].Y >= maxPosition {
						vel[j].Y = -vel[j].Y
					}
					loopCounter++
				}
			}
		}


		// deathCount := 0
		// view := ecs.ViewAll2[Position, Collider](world)
		// // view2 := ecs.ViewAll2[Position, Collider](world)
		// for iter := view.Iterate(); iter.Ok(); {
		// 	aId, aPos, aCol := iter.Next()
		// 	// fmt.Println("1", iter, aId, aPos, aCol)
		// // for view.Ok() {
		// // 	aId, aPos, aCol := view.Iter4()

		// 	for iter2 := view.Iterate(); iter2.Ok(); {
		// 		bId, bPos, bCol := iter2.Next()
		// 		// fmt.Println("2", iter2, bId, bPos, bCol)


		// 	// view2.Reset()
		// 	// for view2.Ok() {
		// 	// 	bId, bPos, bCol := view2.Iter4()

		// 		if aId == bId { continue } // Skip if entity is the same

		// 		dx := aPos.X - bPos.X
		// 		dy := aPos.Y - bPos.Y
		// 		distSq := (dx * dx) + (dy * dy)

		// 		dr := aCol.Radius + bCol.Radius
		// 		drSq := dr * dr

		// 		if drSq > distSq {
		// 			aCol.Count++
		// 		}

		// 		// Kill and spawn one
		// 		// TODO move to outer loop?
		// 		if collisionLimit > 0 && aCol.Count > collisionLimit {
		// 			success := ecs.Delete(world, aId)
		// 			if success {
		// 				deathCount++
		//        break
		// 			}
		// 		}

		// 		loopCounter++
		// 	}
		// }


		// !!!Fastest!!!!
		// Check collisions, increment the count if a collision happens
		deathCount := 0
		view := ecs.ViewAll2[Position, Collider](world)
		view2 := ecs.ViewAll2[Position, Collider](world)
		for view.Ok() {
			ids, pos, col := view.IterChunkClean()

			if len(ids) != len(pos) || len(ids) != len(col) { panic ("ERROR") }
			for j := range ids {
				aId := ids[j]
				aPos := &pos[j]
				aCol := &col[j]

				view2.Reset()
				for view2.Ok() {
					targIdList, targPosList, targCol := view2.IterChunkClean()

					if len(targIdList) != len(targPosList) || len(targIdList) != len(targCol) { panic ("ERROR") }
					for jj := range targIdList {
						bId := targIdList[jj]
						bPos := &targPosList[jj]
						bCol := &targCol[jj]

						if aId == bId { continue } // Skip if entity is the same

						dx := aPos.X - bPos.X
						dy := aPos.Y - bPos.Y
						distSq := (dx * dx) + (dy * dy)

						dr := aCol.Radius + bCol.Radius
						drSq := dr * dr

						if drSq > distSq {
							aCol.Count++
						}

						// Kill and spawn one
						// TODO move to outer loop?
						if collisionLimit > 0 && aCol.Count > collisionLimit {
							success := ecs.Delete(world, aId)
							if success {
								deathCount++
								break
							}
						}

						loopCounter++
					}
				}
			}
		}

		// Spawn new entities, one per each entity we deleted
		for i := 0; i < deathCount; i++ {
			id := world.NewId()
			ent := ecs.NewEntity(
				ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
				ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
				ecs.C(Collider{
					Radius: maxCollider * rand.Float64(),
					Count: 0,
				}),
			)
			ecs.WriteEntity(world, id, ent)
		}

		// world.Print(0)

		dt = time.Since(start)
		fmt.Println(iterCount, dt, loopCounter)
		loopCounter = 0
	}

	ecs.Map(world, func(id ecs.Id, collider *Collider) {
		fmt.Println(id, collider.Count)
	})
}
*/
