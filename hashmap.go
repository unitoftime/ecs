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

func newMap[K intkey, V any](size int) *internalMap[K, V] {
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

//--------------------------------------------------------------------------------
// TODO: Move to generational Ids

// This is useful for testing different map implementations in my workload
type locMap struct {
	// inner *LocMapImpl
	inner *intmap.Map[Id, entLoc]
}

func newLocMap(size int) locMap {
	return locMap{
		// NewLocMapImpl(size),
		intmap.New[Id, entLoc](0),
	}
}
func (m *locMap) Len() int {
	return m.inner.Len()
}

func (m *locMap) Get(k Id) (entLoc, bool) {
	return m.inner.Get(k)
}

func (m *locMap) Put(k Id, val entLoc) {
	m.inner.Put(k, val)
}

func (m *locMap) Delete(k Id) {
	m.inner.Del(k)
}

func (m *locMap) Has(k Id) bool {
	_, has := m.inner.Get(k)
	return has
}

// --------------------------------------------------------------------------------
// const fillFactor64 = 0.5

// // Hashing Reference: https://gist.github.com/badboy/6267743
// func phiMix64(x int) int {
// 	// Note: With this, we are only just a bit faster than swissmap
// 	h := x * (-1_640_531_527) // This is just the int32 version of the 0x9E3779B9
// 	return h ^ (h >> 16)

// 	// // TODO: track collision counts and compare before enabling this
// 	// // Theory: Because ecs.Id is just incremented by 1 each time, it might be effective to just always take the next slot
// 	// return x + x
// }

// type locPair struct {
// 	K Id
// 	V entLoc
// }

// // LocMapImpl is a hashmap where the keys are some any integer type.
// type LocMapImpl struct {
// 	data []locPair
// 	size int

// 	zeroVal    entLoc // value of 'zero' key
// 	hasZeroKey bool   // do we have 'zero' key in the map?
// }

// // New creates a new map with keys being any integer subtype.
// // The map can store up to the given capacity before reallocation and rehashing occurs.
// func NewLocMapImpl(capacity int) *LocMapImpl {
// 	return &LocMapImpl{
// 		data: make([]locPair, arraySize(capacity, fillFactor64)),
// 	}
// }

// // Get returns the value if the key is found.
// func (m *LocMapImpl) Get(key Id) (entLoc, bool) {
// 	if key == InvalidEntity {
// 		if m.hasZeroKey {
// 			return m.zeroVal, true
// 		}
// 		var zero entLoc
// 		return zero, false
// 	}

// 	idx := m.startIndex(key)
// 	p := m.data[idx]

// 	if p.K == InvalidEntity { // end of chain already
// 		var zero entLoc
// 		return zero, false
// 	}
// 	if p.K == key { // we check zero prior to this call
// 		return p.V, true
// 	}

// 	// hash collision, seek next hash match, bailing on first empty
// 	for {
// 		idx = m.nextIndex(idx)
// 		p = m.data[idx]
// 		if p.K == InvalidEntity {
// 			var zero entLoc
// 			return zero, false
// 		}
// 		if p.K == key {
// 			return p.V, true
// 		}
// 	}
// }

// // Put adds or updates key with value val.
// func (m *LocMapImpl) Put(key Id, val entLoc) {
// 	if key == InvalidEntity {
// 		if !m.hasZeroKey {
// 			m.size++
// 		}
// 		m.zeroVal = val
// 		m.hasZeroKey = true
// 		return
// 	}

// 	idx := m.startIndex(key)
// 	p := &m.data[idx]

// 	if p.K == InvalidEntity { // end of chain already
// 		p.K = key
// 		p.V = val
// 		if m.size >= m.sizeThreshold() {
// 			m.rehash()
// 		} else {
// 			m.size++
// 		}
// 		return
// 	} else if p.K == key { // overwrite existing value
// 		p.V = val
// 		return
// 	}

// 	// hash collision, seek next empty or key match
// 	for {
// 		idx = m.nextIndex(idx)
// 		p = &m.data[idx]

// 		if p.K == InvalidEntity {
// 			p.K = key
// 			p.V = val
// 			if m.size >= m.sizeThreshold() {
// 				m.rehash()
// 			} else {
// 				m.size++
// 			}
// 			return
// 		} else if p.K == key {
// 			p.V = val
// 			return
// 		}
// 	}
// }

// // Clear removes all items from the map, but keeps the internal buffers for reuse.
// func (m *LocMapImpl) Clear() {
// 	var zero entLoc
// 	m.hasZeroKey = false
// 	m.zeroVal = zero

// 	// compiles down to runtime.memclr()
// 	for i := range m.data {
// 		m.data[i] = locPair{}
// 	}

// 	m.size = 0
// }

// func (m *LocMapImpl) rehash() {
// 	oldData := m.data
// 	m.data = make([]locPair, 2*len(m.data))

// 	// reset size
// 	if m.hasZeroKey {
// 		m.size = 1
// 	} else {
// 		m.size = 0
// 	}

// 	forEach64(oldData, m.Put)
// 	// for _, p := range oldData {
// 	// 	if p.K != InvalidEntity {
// 	// 		m.Put(p.K, p.V)
// 	// 	}
// 	// }
// }

// // Len returns the number of elements in the map.
// func (m *LocMapImpl) Len() int {
// 	return m.size
// }

// func (m *LocMapImpl) sizeThreshold() int {
// 	return int(math.Floor(float64(len(m.data)) * fillFactor64))
// }

// func (m *LocMapImpl) startIndex(key Id) int {
// 	return phiMix64(int(key)) & (len(m.data) - 1)
// }

// func (m *LocMapImpl) nextIndex(idx int) int {
// 	return (idx + 1) & (len(m.data) - 1)
// }

// func forEach64(pairs []locPair, f func(k Id, v entLoc)) {
// 	for _, p := range pairs {
// 		if p.K != InvalidEntity {
// 			f(p.K, p.V)
// 		}
// 	}
// }

// // Del deletes a key and its value, returning true iff the key was found
// func (m *LocMapImpl) Del(key Id) bool {
// 	if key == InvalidEntity {
// 		if m.hasZeroKey {
// 			m.hasZeroKey = false
// 			m.size--
// 			return true
// 		}
// 		return false
// 	}

// 	idx := m.startIndex(key)
// 	p := m.data[idx]

// 	if p.K == key {
// 		// any keys that were pushed back needs to be shifted nack into the empty slot
// 		// to avoid breaking the chain
// 		m.shiftKeys(idx)
// 		m.size--
// 		return true
// 	} else if p.K == InvalidEntity { // end of chain already
// 		return false
// 	}

// 	for {
// 		idx = m.nextIndex(idx)
// 		p = m.data[idx]

// 		if p.K == key {
// 			// any keys that were pushed back needs to be shifted nack into the empty slot
// 			// to avoid breaking the chain
// 			m.shiftKeys(idx)
// 			m.size--
// 			return true
// 		} else if p.K == InvalidEntity {
// 			return false
// 		}

// 	}
// }

// func (m *LocMapImpl) shiftKeys(idx int) int {
// 	// Shift entries with the same hash.
// 	// We need to do this on deletion to ensure we don't have zeroes in the hash chain
// 	for {
// 		var p locPair
// 		lastIdx := idx
// 		idx = m.nextIndex(idx)
// 		for {
// 			p = m.data[idx]
// 			if p.K == InvalidEntity {
// 				m.data[lastIdx] = locPair{}
// 				return lastIdx
// 			}

// 			slot := m.startIndex(p.K)
// 			if lastIdx <= idx {
// 				if lastIdx >= slot || slot > idx {
// 					break
// 				}
// 			} else {
// 				if lastIdx >= slot && slot > idx {
// 					break
// 				}
// 			}
// 			idx = m.nextIndex(idx)
// 		}
// 		m.data[lastIdx] = p
// 	}
// }

// func nextPowerOf2(x uint32) uint32 {
// 	if x == math.MaxUint32 {
// 		return x
// 	}

// 	if x == 0 {
// 		return 1
// 	}

// 	x--
// 	x |= x >> 1
// 	x |= x >> 2
// 	x |= x >> 4
// 	x |= x >> 8
// 	x |= x >> 16

// 	return x + 1
// }

// func arraySize(exp int, fill float64) int {
// 	s := nextPowerOf2(uint32(math.Ceil(float64(exp) / fill)))
// 	if s < 2 {
// 		s = 2
// 	}
// 	return int(s)
// }
