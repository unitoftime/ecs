package ecs

import (
	"fmt"
	"time"
	"sync"

	"runtime"
)

type System struct {
	Name string
	Func func(dt time.Duration)
}

func NewSystem(lambda func (dt time.Duration)) System {
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

func (s *System) Run(dt time.Duration) time.Duration {
	start := time.Now()
	s.Func(dt)
	return time.Since(start)
}

type SystemLog struct {
	Name string
	Time time.Duration
}

func (s *SystemLog) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Time)
}

// TODO - Just use an atomic here?
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
	sysLogBackFixed, sysLogFrontFixed []SystemLog
	fixedTimeStep time.Duration
	accumulator time.Duration
	gameSpeed int64
}
func NewScheduler() *Scheduler {
	return &Scheduler{
		input: make([]System, 0),
		physics: make([]System, 0),
		render: make([]System, 0),
		sysLogFront: make([]SystemLog, 0),
		sysLogBack: make([]SystemLog, 0),
		sysLogFrontFixed: make([]SystemLog, 0),
		sysLogBackFixed: make([]SystemLog, 0),
		fixedTimeStep: 16 * time.Millisecond,
		accumulator: 0,
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
func (s *Scheduler) Syslog() []SystemLog {
	return s.sysLogFront
}

// Returns the front syslog for fixed-dt systems only. Note: This is only valid for the current frame, you should call this every frame if you use it!
func (s *Scheduler) SyslogFixed() []SystemLog {
	return s.sysLogFrontFixed
}

func (s *Scheduler) GetRenderInterp() float64 {
	return s.accumulator.Seconds() / s.fixedTimeStep.Seconds()
}

// Note: Would be nice to sleep or something to prevent spinning while we wait for work to do
// Could also separate the render loop from the physics loop (requires some thread safety in ECS)
// TODO this doesn't work with vsync because the pause blocks the physics from decrementing the accumulator
func (s *Scheduler) Run(quit *Signal) {
	frameStart := time.Now()
	dt := s.fixedTimeStep
	// var accumulator time.Duration
	s.accumulator = 0

	for !quit.Get() {
		{
			tmpSysLog := s.sysLogFront
			s.sysLogFront = s.sysLogBack
			s.sysLogBack = tmpSysLog
			s.sysLogBack = s.sysLogBack[:0]
		}

		// Input Systems
		for _,sys := range s.input {
			sysTime := sys.Run(dt)

			s.sysLogBack = append(s.sysLogBack, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
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
			for _,sys := range s.physics {
				sysTime := sys.Run(s.fixedTimeStep)

				s.sysLogBackFixed = append(s.sysLogBackFixed, SystemLog{
					Name: sys.Name,
					Time: sysTime,
				})
			}
			s.accumulator -= s.fixedTimeStep
		}

		// Render Systems
		for _,sys := range s.render {
			sysTime := sys.Run(dt)

			s.sysLogBack = append(s.sysLogBack, SystemLog{
				Name: sys.Name,
				Time: sysTime,
			})
		}

		// Edge case for schedules only fixed time steps
		if len(s.input) == 0 && len(s.render) == 0 {
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
