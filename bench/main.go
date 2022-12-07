package main

// TODO - Add ballast and investigate GC pressure?
// TODO - Disable GC: GOCG=-1 go run .
// TODO - manual runtime.GC()

import (
	"fmt"
	"log"
	"time"
	"math/rand"
	"strconv"

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
}
type Count struct {
	Count int32
}

const iterations = 50
const maxPosition = 100.0
const maxSpeed = 10.0
const maxCollider = 1.0

func main() {
	// Create a large heap allocation of 10 GiB

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
	size, err := strconv.Atoi(os.Args[2])
	if err != nil { panic(err) }
	colLimitArg, err := strconv.Atoi(os.Args[3])
	if err != nil { panic(err) }
	collisionLimit := int32(colLimitArg)


	// ballast := make([]byte, 10<<30)

	fmt.Println("Iter", "Time")
	switch program {
	case "native":
		benchNativeComponents(size, collisionLimit)
	case "nativeSplit":
		benchNativeSplit(size, collisionLimit)
	case "ecs-slow":
		benchPhysics(size, collisionLimit)
	case "ecs":
		benchPhysicsOptimized(size, collisionLimit)
	default:
		fmt.Printf("Invalid Program name %s\n", program)
		fmt.Println("Available Options")
		fmt.Println("physics - Runs a physics simulation")
	}

	// fmt.Println(len(ballast))

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

func createWorld(size int) *ecs.World {
	world := ecs.NewWorld()

	for i := 0; i < size; i++ {
		id := world.NewId()
		ent := ecs.NewEntity(
			ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
			ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
			ecs.C(Collider{
				Radius: maxCollider * rand.Float64(),
			}),
			ecs.C(Count{}),
		)
		ecs.WriteEntity(world, id, ent)
	}
	return world
}

func moveCircles(query *ecs.Query2[Position, Velocity], fixedTime float64, maxPosition float64) {
	query.Map(func(ids []ecs.Id, pos []Position, vel []Velocity) {
		if len(ids) != len(pos) || len(ids) != len(vel) { panic("ERR") }
		for i := range ids {
			if ids[i] == ecs.InvalidEntity { continue }
			pos[i].X += vel[i].X * fixedTime
			pos[i].Y += vel[i].Y * fixedTime

			// Bump into the bounding rect
			if pos[i].X <= 0 || pos[i].X >= maxPosition {
				vel[i].X = -vel[i].X
			}
			if pos[i].Y <= 0 || pos[i].Y >= maxPosition {
				vel[i].Y = -vel[i].Y
			}
		}
	})
}

func checkCollisions(world *ecs.World, query *ecs.Query3[Position, Collider, Count], innerQuery *ecs.Query2[Position, Collider], collisionLimit int32, deathCount *int) {

	// Alternative?
	// archetypes.Map2D(func(
	// 	aId []ecs.Id, aPos []Position, aCol []Collider,
	// 	bId []ecs.Id, bPos []Position, bCol []Collider) {

	query.Map(func(aId []ecs.Id, aPos []Position, aCol []Collider, aCnt []Count) {
		innerQuery.Map(func(bId []ecs.Id, bPos []Position, bCol []Collider) {
			if len(aId) != len(aPos) || len(aId) != len(aCol) { panic("ERR") }
			if len(bId) != len(bPos) || len(bId) != len(bCol) { panic("ERR") }
			for i := range aId {
				if aId[i] == ecs.InvalidEntity { continue }
				aPos_i := &aPos[i]
				aCol_i := &aCol[i]
				for j := range bId {
					if bId[i] == ecs.InvalidEntity { continue }
					bPos_j := &bPos[j]
					bCol_j := &bCol[j]
					if aId[i] == bId[j] { continue } // Skip if entity is the same

					dx := aPos_i.X - bPos_j.X
					dy := aPos_i.Y - bPos_j.Y
					distSq := (dx * dx) + (dy * dy)

					dr := aCol_i.Radius + bCol_j.Radius
					drSq := dr * dr

					if drSq > distSq {
						aCnt[i].Count++
					}

					// Kill and spawn one
					// TODO move to outer loop?
					if collisionLimit > 0 && aCnt[i].Count > collisionLimit {
						success := ecs.Delete(world, aId[i])
						if success {
							*deathCount++
							break
						}
					}
				}
			}
		})
	})
}


func benchPhysicsOptimized(size int, collisionLimit int32) {
	world := createWorld(size)

	fixedTime := (15 * time.Millisecond).Seconds()

	start := time.Now()
	dt := time.Since(start)
	for iterCount := 0; iterCount < iterations; iterCount++ {
		start = time.Now()

		// ecs.ExecuteSystem2(world, func(query *ecs.Query2[Position, Velocity]) {
		// 	moveCircles(query, fixedTime, maxPosition)
		// })

		// deathCount := 0
		// ecs.ExecuteSystem2(world, func(query *ecs.Query2[Position, Collider]) {
		// 	checkCollisions(world, query, collisionLimit, &deathCount)
		// })

		{
			query := ecs.NewQuery2[Position, Velocity](world)
			moveCircles(query, fixedTime, maxPosition)
		}

		deathCount := 0
		{
			query := ecs.NewQuery3[Position, Collider, Count](world)
			innerQuery := ecs.NewQuery2[Position, Collider](world)
			checkCollisions(world, query, innerQuery, collisionLimit, &deathCount)
		}

		// fmt.Println("DeathCount:", deathCount)

		// Spawn new entities, one per each entity we deleted
		for i := 0; i < deathCount; i++ {
			id := world.NewId()
			ent := ecs.NewEntity(
				ecs.C(Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}),
				ecs.C(Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}),
				ecs.C(Collider{
					Radius: maxCollider * rand.Float64(),
				}),
				ecs.C(Count{}),
			)
			ecs.WriteEntity(world, id, ent)
		}

		// world.Print(0)

		dt = time.Since(start)
		fmt.Println(iterCount, dt.Seconds())
	}

	// ecs.Map(world, func(id ecs.Id, collider *Collider) {
	// 	fmt.Println(id, collider.Count)
	// })
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
	world := createWorld(size)

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
		})

		// Check collisions, increment the count if a collision happens
		deathCount := 0
		ecs.Map3(world, func(aId ecs.Id, aPos *Position, aCol *Collider, aCnt *Count) {
			ecs.Map2(world, func(bId ecs.Id, bPos *Position, bCol *Collider) {
				if aId == bId { return } // Skip if entity is the same

				dx := aPos.X - bPos.X
				dy := aPos.Y - bPos.Y
				distSq := (dx * dx) + (dy * dy)

				dr := aCol.Radius + bCol.Radius
				drSq := dr * dr

				if drSq > distSq {
					aCnt.Count++
				}

				// Kill and spawn one
				// TODO move to outer loop?
				if collisionLimit > 0 && aCnt.Count > collisionLimit {
					success := ecs.Delete(world, aId)
					if success {
						deathCount++
						return
					}
				}
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
				}),
				ecs.C(Count{}),
			)
			ecs.WriteEntity(world, id, ent)
		}

		// world.Print(0)

		dt = time.Since(start)
		fmt.Println(i, dt.Seconds())
	}

	// ecs.Map(world, func(id ecs.Id, collider *Collider) {
	// 	fmt.Println(id, collider.Count)
	// })
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

func benchNativeComponents(size int, collisionLimit int32) {
	// [uint64]
	// [{float64, float64}]
	// [{float64, float64}]
	// [{float64, int32}]

	// [uint64, uint64]
	// [{float64, float64}, {float64, float64}]
	// [{float64, float64}, {float64, float64}]
	// [{float64, int32}, {float64, int32}]
	ids := make([]ecs.Id, size, size)
	pos := make([]Position, size, size)
	vel := make([]Velocity, size, size)
	col := make([]Collider, size, size)
	cnt := make([]Count, size, size)

	for i := 0; i < size; i++ {
		ids[i] = ecs.Id(i+2)
		pos[i] = Position{maxPosition * rand.Float64(), maxPosition * rand.Float64()}
		vel[i] = Velocity{maxSpeed * rand.Float64(), maxSpeed * rand.Float64()}
		col[i] = Collider{
			Radius: maxCollider * rand.Float64(),
		}
		cnt[i] = Count{}
	}

	start := time.Now()
	dt := time.Since(start)
	fixedTime := (15 * time.Millisecond).Seconds()
	for iterCount := 0; iterCount < iterations; iterCount++ {
		start = time.Now()

		// Update positions
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
		}

		// Check collisions, increment the count if a collision happens
		deathCount := 0
		for i := range ids {
			aId := ids[i]
			aPos := &pos[i]
			aCol := &col[i]
			aCnt := &cnt[i]
			for j := range ids {
				bId := ids[j]
				bPos := &pos[j]
				bCol := &col[j]

				if aId == bId { continue } // Skip if entity is the same

				dx := aPos.X - bPos.X
				dy := aPos.Y - bPos.Y
				distSq := (dx * dx) + (dy * dy)

				dr := aCol.Radius + bCol.Radius
				drSq := dr * dr

				if drSq > distSq {
					aCnt.Count++
				}

				// Kill and spawn one
				// TODO move to outer loop?
				if collisionLimit > 0 && aCnt.Count > collisionLimit {
					deathCount++
				}
			}
		}

		dt = time.Since(start)
		fmt.Println(iterCount, dt.Seconds())
	}

	// for i := range ids {
	// 	fmt.Println(ids[i], col[i].Count)
	// }
}

// struct myStruct {
//        X float64
//        Y float64
// }

// myarray []myStruct

// myArrayX []float64
// myArrayY []float64

// [uint64]
// [{float64, float64}]
// [{float64, float64}]
// [{float64, int32}]

// Holes     [bool]    [true]    ...
// Id        [uint64]  [uint64]  ...
// PosX      [float64] [float64] ...
// PosY      [float64] [float64] ...
// VelX      [float64] [float64] ...
// VelY      [float64] [float64] ...
// ColRadius [float64] [float64] ...
// ColCount  [int32]   [int32]   ...


// Test with this new memory layout
// [uint64]
// PosX [float64]
// PosY [float64]
// VelX [float64]
// VelY [float64]
// ColRadius [float64]
// ColCount [int32]
func benchNativeSplit(size int, collisionLimit int32) {
	ids := make([]ecs.Id, size, size)
	posX := make([]float64, size, size)
	posY := make([]float64, size, size)
	velX := make([]float64, size, size)
	velY := make([]float64, size, size)
	col := make([]float64, size, size)
	cnt := make([]int32, size, size)

	for i := 0; i < size; i++ {
		ids[i] = ecs.Id(i+2)
		posX[i] = maxPosition * rand.Float64()
		posY[i] = maxPosition * rand.Float64()
		velX[i] = maxSpeed * rand.Float64()
		velY[i] = maxSpeed * rand.Float64()
		col[i] = maxCollider * rand.Float64()
		cnt[i] = 0
	}

	fixedTime := 0.015
	for iterCount := 0; iterCount < iterations; iterCount++ {
		start := time.Now()
		// Update positions
		for i := range ids {
			posX[i] += velX[i] * fixedTime
			posY[i] += velY[i] * fixedTime

			// Bump into the bounding rect
			if posX[i] <= 0 || posX[i] >= maxPosition {
				velX[i] = -velX[i]
			}
			if posY[i] <= 0 || posY[i] >= maxPosition {
				velY[i] = -velY[i]
			}
		}

		// Check collisions, increment the count if a collision happens
		deathCount := 0
		for i := range ids {
			aId := ids[i]
			aPosX := &posX[i]
			aPosY := &posY[i]
			aCol := &col[i]
			for j := range ids {
				bId := ids[j]
				bPosX := &posX[j]
				bPosY := &posY[j]
				bCol := &col[j]

				if aId == bId { continue } // Skip if entity is the same

				dx := *aPosX - *bPosX
				dy := *aPosY - *bPosY
				distSq := (dx * dx) + (dy * dy)

				dr := *aCol + *bCol
				drSq := dr * dr

				if drSq > distSq {
					cnt[i]++
				}

				if collisionLimit > 0 && cnt[i] > collisionLimit {
					deathCount++
				}
			}
		}

		dt := time.Since(start)
		fmt.Println(iterCount, dt.Seconds())
	}

	// for i := range ids {
	// 	fmt.Println(ids[i], cnt[i])
	// }
}
// [ id,  id ,  id ]
// [ pos, pos, pos ]
// [ vel, vel,     ]
// [ col, col, col ]
