package ecs

import (
	"runtime"
	"testing"
)

var positionId = C(position{})
var velocityId = C(velocity{})
var accelerationId = C(acceleration{})
var radiusId = C(radius{})

func (p position) CompId() CompId {
	return positionId.CompId()
}
func (p velocity) CompId() CompId {
	return velocityId.CompId()
}
func (p acceleration) CompId() CompId {
	return accelerationId.CompId()
}
func (p radius) CompId() CompId {
	return radiusId.CompId()
}

func (p position) CompWrite(cw W) {
	positionId.WriteVal(cw, p)
}
func (p velocity) CompWrite(cw W) {
	velocityId.WriteVal(cw, p)
}
func (p acceleration) CompWrite(cw W) {
	accelerationId.WriteVal(cw, p)
}
func (p radius) CompWrite(cw W) {
	radiusId.WriteVal(cw, p)
}

type position struct {
	x, y, z float64
}

type velocity struct {
	x, y, z float64
}

type acceleration struct {
	x, y, z float64
}

type radius struct {
	r float64
}

// Helper functions

// Check that this boolean is true
func check(t *testing.T, b bool) {
	if !b {
		_, f, l, _ := runtime.Caller(1)
		t.Errorf("%s:%d - checked boolean is false!", f, l)
	}
}

// Check two things match, if they don't, throw an error
func compare[T comparable](t *testing.T, actual, expected T) {
	if expected != actual {
		_, f, l, _ := runtime.Caller(1)
		t.Errorf("%s:%d - actual(%v) did not match expected(%v)", f, l, actual, expected)
	}
}

func TestWorldReadWrite(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	// Write position
	pos := position{1, 1, 1}
	Write(world, id, C(pos))

	// Check position and velocity
	posOut, ok := Read[position](world, id)
	check(t, ok)
	compare(t, posOut, pos)
	velOut, ok := Read[velocity](world, id)
	check(t, !ok) // We expect this to be false
	compare(t, velOut, velocity{0, 0, 0})

	// Write velocity
	vel := velocity{2, 2, 2}
	Write(world, id, C(vel))

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

func TestWorldReadMultiWrite(t *testing.T) {
	world := NewWorld()
	id := world.NewId()

	pos := position{1, 1, 1}
	vel := velocity{2, 2, 2}
	Write(world, id, C(pos), C(vel))

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
	Write(world, id, C(accel), C(rad))

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

func TestWorldWriteDelete(t *testing.T) {
	world := NewWorld()
	ids := make([]Id, 0)
	for i := 0; i < 1e6; i++ {
		id := world.NewId()
		v := float64(id)
		pos := position{v, v, v}
		vel := velocity{v, v, v}
		Write(world, id, C(pos), C(vel))
		ids = append(ids, id)
	}

	// Verify they are all correct
	for _, id := range ids {
		expected := float64(id)

		posOut, ok := Read[position](world, id)
		check(t, ok)
		compare(t, posOut, position{expected, expected, expected})

		velOut, ok := Read[velocity](world, id)
		check(t, ok)
		compare(t, velOut, velocity{expected, expected, expected})
	}

	// Delete every even index
	for i, id := range ids {
		if i%2 == 0 {
			Delete(world, id)
		}
	}

	// Verify they are all correct
	for i, id := range ids {
		expected := float64(id)

		if i%2 == 0 {
			// Expect these to be deleted in the world
			expected = 0.0
			posOut, ok := Read[position](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
			compare(t, posOut, position{expected, expected, expected})

			velOut, ok := Read[velocity](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
			compare(t, velOut, velocity{expected, expected, expected})
		} else {
			// Expect these to still exist in the world
			posOut, ok := Read[position](world, id)
			check(t, ok)
			compare(t, posOut, position{expected, expected, expected})

			velOut, ok := Read[velocity](world, id)
			check(t, ok)
			compare(t, velOut, velocity{expected, expected, expected})
		}
	}
}

func TestWorldDeleteAllViaComponents(t *testing.T) {
	world := NewWorld()

	id := world.NewId()
	world.Write(id, C(position{}), C(velocity{}), C(acceleration{}), C(radius{}))

	DeleteComponent(world, id, C(position{}))
	DeleteComponent(world, id, C(velocity{}))
	DeleteComponent(world, id, C(acceleration{}))
	DeleteComponent(world, id, C(radius{}))
	_, ok := Read[position](world, id)
	check(t, !ok)
	_, ok = Read[velocity](world, id)
	check(t, !ok)
	_, ok = Read[acceleration](world, id)
	check(t, !ok)
	_, ok = Read[radius](world, id)
	check(t, !ok)

	exists := world.Exists(id)
	check(t, !exists)
}

func TestWorldDeleteComponent(t *testing.T) {
	world := NewWorld()
	ids := make([]Id, 0)
	for i := 0; i < 1e6; i++ {
		id := world.NewId()
		v := float64(id)
		pos := position{v, v, v}
		vel := velocity{v, v, v}
		Write(world, id, C(pos), C(vel))
		ids = append(ids, id)
	}

	// Verify they are all correct
	for _, id := range ids {
		expected := float64(id)

		posOut, ok := Read[position](world, id)
		check(t, ok)
		compare(t, posOut, position{expected, expected, expected})

		velOut, ok := Read[velocity](world, id)
		check(t, ok)
		compare(t, velOut, velocity{expected, expected, expected})
	}

	// different deletes for different modulos
	for i, id := range ids {
		if i%2 == 0 {
			DeleteComponent(world, id, C(position{}))
		} else if i%3 == 0 {
			DeleteComponent(world, id, C(velocity{}))
		} else if i%5 == 0 {
			DeleteComponent(world, id, C(position{}), C(velocity{}))
		} else if i%7 == 0 {
			DeleteComponent(world, id, C(velocity{}))
			DeleteComponent(world, id, C(position{}))
		} else if i%13 == 0 {
			DeleteComponent(world, id, C(radius{})) // Note: This shouldn't do anything
		}
	}

	// Verify they are all correct
	for i, id := range ids {
		expected := float64(id)

		if i%2 == 0 {
			// Expect these to be deleted in the world
			_, ok := Read[position](world, id)
			check(t, !ok) // Expect to be false because we've deleted this

			velOut, ok := Read[velocity](world, id)
			check(t, ok)
			compare(t, velOut, velocity{expected, expected, expected})
		} else if i%3 == 0 {
			// Expect these to still exist in the world
			posOut, ok := Read[position](world, id)
			check(t, ok)
			compare(t, posOut, position{expected, expected, expected})

			_, ok = Read[velocity](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
		} else if i%5 == 0 {
			_, ok := Read[position](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
			_, ok = Read[velocity](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
		} else if i%7 == 0 {
			_, ok := Read[position](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
			_, ok = Read[velocity](world, id)
			check(t, !ok) // Expect to be false because we've deleted this
		} else {
			// Expect these to still exist in the world
			posOut, ok := Read[position](world, id)
			check(t, ok)
			compare(t, posOut, position{expected, expected, expected})
			velOut, ok := Read[velocity](world, id)
			check(t, ok)
			compare(t, velOut, velocity{expected, expected, expected})
		}
	}
}

func TestResources(t *testing.T) {
	world := NewWorld()
	p := position{1, 2, 3}

	p0 := GetResource[position](world)
	compare(t, p0, nil) // should be nil b/c it isnt added yet

	PutResource(world, &p)

	p1 := GetResource[position](world)
	compare(t, p1, &p) // Should match the original pointer
	compare(t, *p1, p)
}
