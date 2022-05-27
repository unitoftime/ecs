package ecs

import (
	"time"
	"sync"
)

type System struct {
	Name string
	Func func(dt time.Duration)
}

func (s *System) Run(dt time.Duration) time.Duration {
	start := time.Now()
	s.Func(dt)
	return time.Since(start)
}

type SystemLog struct {
	Name string
	Time time.Duration
}

type Signal struct {
	mu sync.Mutex
	value bool
}

func (s *Signal) Set(val bool) {
	s.mu.Lock()
	s.value = val
	s.mu.Unlock()
}

func (s *Signal) Get() bool {
	s.mu.Lock()
	ret := s.value
	s.mu.Unlock()
	return ret
}

type Scheduler struct {
	input, physics, render []System
	sysLogBack, sysLogFront []SystemLog
	fixedTimeStep time.Duration
	gameSpeed int64
}
func NewScheduler() *Scheduler {
	return &Scheduler{
		input: make([]System, 0),
		physics: make([]System, 0),
		render: make([]System, 0),
		sysLogFront: make([]SystemLog, 0),
		sysLogBack: make([]SystemLog, 0),
		fixedTimeStep: 16 * time.Millisecond,
		gameSpeed: 1,
	}
}

// TODO make SetGameSpeed and SetFixedTimeStep thread safe. Also, you want them to only change at the end of a frame, else you might get some inconsistencies. Just use a mutex and a single temporary variable
func (s *Scheduler) SetGameSpeed(speed int64) {
	s.gameSpeed = speed
}

func (s *Scheduler) SetFixedTimeStep(t time.Duration) {
	s.fixedTimeStep = t
}

func (s *Scheduler) AppendInput(systems ...System) {
	s.input = append(s.input, systems...)
}

func (s *Scheduler) AppendPhysics(systems ...System) {
	s.physics = append(s.physics, systems...)
}

func (s *Scheduler) AppendRender(systems ...System) {
	s.render = append(s.render, systems...)
}

// Returns the front syslog so the user can analyze it. Note: This is only valid for the current frame, you should call this every frame if you use it!
func (s *Scheduler) SysLog() []SystemLog {
	return s.sysLogFront
}

// Note: Would be nice to sleep or something to prevent spinning while we wait for work to do
// Could also separate the render loop from the physics loop (requires some thread safety in ECS)
// TODO this doesn't work with vsync because the pause blocks the physics from decrementing the accumulator
func (s *Scheduler) Run(quit *Signal) {
	frameStart := time.Now()
	dt := s.fixedTimeStep
	var accumulator time.Duration

	for !quit.Get() {
		tmpSysLog := s.sysLogFront
		s.sysLogFront = s.sysLogBack
		s.sysLogBack = tmpSysLog

		s.sysLogBack = s.sysLogBack[:0]

		// Input Systems
		for _,sys := range s.input {
			sysTime := sys.Run(dt)

			s.sysLogBack = append(s.sysLogBack, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		// Physics Systems
		if accumulator >= s.fixedTimeStep {
			for _,sys := range s.physics {
				sysTime := sys.Run(s.fixedTimeStep)

				s.sysLogBack = append(s.sysLogBack, SystemLog{
					Name: sys.Name,
					Time: sysTime,
				})
			}
			accumulator -= s.fixedTimeStep
		}

		// Render Systems
		for _,sys := range s.render {
			sysTime := sys.Run(dt)

			s.sysLogBack = append(s.sysLogBack, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		// Capture Frame time
		dt = time.Since(frameStart)
		frameStart = time.Now()

		scaledDt := dt.Nanoseconds() * s.gameSpeed
		accumulator += time.Duration(scaledDt)
	}
}


const FixedTimeStep = 16 * time.Millisecond
const GameSpeed int64 = 2
// Note: Would be nice to sleep or something to prevent spinning while we wait for work to do
// Could also separate the render loop from the physics loop (requires some thread safety in ECS)
func RunGame(inputSystems, physicsSystems, renderSystems []System, quit *Signal) {
	frameStart := time.Now()
	dt := FixedTimeStep
	var accumulator time.Duration

	sysLog := make([]SystemLog, 0)

	for !quit.Get() {
		sysLog = sysLog[:0]

		// Input Systems
		for _,sys := range inputSystems {
			sysTime := sys.Run(dt)

			sysLog = append(sysLog, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		// Physics Systems
		if accumulator >= FixedTimeStep {
			for _,sys := range physicsSystems {
				sysTime := sys.Run(FixedTimeStep)

				sysLog = append(sysLog, SystemLog{
					Name: sys.Name,
					Time: sysTime,
				})
			}
			accumulator -= FixedTimeStep
		}

		// Render Systems
		for _,sys := range renderSystems {
			sysTime := sys.Run(dt)

			sysLog = append(sysLog, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		// Capture Frame time
		dt = time.Since(frameStart)
		frameStart = time.Now()

		scaledDt := dt.Nanoseconds() * GameSpeed
		accumulator += time.Duration(scaledDt)
	}
}

func RunGameFixed(physicsSystems []System, quit *Signal) {
	frameStart := time.Now()
	dt := FixedTimeStep

	for !quit.Get() {

		for _,sys := range physicsSystems {
			sys.Run(FixedTimeStep)
		}

		dt = time.Since(frameStart)
		time.Sleep(FixedTimeStep - dt)

		frameStart = time.Now()
	}
}
