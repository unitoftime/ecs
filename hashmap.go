package ecs

import (
	"github.com/unitoftime/ecs/internal/intmap"
)

type intkey interface {
	// comparable
	~int | ~uint | ~int64 | ~uint64 | ~int32 | ~uint32 | ~int16 | ~uint16 | ~int8 | ~uint8 | ~uintptr
}

// This is useful for testing different map implementations in my workload
type internalMap[K intkey, V any] struct {
	inner *intmap.Map[K, V]
}

func newMap[K, V intkey](size int) *internalMap[K, V] {
	return &internalMap[K, V]{
		intmap.New[K, V](0),
	}
}
func (m *internalMap[K, V]) Len() int {
	return m.inner.Len()
}

func (m *internalMap[K, V]) Get(k K) (V, bool) {
	return m.inner.Get(k)
}

func (m *internalMap[K, V]) Put(k K, val V) {
	m.inner.Put(k, val)
}

func (m *internalMap[K, V]) Delete(k K) {
	m.inner.Del(k)
}

func (m *internalMap[K, V]) Has(k K) bool {
	_, has := m.inner.Get(k)
	return has
}
