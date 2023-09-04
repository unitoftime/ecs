package ecs

// Represents a list of commands that need to be executed on the world
type Command struct {
	world *World
	list  map[Id]*writeCmd // TODO - Note to self: if you ever add deletion inside of commands, then the packing commands into a map based on entity Id assumption wont hold, because you'll need some amount of specific ordering
	dynamicBundle dynamicBundle
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

func (w *World) Spawn() Ent {
	return Ent{
		id: w.NewId(),
	}
}

type Ent struct {
	id Id
	comps []Component
}

// func NewWriter[T any]() *Writer[T] {
// 	var t T
// 	return &Writer[T]{
// 		comp: C(t), // TODO: combine when you remove Box[T]
// 	}
// }
// type Writer[T any] struct {
// 	comp Box[T]
// }
// func (w *Writer[T]) Write(ent *Ent, t T) *Ent {
// 	w.Comp = t
// 	ent.comps = append(ent, w.Comp)
// }

func (c *Command) Spawn(bundles ...unbundler) Id {
	c.dynamicBundle.comps = c.dynamicBundle.comps[:0]
	for _, b := range bundles {
		b.unbundleInto(&c.dynamicBundle)
	}
	id := c.world.NewId()
	c.world.Write(id, c.dynamicBundle.comps...)
	return id
}

type dynamicBundle struct {
	comps []Component
}
func (b *dynamicBundle) unbundleInto(d *dynamicBundle) {
	d.comps = append(d.comps, b.comps...)
}

type unbundler interface {
	unbundleInto(*dynamicBundle)
}

type Bundle2[A, B any] struct {
	wa Box[A]
	wb Box[B]
}
func NewBundle2[A, B any]() *Bundle2[A,B] {
	var a A
	var b B
	return &Bundle2[A,B]{
		wa: C(a),
		wb: C(b),
	}
}

func (bun *Bundle2[A,B]) With(a A, b B) *Bundle2[A, B] {
	ret := &Bundle2[A,B]{
		wa: bun.wa,
		wb: bun.wb,
	}
	ret.wa.Comp = a
	ret.wb.Comp = b
	return ret
}

func (b *Bundle2[A,B]) unbundleInto(d *dynamicBundle) {
	d.comps = append(d.comps, b.wa, b.wb)
}
