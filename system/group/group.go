package group

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	s "github.com/unitoftime/ecs/system"
)

type GroupErrorHandler func(err GroupError)

type Group interface {
	GetName() string

	Build()

	GetSystem(name s.SystemName) s.System
	GetAllSystems() []s.System
	AddSystem(system s.System)

	SetLogger(logger *slog.Logger)

	// Run provided function before first update
	OnStart(handler func())
	// Run provided function after group was stopped
	OnStop(handler func())

	OnError(handler GroupErrorHandler)

	GetStatistics() Statistics
}

type group struct {
	name string

	systems    []s.System
	systemsMap map[s.SystemName]s.System

	orderGuard      orderGuard
	componentsGuard ComponentsGuard
	logger          *slog.Logger

	onStartHandlers []func()
	onStopHandlers  []func()
	onErrorHandlers []GroupErrorHandler

	statistics statistics
}

func newGroup(name string, componentsGuard ComponentsGuard) group {
	return group{
		name:            name,
		systems:         []s.System{},
		systemsMap:      map[s.SystemName]s.System{},
		componentsGuard: componentsGuard,
		logger:          slog.Default().With(slog.String("systemsGroup", name)),
		onStartHandlers: []func(){},
		onStopHandlers:  []func(){},
		onErrorHandlers: []GroupErrorHandler{},
		statistics:      statistics{maxUpdatesCount: 10, updates: []*UpdateStatistics{}, updatesLock: &sync.RWMutex{}},
	}
}

func (g *group) Build() {
	g.orderGuard = newOrderGuard(g.systems)
	g.componentsGuard.InitForGroup(g)

	// TODO: Check for stable system execution flow if required. For example two systems need write access to the same component, but there is no order established between them. That way system that will accuire lock firts, will modify resources first. It can be a problem for games where stable simulations are required.
}

func (g *group) GetName() string {
	return g.name
}

func (g *group) GetSystem(name s.SystemName) s.System {
	if system, ok := g.systemsMap[name]; ok {
		return system
	}
	return nil
}

func (g *group) GetAllSystems() []s.System {
	return g.systems
}

func (g *group) AddSystem(system s.System) {
	_, isFixed := system.(s.FixedSystem)
	_, isRealtime := system.(s.RealtimeSystem)
	_, isStep := system.(s.StepSystem)
	if !isFixed && !isRealtime && !isStep {
		panic("system must implement FixedSystem, RealtimeSystem or StepSystem interface")
	}

	if _, ok := g.systemsMap[system.GetName()]; ok {
		panic(fmt.Sprintf("system with name [%s] already exist", system.GetName()))
	}
	g.systemsMap[system.GetName()] = system

	g.systems = append(g.systems, system)
}

func (g *group) SetLogger(logger *slog.Logger) {
	g.logger = logger.With(slog.String("systemsGroup", g.name))
}

func (g *group) notifyError(err error, system s.System) {
	systemName := s.SystemName("")
	if system != nil {
		systemName = system.GetName()
	}

	stackTrace := string(debug.Stack())

	g.logger.Error("Error while running system", "error", err, "trace", stackTrace, "system", string(systemName))
	for _, handler := range g.onErrorHandlers {
		// Handle handle panic to be shure that all handlers will be executed andprint trace of of the error
		defer func() {
			if herr := recover(); herr != nil {
				g.logger.Error("Panic while trying to execute panic handler", "error", err, "trace", stackTrace, "system", string(systemName), "handlerError", herr, "handlerTrace", string(debug.Stack()))
			}
		}()

		handler(&groupError{
			innerError: err,
			systemName: systemName,
			trace:      stackTrace,
		})
	}
}

func (g *group) OnStart(handler func()) {
	g.onStartHandlers = append(g.onStartHandlers, handler)
}
func (g *group) OnStop(handler func()) {
	g.onStopHandlers = append(g.onStopHandlers, handler)
}

func (g *group) OnError(handler GroupErrorHandler) {
	g.onErrorHandlers = append(g.onErrorHandlers, handler)
}

func (g *group) GetStatistics() Statistics {
	return &g.statistics
}
