package group

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	s "github.com/unitoftime/ecs/system"
)

type StepEventHandler func(step int32)

type StepGroup interface {
	Group

	StartStep()
	StopStep()
	NextStep()

	OnBeforeStep(handler StepEventHandler)
	OnAfterStep(handler StepEventHandler)
}

type stepSystemGroup struct {
	group

	stepLock       *semaphore.Weighted
	currentStep    atomic.Int32
	targetStep     atomic.Int32
	runnerContext  *context.Context
	runnerDone     context.CancelFunc
	runnerDoneWait sync.WaitGroup

	onBeforeStepHandlers []StepEventHandler
	onAfterStepHandlers  []StepEventHandler
}

func NewStepGroup(name string, componentsGuard ComponentsGuard) StepGroup {
	stepLock := semaphore.NewWeighted(9223372036854775807)
	err := stepLock.Acquire(context.Background(), 9223372036854775807)
	if err != nil {
		panic(err)
	}

	return &stepSystemGroup{
		group:                newGroup(name, componentsGuard),
		stepLock:             stepLock,
		currentStep:          atomic.Int32{},
		targetStep:           atomic.Int32{},
		runnerContext:        nil,
		runnerDoneWait:       sync.WaitGroup{},
		onBeforeStepHandlers: []StepEventHandler{},
		onAfterStepHandlers:  []StepEventHandler{},
	}
}

func (g *stepSystemGroup) runner(ctx context.Context) {
	defer g.runnerDoneWait.Done()

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
			g.stepLock.Acquire(ctx, 1)
			g.currentStep.Add(1)
		}

		beforeUpdateHandlersStartedTime := time.Now()
		for _, handler := range g.onBeforeStepHandlers {
			handler(g.currentStep.Load())
		}
		beforeUpdateHandlersEndedTime := time.Now()

		g.orderGuard.Reset()

		systemsUpdatesStatistics := []SystemStatistics{}
		systemsUpdatesStatisticsLock := sync.Mutex{}

		systemsCompleted := sync.WaitGroup{}
		for _, system := range g.systems {
			systemsCompleted.Add(1)

			go func(runnedSystem s.StepSystem) {
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
				runnedSystem.RunStep(g.currentStep.Load())
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
			}(system.(s.StepSystem))
		}

		systemsCompleted.Wait()

		afterUpdateHandlersStartedTime := time.Now()
		for _, handler := range g.onAfterStepHandlers {
			handler(g.currentStep.Load())
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

func (g *stepSystemGroup) StartStep() {
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
func (g *stepSystemGroup) StopStep() {
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
func (g *stepSystemGroup) NextStep() {
	g.stepLock.Release(1)
	g.targetStep.Add(1)
}

func (g *stepSystemGroup) AddSystem(system s.System) {
	_, isRealtime := system.(s.StepSystem)
	if !isRealtime {
		panic("system must implement StepSystem interface")
	}

	g.group.AddSystem(system)
}

func (g *stepSystemGroup) OnBeforeStep(handler StepEventHandler) {
	g.onBeforeStepHandlers = append(g.onBeforeStepHandlers, handler)
}

func (g *stepSystemGroup) OnAfterStep(handler StepEventHandler) {
	g.onAfterStepHandlers = append(g.onAfterStepHandlers, handler)
}
