package main

import (
	"os"
	"fmt"
	"time"
	"math/rand"

	"github.com/unitoftime/ecs"
)

type Vec2 struct {
	X, Y float64
}
type Position Vec2
type Velocity Vec2
type Collider struct {
	Radius float64
	Count int
}

func main() {
	program := os.Args[1]
	size := 1000

	switch program {
	case "physics":
		benchPhysics(size)
	default:
		fmt.Printf("Invalid Program name %s\n", program)
		fmt.Println("Available Options")
		fmt.Println("physics - Runs a physics simulation")
	}
}

func benchPhysics(size int) {
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
		ecs.Map3(world, func(id ecs.Id, position *Position, velocity *Velocity, collider *Collider) {
			position.X += velocity.X * fixedTime
			position.Y += velocity.Y * fixedTime

			// Bump into the bounding cube
			if position.X <= 0 || position.X >= maxPosition {
				velocity.X = -velocity.X
			}
			if position.Y <= 0 || position.X >= maxPosition {
				velocity.Y = -velocity.Y
			}
		})

		// Check collisions, increment the count if a collision happens
		ecs.Map2(world, func(id ecs.Id, position *Position, collider *Collider) {
			ecs.Map2(world, func(targId ecs.Id, targPos *Position, targCollider *Collider) {
				if id == targId { return } // Skip if entity is the same
				dx := position.X - targPos.X
				dy := position.Y - targPos.Y
				distSq := (dx * dx) + (dy * dy)

				dr := collider.Radius + targCollider.Radius
				drSq := dr * dr

				if drSq > distSq {
					collider.Count++
				}
			})
		})

		dt = time.Since(start)
		fmt.Println(dt)
	}

	ecs.Map(world, func(id ecs.Id, collider *Collider) {
		fmt.Println(id, collider.Count)
	})
}
