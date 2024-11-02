package ecs

import "fmt"

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
	CmdTypeCustom
)

type singleCmd struct {
	Type    CmdType
	id      Id
	bundler *Bundler
}

func (c singleCmd) apply(world *World) {
	switch c.Type {
	case CmdTypeNone:
		// Do nothing, Command was probably cancelled
	case CmdTypeSpawn:
		c.bundler.Write(world, c.id)
	case CmdTypeWrite:
		c.bundler.Write(world, c.id)
	}
}

type EntityCommand struct {
	cmd *singleCmd
}

func (e EntityCommand) Printout() {
	fmt.Println("---")
	for i := range e.cmd.bundler.Components {
		if e.cmd.bundler.Set[i] {
			fmt.Printf("+%v\n", e.cmd.bundler.Components[i])
		}
	}
	// fmt.Printf("+%v\n", e.cmd.bundler)
}

func (e EntityCommand) Cancel() {
	e.cmd.Type = CmdTypeNone
}

func (e EntityCommand) Empty() bool {
	return (e == EntityCommand{})
}

func (e EntityCommand) Insert(bun Writer) EntityCommand {
	unbundle(bun, e.cmd.bundler)
	return e
}

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
	comp, ok := e.cmd.bundler.Read(t)
	if ok {
		box := comp.(*box[T])
		return box.val, true
	}
	return t, false
}

type CommandQueue struct {
	world    *World
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

func CmdSpawn[T Writer](c *CommandQueue, ub T) {
	bundler := c.NextBundler()
	unbundle(ub, bundler)
	// ub.Unbundle(bundler)
	c.commands = append(c.commands, singleCmd{
		Type:    CmdTypeSpawn,
		id:      c.world.NewId(),
		bundler: bundler,
	})
}

func (c *CommandQueue) Spawn(bun Writer) {
	entCmd := c.SpawnEmpty()
	entCmd.Insert(bun)
}

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

func (c *CommandQueue) Execute() {
	// Perform all commands
	for i := range c.commands {
		c.commands[i].apply(c.world)
	}

	// Cleanup Queue
	c.commands = c.commands[:0]
	c.currentBundlerIndex = 0
}
