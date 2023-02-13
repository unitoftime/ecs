package ecs

// Represents a list of commands that need to be executed on the world
type Command struct {
	world *World
	list  map[Id]*writeCmd // TODO - Note to self: if you ever add deletion inside of commands, then the packing commands into a map based on entity Id assumption wont hold, because you'll need some amount of specific ordering
}

// Create a new command to be executed
func NewCommand(world *World) *Command {
	return &Command{
		world: world,
		list:  make(map[Id]*writeCmd),
	}
}

// Execute the command
func (c *Command) Execute() {
	// TODO - Batch similar commands, if you ever switch to something more complex than just writing

	// Execute all the commands
	for i := range c.list {
		c.list[i].execute(c.world)
	}

	// Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	for k := range c.list {
		delete(c.list, k)
	}
}

// TODO - maybe rename as just Write?
// Adds a write command
func WriteCmd[A any](c *Command, id Id, comp A) {
	cmd, ok := c.list[id]
	if !ok {
		cmd = newWriteCmd(id)
		c.list[id] = cmd
	}

	cmd.comps = append(cmd.comps, C(comp))
}

// type cmd interface {
// 	execute(*World)
// }

type writeCmd struct {
	id    Id
	comps []Component
}

func newWriteCmd(id Id) *writeCmd {
	return &writeCmd{
		id:    id,
		comps: make([]Component, 0, 2), // TODO - guaranteed to at least have 1, but a bit arbitrary
	}
}
func (c *writeCmd) execute(world *World) {
	world.Write(c.id, c.comps...)
}
