package group

import (
	"sync"
	"time"

	"github.com/unitoftime/ecs/system"
)

type SystemStatistics struct {
	Name system.SystemName

	WaitingForOrderStarted            time.Time
	WaitingForComponentsAccessStarted time.Time
	ExecutionStarted                  time.Time
	ExecutionEnded                    time.Time
}

func (s *SystemStatistics) GetWaitingForOrderTime() time.Duration {
	return time.Duration(s.WaitingForComponentsAccessStarted.Nanosecond() - s.WaitingForOrderStarted.Nanosecond())
}

func (s *SystemStatistics) GetWaitingForComponentsAccessTime() time.Duration {
	return time.Duration(s.ExecutionStarted.Nanosecond() - s.WaitingForComponentsAccessStarted.Nanosecond())
}

func (s *SystemStatistics) GetExecutionTime() time.Duration {
	return time.Duration(s.ExecutionEnded.Nanosecond() - s.ExecutionStarted.Nanosecond())
}

type UpdateStatistics struct {
	BeforeUpdateHandlersStarted time.Time
	BeforeUpdateHandlersEnded   time.Time

	SystemsStatistics []SystemStatistics

	AfterUpdateHandlersStarted time.Time
	AfterUpdateHandlersEnded   time.Time
}

type Statistics interface {
	GetUpdates() []*UpdateStatistics
}

type statistics struct {
	maxUpdatesCount int
	updates         []*UpdateStatistics
	updatesLock     *sync.RWMutex
}

func (s *statistics) GetUpdates() []*UpdateStatistics {
	s.updatesLock.RLock()
	defer s.updatesLock.RUnlock()

	updates := make([]*UpdateStatistics, len(s.updates))
	copy(updates, s.updates)

	return updates
}

func (s *statistics) pushUpdate(update UpdateStatistics) {
	s.updatesLock.Lock()
	defer s.updatesLock.Unlock()

	s.updates = append(s.updates, &update)
	if len(s.updates) > s.maxUpdatesCount {
		updates := make([]*UpdateStatistics, len(s.updates)-1)
		copy(updates, s.updates[1:])
		s.updates = updates
	}
}

// TODO: add statistics aggregation for longer periods of time
// TODO: enable/disable statistics
