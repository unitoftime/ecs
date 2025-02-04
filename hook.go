package ecs

type OnAdd struct {
	compId CompId
}

var _onAddId = NewEvent[OnAdd]()

func (p OnAdd) EventId() EventId {
	return _onAddId
}
