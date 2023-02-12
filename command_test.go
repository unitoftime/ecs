package ecs

import (
	"testing"
)

func TestCommandExecution(t *testing.T) {
	world := NewWorld()

	cmd := NewCommand(world)
	query := Query2[position, velocity](world)

	// Write position
	id := world.NewId()
	pos := position{1,1,1}
	WriteCmd(cmd, id, pos)
	cmd.Execute()

	// Check position and velocity
	posOut, velOut := query.Read(id)
	compare(t, *posOut, pos)
	compare(t, velOut, nil)

	// Write velocity
	vel := velocity{2,2,2}
	WriteCmd(cmd, id, vel)
	cmd.Execute()

	// Check position and velocity
	posOut, velOut = query.Read(id)
	compare(t, *posOut, pos)
	compare(t, *velOut, vel)

	compare(t, world.engine.count(position{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(acceleration{}), 0)

	count := 0
	query.MapId(func(id Id, p *position, v *velocity) {
		count++
	})
	compare(t, count, 1)

	// count = 0
	// view := ViewAll2[position, velocity](world)
	// for {
	// 	_, _, _, ok := view.Iter()
	// 	if !ok { break }
	// 	count++
	// }
	// compare(t, count, 1)
}
