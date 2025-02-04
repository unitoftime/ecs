package ecs

import (
	"math/rand"
	"testing"
)

func TestCommandSingleRewrite(t *testing.T) {
	world := NewWorld()
	cmd := NewCommandQueue(world)

	a := world.NewId()
	b := world.NewId()
	c := world.NewId()
	d := world.NewId()
	Write(world, a, C(position{}), C(velocity{}))
	Write(world, b, C(position{}), C(velocity{}))
	Write(world, c, C(position{}), C(velocity{}))
	Write(world, d, C(position{}), C(velocity{}))
	DeleteComponent(world, a, C(velocity{}))
	DeleteComponent(world, b, C(velocity{}))
	DeleteComponent(world, c, C(velocity{}))
	DeleteComponent(world, d, C(velocity{}))

	queryPos := Query1[position](world)
	queryVel := Query1[velocity](world)

	queryPos.MapId(func(id Id, pos *position) {
		cmd.Write(id).Insert(C(velocity{}))
	})
	check(t, *queryPos.Read(a) == position{})
	check(t, *queryPos.Read(b) == position{})
	check(t, *queryPos.Read(c) == position{})
	check(t, *queryPos.Read(d) == position{})

	check(t, queryVel.Read(a) == nil)
	check(t, queryVel.Read(b) == nil)
	check(t, queryVel.Read(c) == nil)
	check(t, queryVel.Read(d) == nil)

	cmd.Execute()

	check(t, *queryPos.Read(a) == position{})
	check(t, *queryPos.Read(b) == position{})
	check(t, *queryPos.Read(c) == position{})
	check(t, *queryPos.Read(d) == position{})

	check(t, *queryVel.Read(a) == velocity{})
	check(t, *queryVel.Read(b) == velocity{})
	check(t, *queryVel.Read(c) == velocity{})
	check(t, *queryVel.Read(d) == velocity{})
}

func TestCommandWrites(t *testing.T) {
	world := NewWorld()
	commands := NewCommandQueue(world)

	type data struct {
		id  Id
		pos position
		vel velocity
		acc acceleration
		rad radius
	}

	expected := make([]data, 1000)
	for i := range expected {
		ent := commands.SpawnEmpty()

		expected[i] = data{
			id:  ent.Id(),
			pos: position{1, rand.Float64() * 100, rand.Float64() * 100},
			vel: velocity{2, rand.Float64() * 100, rand.Float64() * 100},
			acc: acceleration{3, rand.Float64() * 100, rand.Float64() * 100},
			rad: radius{rand.Float64() * 100},
		}
		ent.
			Insert(expected[i].pos).
			Insert(expected[i].vel).
			Insert(expected[i].acc).
			Insert(expected[i].rad)
	}

	commands.Execute()

	for i := range expected {
		id := expected[i].id

		pos, ok := Read[position](world, id)
		check(t, ok)
		compare(t, pos, expected[i].pos)

		vel, ok := Read[velocity](world, id)
		check(t, ok)
		compare(t, vel, expected[i].vel)

		acc, ok := Read[acceleration](world, id)
		check(t, ok)
		compare(t, acc, expected[i].acc)

		rad, ok := Read[radius](world, id)
		check(t, ok)
		compare(t, rad, expected[i].rad)
	}
}

func TestWorldReadWriteNew(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	// Write position
	pos := position{1, 1, 1}
	Write(world, id, pos)

	// Check position and velocity
	posOut, ok := Read[position](world, id)
	check(t, ok)
	compare(t, posOut, pos)
	velOut, ok := Read[velocity](world, id)
	check(t, !ok) // We expect this to be false
	compare(t, velOut, velocity{0, 0, 0})

	// Write velocity
	vel := velocity{2, 2, 2}
	Write(world, id, vel)

	// Check position and velocity
	posOut, ok = Read[position](world, id)
	check(t, ok)
	compare(t, posOut, pos)
	velOut, ok = Read[velocity](world, id)
	check(t, ok)
	compare(t, velOut, vel)

	compare(t, world.engine.count(position{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(position{}, velocity{}), 1)
	compare(t, world.engine.count(acceleration{}), 0)

	// count := 0
	// Map2(world, func(id Id, p *position, v *velocity) {
	// 	count++
	// })
	// compare(t, count, 1)

	// count = 0
	// view := ViewAll2[position, velocity](world)
	// for {
	// 	_, _, _, ok := view.Iter()
	// 	if !ok { break }
	// 	count++
	// }
	// compare(t, count, 1)
}

func TestWorldReadMultiWriteNew(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	pos := position{1, 1, 1}
	vel := velocity{2, 2, 2}
	Write(world, id, pos, vel)

	// Check position and velocity
	posOut, ok := Read[position](world, id)
	check(t, ok)
	compare(t, posOut, pos)
	velOut, ok := Read[velocity](world, id)
	check(t, ok)
	compare(t, velOut, vel)

	// Write accel and size
	accel := acceleration{3, 3, 3}
	rad := radius{4}
	Write(world, id, accel, rad)

	// Check all
	posOut, ok = Read[position](world, id)
	check(t, ok)
	compare(t, posOut, pos)
	velOut, ok = Read[velocity](world, id)
	check(t, ok)
	compare(t, velOut, vel)
	accelOut, ok := Read[acceleration](world, id)
	check(t, ok)
	compare(t, accelOut, accel)
	radOut, ok := Read[radius](world, id)
	check(t, ok)
	compare(t, radOut, rad)
}

func TestCommandReadWrite(t *testing.T) {
	world := NewWorld()
	cmd := GetInjectable[*CommandQueue](world)

	{
		e := cmd.SpawnEmpty().
			Insert(position{1, 2, 3})

		p, ok := ReadComp[position](e)
		check(t, ok)
		compare(t, p, position{1, 2, 3})

		v, ok := ReadComp[velocity](e)
		check(t, !ok)
		compare(t, v, velocity{})

		e.Insert(velocity{4, 5, 6})

		v, ok = ReadComp[velocity](e)
		check(t, ok)
		compare(t, v, velocity{4, 5, 6})
	}

	cmd.Execute()

	{
		e := cmd.SpawnEmpty().
			Insert(position{1, 2, 3})

		p, ok := ReadComp[position](e)
		check(t, ok)
		compare(t, p, position{1, 2, 3})

		v, ok := ReadComp[velocity](e)
		check(t, !ok)
		compare(t, v, velocity{})

		e.Insert(velocity{4, 5, 6})

		v, ok = ReadComp[velocity](e)
		check(t, ok)
		compare(t, v, velocity{4, 5, 6})
	}

	cmd.Execute()
}

func TestCommandCancel(t *testing.T) {
	world := NewWorld()
	cmd := GetInjectable[*CommandQueue](world)

	e := cmd.SpawnEmpty().
		Insert(position{1, 2, 3})

	e.Cancel()

	cmd.Execute()

	p, ok := Read[position](world, e.Id())
	check(t, !ok)
	compare(t, p, position{})
}

type testEvent struct {
	val int
}

var _testEventId = NewEvent[testEvent]()

func (p testEvent) EventId() EventId {
	return _testEventId
}

func TestCommandTrigger(t *testing.T) {
	world := NewWorld()
	cmd := GetInjectable[*CommandQueue](world)

	runCount := 0
	world.AddObserver(
		NewHandler(func(trigger Trigger[testEvent]) {
			compare(t, trigger.Data.val, 55)

			if runCount == 0 {
				compare(t, trigger.Id, 33)
			} else if runCount == 1 {
				compare(t, trigger.Id, 44)
			} else {
				compare(t, trigger.Id, InvalidEntity)
			}

			runCount++
		}))

	cmd.Trigger(testEvent{55}, 33, 44) // Emit for id 33 then 44
	cmd.Execute()

	compare(t, runCount, 2) // Should run twice (33, 44)

	cmd.Trigger(testEvent{55}) // Emit for no specific entity
	cmd.Execute()

	compare(t, runCount, 3) // Should run once (nonspecific)
}
