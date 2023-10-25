package ecs

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

const RenderSystemName system.SystemName = "render::render"

type RenderSystem struct {
	World *ecs.World
}

func (s *RenderSystem) GetName() system.SystemName {
	return RenderSystemName
}

func (s *RenderSystem) GetReadComponents() []ecs.Component {
	return []ecs.Component{ecs.C(RenderPosition{})}
}

func (s *RenderSystem) RunRealtime(delta time.Duration) {
	query := ecs.Query1[RenderPosition](s.World)
	query.MapId(func(id ecs.Id, pos *RenderPosition) {
		rl.DrawCircle(int32(pos.X*10), int32(pos.Y*10), 10, rl.Blue)
	})
}
