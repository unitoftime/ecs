package group

import (
	"context"
	"sync"
	"time"

	s "github.com/unitoftime/ecs/system"
)

type FixedUpdateEventHandler func(delta time.Duration)

type FixedSystemGroup interface {
	Group

	StartFixed()
	StopFixed()

	OnBeforeUpdate(handler FixedUpdateEventHandler)
	OnAfterUpdate(handler FixedUpdateEventHandler)
}

type fixedSystemGroup struct {
	group

	fixedTimeStep time.Duration

	runnerContext  *context.Context
	runnerDone     context.CancelFunc
	runnerDoneWait sync.WaitGroup

	onBeforeUpdateHandlers []FixedUpdateEventHandler
	onAfterUpdateHandlers  []FixedUpdateEventHandler
}

func NewFixedSystemGroup(name string, fixedTimeStep time.Duration, componentsGuard ComponentsGuard) FixedSystemGroup {
	return &fixedSystemGroup{
		group:                  newGroup(name, componentsGuard),
		fixedTimeStep:          fixedTimeStep,
		runnerContext:          nil,
		runnerDoneWait:         sync.WaitGroup{},
		onBeforeUpdateHandlers: []FixedUpdateEventHandler{},
		onAfterUpdateHandlers:  []FixedUpdateEventHandler{},
	}
}

func (g *fixedSystemGroup) runner(ctx context.Context) {
	defer g.runnerDoneWait.Done()
	lastRunTime := time.Now()

	// Handle panics of update handlers
	defer func() {
		if err := recover(); err != nil {
			g.notifyError(err.(error), nil)
		}
	}()

	for {
		timeToWait := time.Now().UnixNano() - lastRunTime.UnixNano() - int64(g.fixedTimeStep)
		if timeToWait > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Nanosecond * time.Duration(timeToWait)):
			}
		} else {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		newRunTime := time.Now()
		delta := time.Duration(newRunTime.UnixNano() - lastRunTime.UnixNano())
		lastRunTime = newRunTime

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
}

func (g *fixedSystemGroup) StartFixed() {
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
func (g *fixedSystemGroup) StopFixed() {
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

func (g *fixedSystemGroup) AddSystem(system s.System) {
	_, isRealtime := system.(s.FixedSystem)
	if !isRealtime {
		panic("system must implement FixedSystem interface")
	}

	g.group.AddSystem(system)
}

func (g *fixedSystemGroup) OnBeforeUpdate(handler FixedUpdateEventHandler) {
	g.onBeforeUpdateHandlers = append(g.onBeforeUpdateHandlers, handler)
}

func (g *fixedSystemGroup) OnAfterUpdate(handler FixedUpdateEventHandler) {
	g.onAfterUpdateHandlers = append(g.onAfterUpdateHandlers, handler)
}
