package ecs

import (
	"fmt"
	"testing"
	"time"
)

// // Note: You can do this to pack more queries in
// type custom struct {
// 	query1 *View2[position, velocity]
// 	query2 *View2[position, velocity]
// }

// func (c *custom) initialize(world *World) any {
// 	return &custom{
// 		GetInjectable[*View2[position, velocity]](world),
// 		GetInjectable[*View2[position, velocity]](world),
// 	}
// }

func physicsSystem(dt time.Duration, query *View2[position, velocity]) {
	query.MapId(func(id Id, pos *position, vel *velocity) {
		pos.x += vel.x * dt.Seconds()
		pos.y += vel.y * dt.Seconds()
		pos.z += vel.z * dt.Seconds()
	})
}

func TestSystemCreationNew(t *testing.T) {
	world := setupPhysics(100)
	sys := NewSystem1(physicsSystem).Build(world)

	fmt.Println("NAME", sys.Name)
	for range 100 {
		sys.step(16 * time.Millisecond)
	}
}

var lastTime time.Time

func TestSchedulerPhysics(t *testing.T) {
	world := NewWorld()
	scheduler := NewScheduler(world)
	scheduler.AppendPhysics(System{
		Name: "TestSystem",
		Func: func(dt time.Duration) {
			fmt.Printf("%v - %v\n", dt, time.Since(lastTime))
			lastTime = time.Now()
		},
	})
	lastTime = time.Now()
	go scheduler.Run()
	time.Sleep(1 * time.Second)
	scheduler.SetQuit(true)
}

var lastTimeInput, lastTimePhysics, lastTimeRender time.Time

func TestSchedulerAll(t *testing.T) {
	world := NewWorld()
	scheduler := NewScheduler(world)
	scheduler.AppendInput(System{
		Name: "TestSystemInput",
		Func: func(dt time.Duration) {
			fmt.Printf("Input:   %v - %v\n", dt, time.Since(lastTimeInput))
			lastTimeInput = time.Now()
			time.Sleep(1 * time.Millisecond)
		},
	})
	scheduler.AppendPhysics(System{
		Name: "TestSystemPhysics",
		Func: func(dt time.Duration) {
			fmt.Printf("Physics: %v - %v\n", dt, time.Since(lastTimePhysics))
			lastTimePhysics = time.Now()
		},
	})
	scheduler.AppendRender(System{
		Name: "TestSystemRender",
		Func: func(dt time.Duration) {
			fmt.Printf("Render:  %v - %v\n", dt, time.Since(lastTimeRender))
			lastTimeRender = time.Now()
			time.Sleep(100 * time.Millisecond)
		},
	})
	lastTimeInput = time.Now()
	lastTimePhysics = time.Now()
	lastTimeRender = time.Now()
	go scheduler.Run()
	time.Sleep(1 * time.Second)
	scheduler.SetQuit(true)
}
