package ecs

import (
	"fmt"
	"sync"
	"time"

	"runtime"
)

// Represents an individual system
type System struct {
	Name string
	Func func(dt time.Duration)
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
func (s *System) Run(dt time.Duration) time.Duration {
	start := time.Now()
	s.Func(dt)
	return time.Since(start)
}

// A log of a system and the time it took to execute
type SystemLog struct {
	Name string
	Time time.Duration
}

func (s *SystemLog) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Time)
}

// TODO - Just use an atomic here?
type signal struct {
	mu    sync.Mutex
	value bool
}

func (s *signal) Set(val bool) {
	s.mu.Lock()
	s.value = val
	s.mu.Unlock()
}

func (s *signal) Get() bool {
	s.mu.Lock()
	ret := s.value
	s.mu.Unlock()
	return ret
}

// Scheduler is a place to put your systems and have them run.
// There are two types of systems: Fixed time systems and dynamic time systems
// 1. Fixed time systems will execute on a fixed time step
// 2. Dynamic time systems will execute as quickly as they possibly can
// The scheduler may change in the future, but right now how it works is simple:
// Input: Execute input systems (Dynamic time systems)
// Physics: Execute physics systems (Fixed time systems)
// Render: Execute render systems (Dynamic time systems)
type Scheduler struct {
	input, physics, render            []System
	sysLogBack, sysLogFront           []SystemLog
	sysLogBackFixed, sysLogFrontFixed []SystemLog
	fixedTimeStep                     time.Duration
	accumulator                       time.Duration
	gameSpeed                         int64
	quit                              signal
	pauseRender                       signal
	maxLoopCount                      int
}

// Creates a scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		input:            make([]System, 0),
		physics:          make([]System, 0),
		render:           make([]System, 0),
		sysLogFront:      make([]SystemLog, 0),
		sysLogBack:       make([]SystemLog, 0),
		sysLogFrontFixed: make([]SystemLog, 0),
		sysLogBackFixed:  make([]SystemLog, 0),
		fixedTimeStep:    16 * time.Millisecond,
		accumulator:      0,
		gameSpeed:        1,
	}
}

// TODO make SetGameSpeed and SetFixedTimeStep thread safe.

// Sets the rate at which time accumulates. Also, you want them to only change at the end of a frame, else you might get some inconsistencies. Just use a mutex and a single temporary variable
func (s *Scheduler) SetGameSpeed(speed int64) {
	s.gameSpeed = speed
}

// Tells the scheduler to exit. Scheduler will finish executing its remaining tick before closing.
func (s *Scheduler) SetQuit(value bool) {
	s.quit.Set(true)
}

// Pauses the set of render systems (ie they will be skipped).
// Deprecated: This API is tentatitive
func (s *Scheduler) PauseRender(value bool) {
	s.pauseRender.Set(value)
}

// Sets the amount of time required before the fixed time systems will execute
func (s *Scheduler) SetFixedTimeStep(t time.Duration) {
	s.fixedTimeStep = t
}

// Adds a system to the list of input systems
func (s *Scheduler) AppendInput(systems ...System) {
	s.input = append(s.input, systems...)
}

// Adds a system to the list of physics systems
func (s *Scheduler) AppendPhysics(systems ...System) {
	s.physics = append(s.physics, systems...)
}

// Adds a system to the list of render systems
func (s *Scheduler) AppendRender(systems ...System) {
	s.render = append(s.render, systems...)
}

// Sets the accumulator maximum point so that if the accumulator gets way to big, we will reset it and continue on, dropping all physics ticks that would have been executed. This is useful in a runtime like WASM where the browser may not let us run as frequently as we may need (for example, when the tab is hidden or minimized).
// Note: This must be set before you call scheduler.Run()
// Note: The default value is 0, which will force every physics tick to run. I highly recommend setting this to something if you plan to build for WASM!
func (s *Scheduler) SetMaxPhysicsLoopCount(count int) {
	s.maxLoopCount = count
}

// Returns the front syslog so the user can analyze it. Note: This is only valid for the current frame, you should call this every frame if you use it!
func (s *Scheduler) Syslog() []SystemLog {
	return s.sysLogFront
}

// Returns the front syslog for fixed-dt systems only. Note: This is only valid for the current frame, you should call this every frame if you use it!
func (s *Scheduler) SyslogFixed() []SystemLog {
	return s.sysLogFrontFixed
}

// Returns an interpolation value which represents how close we are to the next fixed time step execution. Can be useful for interpolating dynamic time systems to the fixed time systems. I might rename this
func (s *Scheduler) GetRenderInterp() float64 {
	return s.accumulator.Seconds() / s.fixedTimeStep.Seconds()
}

// Note: Would be nice to sleep or something to prevent spinning while we wait for work to do
// Could also separate the render loop from the physics loop (requires some thread safety in ECS)
func (s *Scheduler) Run() {
	frameStart := time.Now()
	dt := s.fixedTimeStep
	// var accumulator time.Duration
	s.accumulator = 0
	maxLoopCount := time.Duration(s.maxLoopCount)

	for !s.quit.Get() {
		{
			tmpSysLog := s.sysLogFront
			s.sysLogFront = s.sysLogBack
			s.sysLogBack = tmpSysLog
			s.sysLogBack = s.sysLogBack[:0]
		}

		// Input Systems
		for _, sys := range s.input {
			sysTime := sys.Run(dt)

			s.sysLogBack = append(s.sysLogBack, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		if maxLoopCount > 0 {
			if s.accumulator > (maxLoopCount * s.fixedTimeStep) {
				s.accumulator = 0
			}
		}

		// TODO - If we get a double run, then all are accumulated
		if s.accumulator >= s.fixedTimeStep {
			tmpSysLog := s.sysLogFrontFixed
			s.sysLogFrontFixed = s.sysLogBackFixed
			s.sysLogBackFixed = tmpSysLog
			s.sysLogBackFixed = s.sysLogBackFixed[:0]
		}
		// Physics Systems
		for s.accumulator >= s.fixedTimeStep {
			for _, sys := range s.physics {
				sysTime := sys.Run(s.fixedTimeStep)

				s.sysLogBackFixed = append(s.sysLogBackFixed, SystemLog{
					Name: sys.Name,
					Time: sysTime,
				})
			}
			s.accumulator -= s.fixedTimeStep
		}

		// Render Systems
		if !s.pauseRender.Get() {
			for _, sys := range s.render {
				sysTime := sys.Run(dt)

				s.sysLogBack = append(s.sysLogBack, SystemLog{
					Name: sys.Name,
					Time: sysTime,
				})
			}
		}

		// Edge case for schedules only fixed time steps
		if len(s.input) == 0 && len(s.render) == 0 {
			// Note: This is guaranteed to be positive because the physics execution loops until the accumulator is less than fixedtimestep
			time.Sleep(s.fixedTimeStep - s.accumulator)
		}

		// Capture Frame time
		dt = time.Since(frameStart)
		frameStart = time.Now()

		scaledDt := dt.Nanoseconds() * s.gameSpeed
		s.accumulator += time.Duration(scaledDt)
		// fmt.Println(dt, s.accumulator)
	}
}

// // TODO! - Helpful starting point of commands? Maybe pass a commandlist to systems with dt as they execute. Maybe wrap dt and commandlist inside some general thing that gets passed to systems
// type Command struct {
// 	Id ecs.Id // If Id is ecs.InvalidEntity, we will spawn this as a new entity
// 	Entity *ecs.Entity
// }

// type CommandList struct {
// 	world *ecs.World
// 	list []Command
// }
// func NewCommandList(world *ecs.World) *CommandList {
// 	return &CommandList{
// 		world: world,
// 		list: make([]Command, 0),
// 	}
// }

// func (l *CommandList) Add(c Command) {
// 	l.list = append(l.list, c)
// }

// func (l *CommandList) Map(lambda func(c *Command)) {
// 	for i := range l.list {
// 		lambda(&l.list[i])
// 	}
// }

// func (l *CommandList) Execute() {
// 	for _, c := range l.list {
// 		id := c.Id
// 		if id == ecs.InvalidEntity {
// 			id = l.world.NewId()
// 		}

// 		ecs.WriteEntity(l.world, id, c.Entity)
// 	}
// }
