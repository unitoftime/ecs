package ecs

import "testing"

func TestQuery(t *testing.T) {
	world := NewWorld()
	ids := make([]Id, 0)
	for i := 0; i < 1e6; i++ {
		id := world.NewId()
		v := float64(id)
		pos := position{v, v, v}
		vel := velocity{v, v, v}
		Write(world, id, pos, vel)
		ids = append(ids, id)
	}

	m := make(map[Id]struct{})
	query := Query1[position](world)
	query.MapId(func(id Id, pos *position) {
		compare(t, *pos, position{float64(id), float64(id), float64(id)})
		m[id] = struct{}{}
	})

	for _, id := range ids {
		_, ok := m[id]
		check(t, ok)
	}
}
