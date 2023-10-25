package group

import (
	"context"
	"sync"
	"time"

	s "github.com/unitoftime/ecs/system"
)

type RealtimeUpdateEventHandler func(delta time.Duration)

type RealtimeGroup interface {
	Group

	RunRealtime(delta time.Duration)
	StartRealtime()
	StopRealtime()

	OnBeforeUpdate(handler RealtimeUpdateEventHandler)
	OnAfterUpdate(handler RealtimeUpdateEventHandler)
}

type realtimeSystemGroup struct {
	group

	runnerContext  *context.Context
	runnerDone     context.CancelFunc
	runnerDoneWait sync.WaitGroup

	onBeforeUpdateHandlers []RealtimeUpdateEventHandler
	onAfterUpdateHandlers  []RealtimeUpdateEventHandler
}

func NewRealtimeGroup(name string, componentsGuard ComponentsGuard) RealtimeGroup {
	return &realtimeSystemGroup{
		group:                  newGroup(name, componentsGuard),
		runnerContext:          nil,
		runnerDoneWait:         sync.WaitGroup{},
		onBeforeUpdateHandlers: []RealtimeUpdateEventHandler{},
		onAfterUpdateHandlers:  []RealtimeUpdateEventHandler{},
	}
}

func (g *realtimeSystemGroup) RunRealtime(delta time.Duration) {
	beforeUpdateHandlersStartedTime := time.Now()
	for _, handler := range g.onBeforeUpdateHandlers {
		handler(delta)
	}
	beforeUpdateHandlersEndedTime := time.Now()

	g.orderGuard.Reset()

	systemsUpdatesStatistics := []SystemStatistics{}
	systemsUpdatesStatisticsLock := sync.Mutex{}

	systemsCompleted := sync.WaitGroup{}
	for _, system := range g.systems {
		systemsCompleted.Add(1)

		go func(runnedSystem s.RealtimeSystem) {
			// Handle panics of the running system
			defer func() {
				if err := recover(); err != nil {
					g.notifyError(err.(error), runnedSystem.(s.System))
				}
			}()

			defer systemsCompleted.Done()

			waitingForOrderStartedTime := time.Now()
			g.orderGuard.Lock(runnedSystem.(s.System))
			defer g.orderGuard.Release(runnedSystem.(s.System))
			waitingForComponentsAccessTime := time.Now()
			g.componentsGuard.Lock(runnedSystem.(s.System))
			defer g.componentsGuard.Release(runnedSystem.(s.System))

			executionStartedTime := time.Now()
			runnedSystem.RunRealtime(delta)
			executionEndedTime := time.Now()

			systemsUpdatesStatisticsLock.Lock()
			defer systemsUpdatesStatisticsLock.Unlock()
			systemsUpdatesStatistics = append(systemsUpdatesStatistics, SystemStatistics{
				Name:                              runnedSystem.(s.System).GetName(),
				WaitingForOrderStarted:            waitingForOrderStartedTime,
				WaitingForComponentsAccessStarted: waitingForComponentsAccessTime,
				ExecutionStarted:                  executionStartedTime,
				ExecutionEnded:                    executionEndedTime,
			})
		}(system.(s.RealtimeSystem))
	}

	systemsCompleted.Wait()

	afterUpdateHandlersStartedTime := time.Now()
	for _, handler := range g.onAfterUpdateHandlers {
		handler(delta)
	}
	afterUpdateHandlersEndedTime := time.Now()

	g.statistics.pushUpdate(UpdateStatistics{
		BeforeUpdateHandlersStarted: beforeUpdateHandlersStartedTime,
		BeforeUpdateHandlersEnded:   beforeUpdateHandlersEndedTime,
		SystemsStatistics:           systemsUpdatesStatistics,
		AfterUpdateHandlersStarted:  afterUpdateHandlersStartedTime,
		AfterUpdateHandlersEnded:    afterUpdateHandlersEndedTime,
	})
}

func (g *realtimeSystemGroup) runner(ctx context.Context) {
	defer g.runnerDoneWait.Done()
	lastRunTime := time.Now()

	// Handle panics of update handlers
	defer func() {
		if err := recover(); err != nil {
			g.notifyError(err.(error), nil)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		newRunTime := time.Now()
		delta := time.Duration(newRunTime.UnixNano() - lastRunTime.UnixNano())
		lastRunTime = newRunTime

		g.RunRealtime(delta)
	}
}

func (g *realtimeSystemGroup) StartRealtime() {
	if g.runnerContext != nil {
		return
	}

	for _, startHandler := range g.onStartHandlers {
		startHandler()
	}

	ctx, cancell := context.WithCancel(context.Background())
	g.runnerContext = &ctx
	g.runnerDone = cancell
	g.runnerDoneWait.Add(1)
	go g.runner(ctx)
}
func (g *realtimeSystemGroup) StopRealtime() {
	if g.runnerContext == nil {
		return
	}

	g.runnerDone()
	g.runnerContext = nil
	g.runnerDoneWait.Wait()

	for _, stopHandler := range g.onStopHandlers {
		stopHandler()
	}
}

func (g *realtimeSystemGroup) AddSystem(system s.System) {
	_, isRealtime := system.(s.RealtimeSystem)
	if !isRealtime {
		panic("system must implement RealtimeSystem interface")
	}

	g.group.AddSystem(system)
}

func (g *realtimeSystemGroup) OnBeforeUpdate(handler RealtimeUpdateEventHandler) {
	g.onBeforeUpdateHandlers = append(g.onBeforeUpdateHandlers, handler)
}

func (g *realtimeSystemGroup) OnAfterUpdate(handler RealtimeUpdateEventHandler) {
	g.onAfterUpdateHandlers = append(g.onAfterUpdateHandlers, handler)
}
