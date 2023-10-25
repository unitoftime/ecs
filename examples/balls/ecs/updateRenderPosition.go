package ecs

import (
	"time"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const UpdateRenderPositionSystemName system.SystemName = "render::updatePosition"

type UpdateRenderPositionSystem struct {
	World *ecs.World
}

func (s *UpdateRenderPositionSystem) GetName() system.SystemName {
	return UpdateRenderPositionSystemName
}

func (s *UpdateRenderPositionSystem) GetReadComponents() []ecs.Component {
	return []ecs.Component{ecs.C(Position{})}
}

func (s *UpdateRenderPositionSystem) GetWriteComponents() []ecs.Component {
	return []ecs.Component{ecs.C(RenderPosition{})}
}

func (s *UpdateRenderPositionSystem) RunRealtime(delta time.Duration) {
	lerpMultiplier := delta.Seconds() * 10
	if lerpMultiplier >= 1 {
		lerpMultiplier = 1
	}

	query := ecs.Query2[Position, RenderPosition](s.World)
	query.MapId(func(id ecs.Id, pos *Position, rend *RenderPosition) {
		xDif := pos.X - rend.X
		rend.X = rend.X + xDif*lerpMultiplier
		yDif := pos.Y - rend.Y
		rend.Y = rend.Y + yDif*lerpMultiplier
	})
}
