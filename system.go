package ecs

import (
	"fmt"
	"sync/atomic"
	"time"

	"runtime"
)

// Represents an individual system
type System struct {
	Name string
	Func func(dt time.Duration)
}

func (s System) Build(world *World) System {
	return s
}

// Create a new system. The system name will be automatically created based on the function name that calls this function
func NewSystem(lambda func(dt time.Duration)) System {
	systemName := "UnknownSystemName"

	pc, _, _, ok := runtime.Caller(1)
	if ok {
		details := runtime.FuncForPC(pc)
		systemName = details.Name()
	}

	return System{
		Name: systemName,
		Func: lambda,
	}
}

// Executes the system once, returning the time taken.
// This is mostly used by the scheduler, but you can use it too.
func (s *System) step(dt time.Duration) {
	// Note: Disable timing
	s.Func(dt)
	// return 0

	// fmt.Println(s.Name) // Spew

	// start := time.Now()
	// s.Func(dt)

	// return time.Since(start)
}

// A log of a system and the time it took to execute
type SystemLog struct {
	Name string
	Time time.Duration
}

func (s *SystemLog) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Time)
}

// // TODO - Just use an atomic here?
// type signal struct {
// 	mu    sync.Mutex
// 	value bool
// }

// func (s *signal) Set(val bool) {
// 	s.mu.Lock()
// 	s.value = val
// 	s.mu.Unlock()
// }

// func (s *signal) Get() bool {
// 	s.mu.Lock()
// 	ret := s.value
// 	s.mu.Unlock()
// 	return ret
// }

// Scheduler is a place to put your systems and have them run.
// There are two types of systems: Fixed time systems and dynamic time systems
// 1. Fixed time systems will execute on a fixed time step
// 2. Dynamic time systems will execute as quickly as they possibly can
// The scheduler may change in the future, but right now how it works is simple:
// Input: Execute input systems (Dynamic time systems)
// Physics: Execute physics systems (Fixed time systems)
// Render: Execute render systems (Dynamic time systems)
type Scheduler struct {
	world                             *World
	systems                           [][]System
	sysTimeFront, sysTimeBack         [][]SystemLog // Rotating log of how long each system takes
	stageTimingFront, stageTimingBack []SystemLog   // Rotating log of how long each stage takes

	fixedTimeStep time.Duration
	accumulator   time.Duration
	gameSpeed     float64
	quit          atomic.Bool
	pauseRender   atomic.Bool
	maxLoopCount  int
}

// Creates a scheduler
func NewScheduler(world *World) *Scheduler {
	return &Scheduler{
		world:        world,
		systems:      make([][]System, StageLast+1),
		sysTimeFront: make([][]SystemLog, StageLast+1),
		sysTimeBack:  make([][]SystemLog, StageLast+1),

		fixedTimeStep: 16 * time.Millisecond,
		accumulator:   0,
		gameSpeed:     1,
	}
}

// TODO make SetGameSpeed and SetFixedTimeStep thread safe.

// Sets the rate at which time accumulates. Also, you want them to only change at the end of a frame, else you might get some inconsistencies. Just use a mutex and a single temporary variable
func (s *Scheduler) SetGameSpeed(speed float64) {
	s.gameSpeed = speed
}

// Tells the scheduler to exit. Scheduler will finish executing its remaining tick before closing.
func (s *Scheduler) SetQuit(value bool) {
	s.quit.Store(true)
}

// Returns the quit value of the scheduler
func (s *Scheduler) Quit() bool {
	return s.quit.Load()
}

// Pauses the set of render systems (ie they will be skipped).
// Deprecated: This API is tentatitive
func (s *Scheduler) PauseRender(value bool) {
	s.pauseRender.Store(value)
}

// Sets the amount of time required before the fixed time systems will execute
func (s *Scheduler) SetFixedTimeStep(t time.Duration) {
	s.fixedTimeStep = t
}

type Stage uint8

func (s Stage) String() string {
	switch s {
	case StageStartup:
		return "StageStartup"
	case StagePreUpdate:
		return "StagePreUpdate"
	case StageFixedUpdate:
		return "StageFixedUpdate"
	case StageUpdate:
		return "StageUpdate"
	case StageLast:
		return "StageLast"
	}
	return "Unknown"
}

const (
	// StagePreStartup
	StageStartup Stage = iota
	// StagePostStartup
	// StageFirst
	StagePreUpdate // Note: Used to be Input
	// StageStateTransition
	StageFixedUpdate
	// StagePostFixedUpdate
	StageUpdate
	// StagePostUpdate
	StageLast
)

// Returns true if the scheduler only has fixed systems
func (s *Scheduler) isFixedOnly() bool {
	return len(s.systems[StagePreUpdate]) == 0 && len(s.systems[StageUpdate]) == 0 && len(s.systems[StageLast]) == 0
}

func (s *Scheduler) ClearSystems(stage Stage) {
	// Note: Make a new slices so that any of the old system pointers get released
	s.systems[stage] = make([]System, 0)
}

func (s *Scheduler) AddSystems(stage Stage, systems ...SystemBuilder) {
	for _, sys := range systems {
		system := sys.Build(s.world)
		s.systems[stage] = append(s.systems[stage], system)
	}
}

func (s *Scheduler) SetSystems(stage Stage, systems ...SystemBuilder) {
	s.ClearSystems(stage)
	s.AddSystems(stage, systems...)
}

// Sets the accumulator maximum point so that if the accumulator gets way to big, we will reset it and continue on, dropping all physics ticks that would have been executed. This is useful in a runtime like WASM where the browser may not let us run as frequently as we may need (for example, when the tab is hidden or minimized).
// Note: This must be set before you call scheduler.Run()
// Note: The default value is 0, which will force every physics tick to run. I highly recommend setting this to something if you plan to build for WASM!
func (s *Scheduler) SetMaxPhysicsLoopCount(count int) {
	s.maxLoopCount = count
}

func (s *Scheduler) Syslog(stage Stage) []SystemLog {
	return s.sysTimeFront[stage]
}

// Returns an interpolation value which represents how close we are to the next fixed time step execution. Can be useful for interpolating dynamic time systems to the fixed time systems. I might rename this
func (s *Scheduler) GetRenderInterp() float64 {
	return s.accumulator.Seconds() / s.fixedTimeStep.Seconds()
}

func (s *Scheduler) runUntrackedStage(stage Stage, dt time.Duration) {
	for _, sys := range s.systems[stage] {
		sys.step(dt)
		s.world.cmd.Execute()
	}
}

func (s *Scheduler) runStage(stage Stage, dt time.Duration) {
	start := time.Now()

	for _, sys := range s.systems[stage] {
		sysStart := time.Now()
		sys.step(dt)
		s.world.cmd.Execute()

		{
			tmp := s.sysTimeFront[stage]
			s.sysTimeFront[stage] = s.sysTimeBack[stage]
			s.sysTimeBack[stage] = tmp[:0]
		}
		s.sysTimeBack[stage] = append(s.sysTimeBack[stage], SystemLog{
			Name: sys.Name,
			Time: time.Since(sysStart),
		})
	}

	{
		tmp := s.stageTimingFront
		s.stageTimingFront = s.stageTimingBack
		s.stageTimingBack = tmp[:0]
	}
	s.stageTimingBack = append(s.stageTimingBack, SystemLog{
		Name: "STAGE NAME TODO",
		Time: time.Since(start),
	})
}

// Performs a single step of the scheduler with the provided time
func (s *Scheduler) Step(dt time.Duration) {
	// Pre Update
	s.runStage(StagePreUpdate, dt)

	maxLoopCount := time.Duration(s.maxLoopCount)
	if maxLoopCount > 0 {
		if s.accumulator > (maxLoopCount * s.fixedTimeStep) {
			s.accumulator = s.fixedTimeStep // Just run one loop
		}
	}

	// Physics Systems
	for s.accumulator >= s.fixedTimeStep {
		s.runStage(StageFixedUpdate, s.fixedTimeStep)
		s.accumulator -= s.fixedTimeStep
	}

	// Render Systems
	if !s.pauseRender.Load() {
		s.runStage(StageUpdate, dt)
	}
}

// Note: Would be nice to sleep or something to prevent spinning while we wait for work to do
// Could also separate the render loop from the physics loop (requires some thread safety in ECS)
func (s *Scheduler) Run() {
	s.runUntrackedStage(StageStartup, 0)

	frameStart := time.Now()
	dt := s.fixedTimeStep
	s.accumulator = 0

	for !s.quit.Load() {
		s.Step(dt)

		// Edge case for schedules only fixed time steps
		if s.isFixedOnly() {
			// Note: This is guaranteed to be positive because the physics execution loops until the accumulator is less than fixedtimestep
			time.Sleep(s.fixedTimeStep - s.accumulator)
		}

		// Capture Frame time
		now := time.Now()
		dt = now.Sub(frameStart)
		frameStart = now

		scaledDt := float64(dt.Nanoseconds()) * s.gameSpeed
		s.accumulator += time.Duration(scaledDt)
	}
}

// //Separates physics loop from render loop
// func (s *Scheduler) Run2() {
// 	var worldMu sync.Mutex

// 	frameStart := time.Now()
// 	dt := s.fixedTimeStep
// 	// var accumulator time.Duration
// 	s.accumulator = 0
// 	maxLoopCount := time.Duration(s.maxLoopCount)

// 	// physicsTicker := time.NewTicker(s.fixedTimeStep)
// 	// defer physicsTicker.Stop()
// 	go func() {
// 		// for {
// 		// 	_, more := <-physicsTicker.C
// 		// 	if !more { break } // Exit early, ticker channel is closed
// 		// 	// fmt.Println(phyTime)
// 		// 	worldMu.Lock()
// 		// 	for _, sys := range s.physics {
// 		// 		sys.Run(s.fixedTimeStep)
// 		// 	}
// 		// 	worldMu.Unlock()
// 		// }

// 		for !s.quit.Load() {
// 			worldMu.Lock()
// 			if maxLoopCount > 0 {
// 				if s.accumulator > (maxLoopCount * s.fixedTimeStep) {
// 					s.accumulator = s.fixedTimeStep // Just run one loop
// 				}
// 			}
// 			for s.accumulator >= s.fixedTimeStep {
// 				for _, sys := range s.physics {
// 					sys.Run(s.fixedTimeStep)
// 				}
// 				s.accumulator -= s.fixedTimeStep
// 			}

// 			worldMu.Unlock()
// 			time.Sleep(s.fixedTimeStep - s.accumulator)
// 		}
// 	}()

// 	for !s.quit.Load() {
// 		worldMu.Lock()

// 		for _, sys := range s.render {
// 			sys.Run(dt)
// 		}

// 		for _, sys := range s.input {
// 			sys.Run(dt)
// 		}

// 		// Capture Frame time
// 		now := time.Now()
// 		dt = now.Sub(frameStart)
// 		frameStart = now

// 		s.accumulator += dt
// 		worldMu.Unlock()
// 	}
// }
