[![Go Reference](https://pkg.go.dev/badge/github.com/unitoftime/ecs.svg)](https://pkg.go.dev/github.com/unitoftime/ecs)
[![Build](https://github.com/unitoftime/ecs/actions/workflows/build.yml/badge.svg)](https://github.com/unitoftime/ecs/actions/workflows/build.yml)

This is an ecs library I wrote for doing game development in Go. I'm actively using it and its pretty stable, but I do find bugs every once in a while. I might vary the APIs in the future if native iterators are added.

### Overview
Conceptually you can imagine an ECS as one big table, where an `Id` column associates an *Entity Id* with various other component columns. Kind of like this:

| Id | Position | Rotation | Size |
|:--:|:--------:|:--------:|:----:|
| 0  | {1, 1}   | 3.14     | 11   |
| 1  | {2, 2}   | 6.28     | 22   |

We use an archetype-based storage mechanism. Which simply means we have a specific table for a specific component layout. This means that if you add or remove components it can be somewhat expensive, because we have to copy the entire entity to the new table.

### Basic Usage
Import the library: `import "github.com/unitoftime/ecs"`

Create Components like you normally would:
```
type Position struct {
    X, Y float64
}

type Rotation float64
```

Create a `World` to store all of your data
```
world := ecs.NewWorld()
```

Create an entity and add components to it
```
id := world.NewId()
ecs.Write(world, id,
    ecs.C(Position{1, 1}),
    ecs.C(Rotation(3.14)),
    // Note: Try to reduce the number of write calls by packing as many components as you can in
)

// Side-Note: I'm trying to get rid of the `ecs.C(...)` boxing, but I couldn't figure out how when
//            I first wrote the ECS. I'll try to get back to fixing that because ideally you
//            shouldn't have to worry about it. For now though, you have to box your components
//            to the `ecs.Component` interface type before passing them in, so `ecs.C(...)`
//            does that for you.
```

Create a View, by calling `QueryN`:
```
query := ecs.Query2[Position, Rotation](world)
```

Iterate on the query. You basically pass in a lambda, and internally the library calls it for every entity in the world which has all of the components specified. Notably your lambda takes pointer values which represent a pointer to the internally stored component. So modifying these pointers will modify the entity's data.
```
query.MapId(func(id ecs.Id, pos *Position, rot *Rotation) {
    pos.X += 1
    pos.Y += 1

    rot += 0.01
})
```

There are several map functions you can use, each with varying numbers of parameters. I support up to `Map12`. They all look like this:
```
ecs.MapN(world, func(id ecs.Id, a *ComponentA, /*... */, n *ComponentN) {
    // Do your work
})
```

#### Advanced queries
You can also filter your queries for more advanced usage:
```
// Returns a view of Position and Velocity, but only if the entity also has the `Rotation` component.
query := ecs.Query2[Position, Velocity](world, ecs.With(Rotation))

// Returns a view of Position and Velocity, but if velocity is missing on the entity, will just return nil during the `MapId(...)`. You must do nil checks for all components included in the `Optional()`!
query := ecs.Query2[Position, Velocity](world, ecs.Optional(Velocity))
```

### Systems
There are several types of systems:
    1. Realtime - should be runned as fast as possible (without any deplays). Usefull for Rendering or collecting inputs from user.
    2. Fixed - should be runned ones a specified period of time. Usefull for physics.
    3. Step - can only be runned by manually triggering. Usefull for simulations, like stable updates in P2P games.

To create system, implement one of the interfaces `RealtimeSystem`, `FixedSystem`, `StepSystem`:
```
type Position struct {
    X, Y float64
}
type Velocity struct {
    X, Y float64
}

// --- world initialization here ---

type movementSystem struct {}

func (s *movementSystem) GetName() {
    return "physics::movement"
}

func (s *movementSystem) RunFixed(delta time.Duration) {
    query := ecs.Query2[Position, Velocity](world)
    query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
        // No need to adjust here to delta time between updates, because
        // physics steps should always be the same. Delta time is still available to systems, where
        // it can be important (f.e. if we want to detect "throttling")
        pos.X += vel.X
        pos.Y += vel.Y

        // Add some drag
        vel.X *= 0.99
        vel.Y *= 0.99
    })
}
```

To shedule the system, we need to have a system group of one of the types `RealtimeGroup`, `FixedGroup`, `StepGroup`. There are several rules regarding system groups:
1. One system cant belong to multiple groups.
2. Groups are running in parallel and there is no synchronization between them. 
3. All the systems in the group can be automatically runned in parallel.
4. Groups share lock to the components on the systems level, so there will be no race conditions while accessing components data.
5. Group can only contain systems with theirs type.

```
componentsGuard = group.NewComponentsGuard()
physics = group.NewFixedGroup("physics", 50*time.Millisecond, componentsGuard)
physics.AddSystem(&movementSystem{})
physics.Build()
physics.StartFixed()
defer physics.StopFixed()
```

#### Order
You can specify systems order by implementing `system.RunBeforeSystem` or/and `system.RunAfterSystem` interface for the system. You cant implement order between systems from multiple groups.

#### Automatic parallelism
In order for sheduler to understand which systems can be runed in parallel, you have to specify components used by systems. By default, if component will not be specified, system will be marked as `exclusive` and will only be runned by locking entire ECS world.
You can do this by implementing `system.ReadComponentsSystem` or/and `system.WriteComponentsSystem` interface for the system. Make sure to use `system.ReadComponentsSystem` as much as possible, because with read access, multiple systems can be runned at the same time.

### Commands

Commands will eventually replace `ecs.Write(...)` once I figure out how their usage will work. Commands essentially buffer some work on the ECS so that the work can be executed later on. You can use them in loop safe ways by calling `Execute()` after your loop has completed. Right now they work like this:
```
cmd := ecs.NewCommand(world)
WriteCmd(cmd, id, Position{1,1,1})
WriteCmd(cmd, id, Velocity{1,1,1})
cmd.Execute()
```

### Still In Progress
- [ ] Improving iterator performance: See: https://github.com/golang/go/discussions/54245
- [+] Automatic multithreading
- [ ] Without() filter

### Videos
Hopefully, eventually I can have some automated test-bench that runs and measures performance, but for now you'll just have to refer to my second video and hopefully trust me. Of course, you can run the benchmark in the `bench` folder to measure how long frames take on your computer.

1. How it works: https://youtu.be/71RSWVyOMEY
2. Simulation Performance: https://youtu.be/i2gWDOgg50k
