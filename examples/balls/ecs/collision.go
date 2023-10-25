package ecs

import (
	"math"
	"time"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const CollisionSystemName system.SystemName = "physics::collision"

type CollisionSystem struct {
	World       *ecs.World
	SpacialHash map[IntVector2][]SpacialCellObject
}

func (s *CollisionSystem) GetName() system.SystemName {
	return CollisionSystemName
}

func (s *CollisionSystem) GetReadComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Position{})}
}

func (s *CollisionSystem) GetWriteComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Velocity{})}
}

func (s *CollisionSystem) GetRunAfter() []system.SystemName {
	return []system.SystemName{MovementSystemName, SpacialSystemName}
}

func (s *CollisionSystem) RunFixed(delta time.Duration) {
	query := ecs.Query2[Position, Velocity](s.World)
	query.MapId(func(id ecs.Id, pos *Position, vel *Velocity) {
		cellId := IntVector2{
			X: int32(math.Round(pos.X / SpacialCellSize)),
			Y: int32(math.Round(pos.Y / SpacialCellSize)),
		}
		objects := s.SpacialHash[cellId]

		for _, object := range objects {
			solveCollision := func(otherCell IntVector2) {
				if otherObjects, ok := s.SpacialHash[otherCell]; ok {
					speed := math.Sqrt(math.Pow(object.X, 2) + math.Pow(object.Y, 2))

					for _, otherObject := range otherObjects {
						distance := math.Sqrt(math.Pow(otherObject.X-object.X, 2) + math.Pow(otherObject.Y-object.Y, 2))
						if distance < 2 {
							vel.X = (object.X - otherObject.X) / distance * speed
							vel.Y = (object.Y - otherObject.Y) / distance * speed
						}
					}
				}
			}

			solveCollision(cellId)
			solveCollision(IntVector2{X: cellId.X + 1, Y: cellId.Y + 1})
			solveCollision(IntVector2{X: cellId.X + 1, Y: cellId.Y - 1})
			solveCollision(IntVector2{X: cellId.X + 1, Y: cellId.Y})
			solveCollision(IntVector2{X: cellId.X - 1, Y: cellId.Y + 1})
			solveCollision(IntVector2{X: cellId.X - 1, Y: cellId.Y - 1})
			solveCollision(IntVector2{X: cellId.X - 1, Y: cellId.Y})
			solveCollision(IntVector2{X: cellId.X, Y: cellId.Y + 1})
			solveCollision(IntVector2{X: cellId.X, Y: cellId.Y - 1})
		}
	})
}
