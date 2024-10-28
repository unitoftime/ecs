package ecs

import "fmt"

// TODO: You should move to this (ie archetype graph (or bitmask?). maintain the current archetype node, then traverse to nodes (and add new ones) based on which components are added): https://ajmmertens.medium.com/building-an-ecs-2-archetypes-and-vectorization-fe21690805f9
// Dynamic component Registry
type componentRegistry struct {
	archSet  [][]archetypeId               // Contains the set of archetypeIds that have this component
	archMask map[archetypeMask]archetypeId // Contains a mapping of archetype bitmasks to archetypeIds

	revArchMask []archetypeMask // Contains the reverse mapping of archetypeIds to archetype masks. Indexed by archetypeId
}

func newComponentRegistry() *componentRegistry {
	r := &componentRegistry{
		archSet:     make([][]archetypeId, maxComponentId+1), // TODO: hardcoded to max component
		archMask:    make(map[archetypeMask]archetypeId),
		revArchMask: make([]archetypeMask, 0),
	}
	return r
}

func (r *componentRegistry) getArchetypeId(engine *archEngine, mask archetypeMask) archetypeId {
	archId, ok := r.archMask[mask]
	if !ok {
		componentIds := mask.getComponentList()
		archId = engine.newArchetypeId(mask, componentIds)
		r.archMask[mask] = archId

		if int(archId) != len(r.revArchMask) {
			panic(fmt.Sprintf("ecs: archId must increment. Expected: %d, Got: %d", len(r.revArchMask), archId))
		}
		r.revArchMask = append(r.revArchMask, mask)

		// Add this archetypeId to every component's archList
		for _, compId := range componentIds {
			r.archSet[compId] = append(r.archSet[compId], archId)
		}
	}
	return archId
}

// This is mostly for the without filter
func (r *componentRegistry) archIdOverlapsMask(archId archetypeId, compArchMask archetypeMask) bool {
	archMaskToCheck := r.revArchMask[archId]

	resultArchMask := archMaskToCheck.bitwiseAnd(compArchMask)
	if resultArchMask != blankArchMask {
		// If the resulting arch mask is nonzero, it means that both the component mask and the base mask had the same bit set, which means the arch had one of the components
		return true
	}
	return false
}
