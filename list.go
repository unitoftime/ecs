package ecs

type list[T any] struct {
	list []T
}

func newList[T any]() list[T] {
	return list[T]{
		list: make([]T, 0),
	}
}
func (l *list[T]) Add(t T) {
	l.list = append(l.list, t)
}
