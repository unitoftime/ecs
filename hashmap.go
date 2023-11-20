package ecs

import (
	// "github.com/dolthub/swiss"
	// "github.com/brentp/intintmap"
	// "github.com/unitoftime/intmap"
	"github.com/unitoftime/ecs/internal/intmap"
)

type intkey interface {
	// comparable
	~int | ~uint | ~int64 | ~uint64 | ~int32 | ~uint32 | ~int16 | ~uint16 | ~int8 | ~uint8 | ~uintptr
}

// This is useful for testing different map implementations in my workload
type internalMap[K intkey, V any] struct {
	// inner map[K]V
	// inner *swiss.Map[K, V]
	inner *intmap.Map[K, V]
}

func newMap[K,V intkey](size int) *internalMap[K,V] {
	return &internalMap[K,V]{
		// make(map[K]V),
		// swiss.NewMap[K, V](0), // Swissmap
		intmap.New[K, V](0),
	}
}
func (m *internalMap[K,V]) Len() int {
	return m.inner.Len()
}

func (m *internalMap[K,V]) Get(k K) (V, bool) {
	// v,ok := m.inner[k]
	// return v, ok
	return m.inner.Get(k)
}

func (m *internalMap[K,V]) Put(k K, val V) {
	// m.inner[k] = val
	m.inner.Put(k, val)
}

func (m *internalMap[K,V]) Delete(k K) {
	// delete(m.inner, k)
	// m.inner.Delete(k)
	m.inner.Del(k)
}

func (m *internalMap[K,V]) Has(k K) bool {
	// _, has := m.inner[k]
	// return m.inner.Has(k)
	_, has := m.inner.Get(k)
	return has
}
