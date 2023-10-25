package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/unitoftime/ecs"
	ballESC "github.com/unitoftime/ecs/examples/balls/ecs"
	"github.com/unitoftime/ecs/system/group"
)

func main() {
	rl.InitWindow(800, 450, "ECS balls example")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	world := ecs.NewWorld()
	id := world.NewId()
	ecs.Write(
		world,
		id,
		ecs.C(ballESC.Position{X: 0, Y: 0}),
		ecs.C(ballESC.Velocity{X: 0, Y: 0}),
		ecs.C(ballESC.RenderPosition{X: 0, Y: 0}),
	)

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
	physics.Build()
	physics.StartFixed()
	defer physics.StopFixed()

	// Init render
	render := group.NewRealtimeGroup("render", componentsGuard)
	render.OnBeforeUpdate(func(delta time.Duration) {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.DrawText("Collision!", 190, 200, 20, rl.LightGray)
	})
	render.OnAfterUpdate(func(delta time.Duration) {
		rl.EndDrawing()
	})
	render.AddSystem(&ballESC.RenderSystem{World: world})
	render.Build()

	for !rl.WindowShouldClose() {
		render.RunRealtime(time.Duration(int64(float64(rl.GetFrameTime()) * float64(time.Second))))
	}
}
