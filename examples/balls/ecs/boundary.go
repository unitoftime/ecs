package ecs

import (
	"time"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const BoundarySystemName system.SystemName = "physics::boundary"
const WorldSize = 50.0

type BoundarySystem struct {
	World *ecs.World
}

func (s *BoundarySystem) GetName() system.SystemName {
	return BoundarySystemName
}

func (s *BoundarySystem) GetWriteComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Velocity{}), ecs.C(Position{})}
}

func (s *BoundarySystem) RunFixed(delta time.Duration) {
	query := ecs.Query2[Position, Velocity](s.World)
	query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
		if pos.X < 0 {
			pos.X = 0
			if vel.X < 0 {
				vel.X = -vel.X
			}
		}
		if pos.X > WorldSize {
			pos.X = WorldSize
			if vel.X > 0 {
				vel.X = -vel.X
			}
		}

		if pos.Y < 0 {
			pos.Y = 0
			if vel.Y < 0 {
				vel.Y = -vel.Y
			}
		}
		if pos.Y > WorldSize {
			pos.Y = WorldSize
			if vel.Y > 0 {
				vel.Y = -vel.Y
			}
		}
	})
}
