package main

import (
	"fmt"
	"time"

	"github.com/unitoftime/ecs"
)

// This example illustrates the primary use cases for the ecs

type Name string

type Position struct {
	X, Y, Z float64
}

type Velocity struct {
	X, Y, Z float64
}

type PrintMessage struct {
	Msg string
}

var PrintMessageEventId = ecs.NewEvent[PrintMessage]()

func (p PrintMessage) EventId() ecs.EventId {
	return PrintMessageEventId
}

func main() {
	// Create a New World
	world := ecs.NewWorld()

	// You can manually spawn entities like this
	{
		cmd := ecs.NewCommandQueue(world)

		// Add entities
		cmd.SpawnEmpty().
			Insert(ecs.C(Position{1, 2, 3})).
			Insert(ecs.C(Velocity{1, 2, 3})).
			Insert(ecs.C(Name("My First Entity")))
		cmd.Execute()
	}

	// Adding Component hooks
	{
		world.SetHookOnAdd(ecs.C(Velocity{}),
			ecs.NewHandler(func(trigger ecs.Trigger[ecs.OnAdd]) {
				fmt.Println("Hook:", trigger)
			}))

		cmd := ecs.NewCommandQueue(world)
		cmd.SpawnEmpty().
			Insert(ecs.C(Position{1, 2, 3})).
			Insert(ecs.C(Velocity{1, 2, 3})).
			Insert(ecs.C(Name("My First Entity")))
		cmd.Execute()

	}

	// Adding Observers
	{
		cmd := ecs.NewCommandQueue(world)

		// You can add observer handlers which run as a result of triggerred events
		world.AddObserver(
			ecs.NewHandler(func(trigger ecs.Trigger[PrintMessage]) {
				fmt.Println("Observer 1:", trigger.Data.Msg)
			}))
		world.AddObserver(
			ecs.NewHandler(func(trigger ecs.Trigger[PrintMessage]) {
				fmt.Println("Observer 2!", trigger.Data.Msg)
			}))

		cmd.Trigger(PrintMessage{"Hello"})
		cmd.Trigger(PrintMessage{"World"})

		cmd.Execute()
	}

	scheduler := ecs.NewScheduler(world)

	// Append physics systems, these run on a fixed time step, so dt will always be constant
	scheduler.AddSystems(ecs.StageFixedUpdate,
		// Comment out if you want to spawn a new entity every frame
		// ecs.NewSystem1(SpawnSystem),

		// Option A: Create a function that returns a system
		MoveSystemOption_A(world),

		// Option B: Use the dynamic injection to create a system for you
		ecs.NewSystem1(MoveSystemOption_B),

		ecs.NewSystem1(PrintSystem),

		ecs.NewSystem1(TriggerSystem),
	)

	// Also, add render systems if you want, These run as fast as possible
	// scheduler.AppendRender()

	// This will block until the scheduler exits `scheudler.SetQuit(true)`
	scheduler.Run()
}

// Note: This system wasn't added to the scheduler, so that I wouldn't constantly spawn entities in the physics loop
// But, you can rely on commands to get injected for you, just like a query.
func SpawnSystem(dt time.Duration, commands *ecs.CommandQueue) {
	// Note: The scheduler will automatically call .Execute() the command queue
	cmd := commands.SpawnEmpty()

	name := Name(fmt.Sprintf("My Entity %d", cmd.Id()))
	cmd.Insert(ecs.C(Position{1, 2, 3})).
		Insert(ecs.C(Velocity{1, 2, 3})).
		Insert(ecs.C(name))
}

// Option A: Define and return a system based on a closure
// - Provides a bit more flexibility if you need to establish variables ahead of the system
func MoveSystemOption_A(world *ecs.World) ecs.System {
	query := ecs.Query2[Position, Velocity](world)

	return ecs.NewSystem(func(dt time.Duration) {
		query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
			sec := dt.Seconds()

			pos.X += vel.X * sec
			pos.Y += vel.Y * sec
			pos.Z += vel.Z * sec
		})
	})
}

// Option 2: Define a system and have all the queries created and injected for you
// - Can be used for simpler systems that dont need to track much system-internal state
// - Use the `ecs.NewSystemN(world, systemFunction)` syntax (Where N represents the number of required resources)
func MoveSystemOption_B(dt time.Duration, query *ecs.View2[Position, Velocity]) {
	query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
		sec := dt.Seconds()

		pos.X += vel.X * sec
		pos.Y += vel.Y * sec
		pos.Z += vel.Z * sec
	})
}

// A system that prints all entity names and their positions
func PrintSystem(dt time.Duration, query *ecs.View2[Name, Position]) {
	query.MapId(func(id ecs.Id, name *Name, pos *Position) {
		fmt.Printf("%s: %v\n", *name, pos)
	})
}

func TriggerSystem(dt time.Duration, cmd *ecs.CommandQueue) {
	cmd.Trigger(PrintMessage{"Hello"})
}
