package ecs

import (
	"time"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const MovementSystemName system.SystemName = "physics::movement"

type MovementSystem struct {
	World *ecs.World
}

func (s *MovementSystem) GetName() system.SystemName {
	return MovementSystemName
}

func (s *MovementSystem) GetReadComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Velocity{})}
}

func (s *MovementSystem) GetWriteComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Position{})}
}

func (s *MovementSystem) RunFixed(delta time.Duration) {
	query := ecs.Query2[Position, Velocity](s.World)
	query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
		pos.X += vel.X
		pos.Y += vel.Y
	})
}
