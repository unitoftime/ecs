package ecs

import "reflect"

type EventId int

var eventRegistryCounter EventId = 0
var registeredEvents = make(map[reflect.Type]EventId, 0)

// This function is not thread safe
func NewEvent[T any]() EventId {
	var t T
	typeof := reflect.TypeOf(t)
	eventId, ok := registeredEvents[typeof]
	if !ok {
		eventId = eventRegistryCounter
		registeredEvents[typeof] = eventId
		eventRegistryCounter++
	}
	return eventId
}

type Event interface {
	EventId() EventId
}

type Trigger[T Event] struct {
	Id   Id // If set, it is the entity Id that this event was triggered on
	Data T
}

type Handler interface {
	Run(id Id, event any)
	EventTrigger() EventId
}

type handlerData[E Event] struct {
	lambda func(Trigger[E])
}

func (h handlerData[E]) Run(id Id, event any) {
	e := event.(E)
	trigger := Trigger[E]{id, e}
	h.lambda(trigger)
}

func (h handlerData[E]) EventTrigger() EventId {
	var e E
	return e.EventId()
}

func NewHandler[E Event](f func(trigger Trigger[E])) handlerData[E] {
	return handlerData[E]{
		lambda: f,
	}
}
