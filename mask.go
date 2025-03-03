package ecs

import "fmt"

// Note: you can increase max component size by increasing maxComponentId and archetypeMask
// TODO: I should have some kind of panic if you go over maximum component size
const numMaskBlocks = 4
const maxComponentId = (numMaskBlocks * 64) - 1 // 4 maskBlocks = 255 components

var blankArchMask archetypeMask

// Supports maximum 256 unique component types
type archetypeMask [numMaskBlocks]uint64 // TODO: can/should I make this configurable?
func (a archetypeMask) String() string {
	return fmt.Sprintf("0x%x%x%x%x", a[0], a[1], a[2], a[3])
}

func buildArchMask(comps ...Component) archetypeMask {
	var mask archetypeMask
	for _, comp := range comps {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		c := comp.CompId()
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}
func buildArchMaskFromAny(comps ...any) archetypeMask {
	var mask archetypeMask
	for _, comp := range comps {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		c := name(comp)
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}
func buildArchMaskFromId(compIds ...CompId) archetypeMask {
	var mask archetypeMask
	for _, c := range compIds {
		// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
		idx := c / 64
		offset := c - (64 * idx)
		mask[idx] |= (1 << offset)
	}
	return mask
}

func (m *archetypeMask) addComponent(compId CompId) {
	// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
	idx := compId / 64
	offset := compId - (64 * idx)
	m[idx] |= (1 << offset)
}

func (m *archetypeMask) removeComponent(compId CompId) {
	// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
	idx := compId / 64
	offset := compId - (64 * idx)
	m[idx] &= ^(1 << offset)
}

// Performs a bitwise OR on the base mask `m` with the added mask `a`
func (m archetypeMask) bitwiseOr(a archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] | a[i]
	}
	return m
}

// Performs a bitwise AND on the base mask `m` with the added mask `a`
func (m archetypeMask) bitwiseAnd(a archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] & a[i]
	}
	return m
}

// Clears every bit in m based on the bits set in 'c'
func (m archetypeMask) bitwiseClear(c archetypeMask) archetypeMask {
	for i := range m {
		m[i] = m[i] & (^c[i])
	}
	return m
}

// m: 0x1010
// c: 0x1100
//!c: 0x0011
// f: 0x0010

// Checks to ensure archetype m contains archetype a
// Returns true if every bit in m is also set in a
// Returns false if at least one set bit in m is not set in a
func (m archetypeMask) contains(a archetypeMask) bool {
	// Logic: Bitwise AND on every segment, if the 'check' result doesn't match m[i] for that segment
	// then we know there was a bit in a[i] that was not set
	var check uint64
	for i := range m {
		check = m[i] & a[i]
		if check != m[i] {
			return false
		}
	}
	return true
}

// Checks to see if a mask m contains the supplied componentId
// Returns true if the bit location in that mask is set, else returns false
func (m archetypeMask) hasComponent(compId CompId) bool {
	// Ranges: [0, 64), [64, 128), [128, 192), [192, 256)
	idx := compId / 64
	offset := compId - (64 * idx)
	return (m[idx] & (1 << offset)) != 0
}

// Generates and returns a list of every componentId that this archetype contains
func (m archetypeMask) getComponentList() []CompId {
	ret := make([]CompId, 0)
	for compId := CompId(0); compId <= maxComponentId; compId++ {
		if m.hasComponent(compId) {
			ret = append(ret, compId)
		}
	}
	return ret
}
