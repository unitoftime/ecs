package ecs

// type singleCmd interface {
// 	apply(*World)
// }

// type spawnCmd struct {
// 	bundler *Bundler
// }
// func (c spawnCmd) apply(world *World) {
// 	id := world.NewId()
// 	c.bundler.Write(world, id)
// }

type CmdType uint8

const (
	CmdTypeNone CmdType = iota
	CmdTypeSpawn
	CmdTypeWrite
	CmdTypeTrigger
	CmdTypeDelete
	// CmdTypeCustom
)

type singleCmd struct {
	Type    CmdType
	id      Id
	bundler *Bundler
	event   Event
}

func (c *singleCmd) apply(world *World) {
	switch c.Type {
	case CmdTypeNone:
		// Do nothing, Command was probably cancelled
	case CmdTypeSpawn:
		if world.cmd.preWrite != nil {
			world.cmd.preWrite(EntityCommand{c})
		}
		c.bundler.Write(world, c.id) // TODO: This could probably use a Spawn function which would be faster
	case CmdTypeWrite:
		if world.cmd.preWrite != nil {
			world.cmd.preWrite(EntityCommand{c})
		}
		c.bundler.Write(world, c.id)
	case CmdTypeTrigger:
		world.Trigger(c.event, c.id)
	case CmdTypeDelete:
		if world.cmd.preDelete != nil {
			world.cmd.preDelete(c.id)
		}
		Delete(world, c.id)
	}
}

type EntityCommand struct {
	cmd *singleCmd
}

// func (e EntityCommand) Printout() {
// 	fmt.Println("---")
// 	for i := range e.cmd.bundler.Components {
// 		if e.cmd.bundler.Set[i] {
// 			fmt.Printf("+%v\n", e.cmd.bundler.Components[i])
// 		}
// 	}
// 	// fmt.Printf("+%v\n", e.cmd.bundler)
// }

func (e EntityCommand) Cancel() {
	e.cmd.Type = CmdTypeNone
}

// Removes the supplied component type from this entity command.
// TODO: Should this also remove it from the world? if it exists there?
func (e EntityCommand) Remove(comp Component) {
	compId := comp.CompId()
	e.cmd.bundler.Remove(compId)
}

func (e EntityCommand) Empty() bool {
	return (e == EntityCommand{})
}

func (e EntityCommand) Insert(bun Writer) EntityCommand {
	unbundle(bun, e.cmd.bundler)
	return e
}

// // Inserts the component if it is missing
// func (e EntityCommand) InsertIfMissing(bun Component) EntityCommand {
// 	if e.cmd.bundler.Has(bun) {
// 		return e
// 	}

// 	unbundle(bun, e.cmd.bundler)
// 	return e
// }

func (e EntityCommand) Id() Id {
	return e.cmd.id
}

// func (e EntityCommand) Remove(bun Bundle) EntityCommand {
// 	bun.Unbundle(e.cmd.bundler)
// 	return e
// }

//	func (e EntityCommand) Add(seq iter.Seq[Component]) EntityCommand {
//		for c := range seq {
//			e.cmd.bundler.Add(c)
//		}
//		return e
//	}
func ReadComp[T Component](e EntityCommand) (T, bool) {
	var t T
	// comp, ok := e.cmd.bundler.Read(t)
	// if ok {
	// 	box := comp.(*box[T])
	// 	return box.val, true
	// }

	comp, ok := readBundle[T](e.cmd.bundler)
	if ok {
		return comp, true
	}
	return t, false
}

type CommandQueue struct {
	world    *World
	preWrite func(EntityCommand)
	preDelete func(Id)
	commands []singleCmd

	currentBundlerIndex int
	bundlers            []*Bundler
}

func NewCommandQueue(world *World) *CommandQueue {
	return &CommandQueue{
		world: world,
	}
}
func (c *CommandQueue) initialize(world *World) any {
	return NewCommandQueue(world)
}

func (c *CommandQueue) NextBundler() *Bundler {
	if c.currentBundlerIndex >= len(c.bundlers) {
		bundler := &Bundler{}
		c.bundlers = append(c.bundlers, bundler)
		c.currentBundlerIndex = len(c.bundlers)
		return bundler
	} else {
		bundler := c.bundlers[c.currentBundlerIndex]
		bundler.Clear()
		c.currentBundlerIndex++
		return bundler
	}
}

func unbundle(bundle Writer, bundler *Bundler) {
	wd := W{bundler: bundler}
	bundle.CompWrite(wd)
}

func remove(bundle Writer, bundler *Bundler) {
	wd := W{bundler: bundler}
	bundle.CompWrite(wd)
}

// func CmdSpawn[T Writer](c *CommandQueue, ub T) {
// 	bundler := c.NextBundler()
// 	unbundle(ub, bundler)
// 	// ub.Unbundle(bundler)
// 	c.commands = append(c.commands, singleCmd{
// 		Type:    CmdTypeSpawn,
// 		id:      c.world.NewId(),
// 		bundler: bundler,
// 	})
// }

// func (c *CommandQueue) Spawn(bun Writer) {
// 	entCmd := c.SpawnEmpty()
// 	entCmd.Insert(bun)
// }

func (c *CommandQueue) SpawnEmpty() EntityCommand {
	bundler := c.NextBundler()

	c.commands = append(c.commands, singleCmd{
		Type:    CmdTypeSpawn,
		id:      c.world.NewId(),
		bundler: bundler,
	})
	return EntityCommand{
		cmd: &(c.commands[len(c.commands)-1]),
	}
}

// // Pushes a command to delete the entity
// func (c *CommandQueue) Delete(id Id) {
// 	c.commands = append(c.commands, singleCmd{
// 		Type: CmdTypeDelete,
// 		id: id,
// 	})
// }

func (c *CommandQueue) Write(id Id) EntityCommand {
	bundler := c.NextBundler()

	c.commands = append(c.commands, singleCmd{
		Type:    CmdTypeWrite,
		id:      id,
		bundler: bundler,
	})
	return EntityCommand{
		cmd: &(c.commands[len(c.commands)-1]),
	}
}

func (c *CommandQueue) Trigger(event Event, ids ...Id) {
	// Special Case: no ids provided, so just trigger a single, unrelated
	if len(ids) == 0 {
		c.commands = append(c.commands, singleCmd{
			Type:  CmdTypeTrigger,
			id:    InvalidEntity,
			event: event,
		})

		return
	}

	for _, id := range ids {
		c.commands = append(c.commands, singleCmd{
			Type:  CmdTypeTrigger,
			id:    id,
			event: event,
		})
	}
}

// Adds a prewrite function to be executed before every write or spawn command is executed
// Useful for ensuring entities are fully formed before pushing them into the ECS
func (c *CommandQueue) SetPrewrite(lambda func(EntityCommand)) {
	c.preWrite = lambda
}

// // Adds a predelite function to be executed before every delete command is executed
// // Useful for ensuring any external datastructures get cleaned up when an entity is deleted
// func (c *CommandQueue) SetPredelete(lambda func(Id)) {
// 	c.preDelete = lambda
// }

func (c *CommandQueue) Execute() {
	// Perform all commands
	// Note: We must check length every time in case calling one command adds more commands
	for i := 0; i < len(c.commands); i++ {
		c.commands[i].apply(c.world)
	}

	// Cleanup Queue
	c.commands = c.commands[:0]
	c.currentBundlerIndex = 0
}

// TODO: Maybe?
// func (c *CommandQueue) ExecutePostWrite(postWrite func (ecs.Id)) {
// 	// Perform all commands
// 	for i := range c.commands {
// 		c.commands[i].apply(c.world)
// 	}
// 	for i := range c.commands {
// 		if c.commands[i].id == InvalidEntity {
// 			continue
// 		}
// 		postWrite(c.commands[i].id)
// 	}

// 	// Cleanup Queue
// 	c.commands = c.commands[:0]
// 	c.currentBundlerIndex = 0
// }
