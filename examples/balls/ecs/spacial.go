package ecs

import (
	"math"
	"time"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const SpacialCellSize = 2.0
const SpacialSystemName system.SystemName = "physics::spacial"

type SpacialCellObject struct {
	Entity ecs.Id
	X      float64
	Y      float64
}

type SpacialSystem struct {
	World       *ecs.World
	SpacialHash map[IntVector2][]SpacialCellObject
}

func (s *SpacialSystem) GetName() system.SystemName {
	return SpacialSystemName
}

func (s *SpacialSystem) GetReadComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Position{})}
}

func (s *SpacialSystem) GetRunAfter() []system.SystemName {
	return []system.SystemName{MovementSystemName}
}

func (s *SpacialSystem) RunFixed(delta time.Duration) {
	for iv := range s.SpacialHash {
		delete(s.SpacialHash, iv)
	}

	query := ecs.Query1[Position](s.World)
	query.MapId(func(id ecs.Id, pos *Position) {
		cellId := IntVector2{
			X: int32(math.Round(pos.X / SpacialCellSize)),
			Y: int32(math.Round(pos.Y / SpacialCellSize)),
		}
		if _, ok := s.SpacialHash[cellId]; !ok {
			s.SpacialHash[cellId] = []SpacialCellObject{}
		}

		cell := s.SpacialHash[cellId]
		cell = append(cell, SpacialCellObject{
			Entity: id,
			X:      pos.X,
			Y:      pos.Y,
		})
		s.SpacialHash[cellId] = cell
	})
}
