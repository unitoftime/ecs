package group

import (
	s "github.com/unitoftime/ecs/system"
)

type GroupError interface {
	error

	SystemName() s.SystemName
	StackTrace() string
}

type groupError struct {
	innerError error
	systemName s.SystemName
	trace      string
}

func (e *groupError) Error() string {
	return e.innerError.Error()
}

func (e *groupError) SystemName() s.SystemName {
	return e.systemName
}

func (e *groupError) StackTrace() string {
	return e.trace
}
