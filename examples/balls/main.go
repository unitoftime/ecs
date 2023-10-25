package main

import (
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/unitoftime/ecs"
	ballESC "github.com/unitoftime/ecs/examples/balls/ecs"
	"github.com/unitoftime/ecs/system/group"
)

func main() {
	rl.InitWindow(500, 500, "ECS balls example")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	world := ecs.NewWorld()

	for i := 0; i < 200; i++ {
		posX := rand.Float64() * 50
		posY := rand.Float64() * 50

		id := world.NewId()
		ecs.Write(
			world,
			id,
			ecs.C(ballESC.Position{X: posX, Y: posY}),
			ecs.C(ballESC.Velocity{X: rand.Float64()/2 - 0.5, Y: rand.Float64()/2 - 0.5}),
			ecs.C(ballESC.RenderPosition{X: posX, Y: posY}),
		)
	}

	componentsGuard := group.NewComponentsGuard()

	// Init physics
	physics := group.NewFixedSystemGroup("physics", 50*time.Millisecond, componentsGuard)
	physics.AddSystem(&ballESC.MovementSystem{World: world})
	sharedPhysicsSpacialHash := make(map[ballESC.IntVector2][]ballESC.SpacialCellObject)
	physics.AddSystem(&ballESC.SpacialSystem{
		World:       world,
		SpacialHash: sharedPhysicsSpacialHash,
	})
	physics.AddSystem(&ballESC.CollisionSystem{
		World:       world,
		SpacialHash: sharedPhysicsSpacialHash,
	})
	physics.AddSystem(&ballESC.BoundarySystem{
		World: world,
	})
	physics.Build()
	physics.StartFixed()
	defer physics.StopFixed()

	// Init render
	render := group.NewRealtimeGroup("render", componentsGuard)
	render.OnBeforeUpdate(func(delta time.Duration) {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
	})
	render.OnAfterUpdate(func(delta time.Duration) {
		rl.DrawText("Collision!", 180, 225, 40, rl.Black)
		rl.EndDrawing()
	})
	render.AddSystem(&ballESC.UpdateRenderPositionSystem{World: world})
	render.AddSystem(&ballESC.RenderSystem{World: world})
	render.Build()

	for !rl.WindowShouldClose() {
		render.RunRealtime(time.Duration(int64(float64(rl.GetFrameTime()) * float64(time.Second))))
	}
}
