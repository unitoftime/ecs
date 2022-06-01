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

	// program := os.Args[1]
	program := "physics"
	size := 10000

	switch program {
	case "physics":
		benchPhysics(size, 0)
	case "physicsDelete":
		benchPhysics(size, 100)
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

func benchPhysics(size int, collisionLimit int32) {
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
		ecs.Map2(world, func(id ecs.Id, position *Position, collider *Collider) {
			ecs.SmartMap2(world, func(targId ecs.Id, targPos *Position, targCollider *Collider) bool {
				if id == targId { return true } // Skip if entity is the same
				dx := position.X - targPos.X
				dy := position.Y - targPos.Y
				distSq := (dx * dx) + (dy * dy)

				dr := collider.Radius + targCollider.Radius
				drSq := dr * dr

				if drSq > distSq {
					collider.Count++
				}

				// Kill and spawn one
				// TODO move to outer loop?
				if collisionLimit > 0 && collider.Count > collisionLimit {
					success := ecs.Delete(world, id)
					if success {
						deathCount++
						return false
					}
				}
				return true
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
		fmt.Println(dt)
	}

	// ecs.Map(world, func(id ecs.Id, collider *Collider) {
	// 	fmt.Println(id, collider.Count)
	// })
}
