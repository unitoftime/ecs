package ecs

type OnAdd struct {
	compId CompId
}

var _onAddId = NewEvent[OnAdd]()

func (p OnAdd) EventId() EventId {
	return _onAddId
}

//--------------------------------------------------------------------------------

func (e *archEngine) runFinalizedHooks(id Id) {
	// Run, then clear add hooks
	for i := range e.finalizeOnAdd {
		e.runAddHook(id, e.finalizeOnAdd[i])
	}
	e.finalizeOnAdd = e.finalizeOnAdd[:0]

	// TODO: Run other hooks?
}

func (e *archEngine) runAddHook(id Id, compId CompId) {
	current := e.onAddHooks[compId]
	if current == nil {
		return
	}

	current.Run(id, OnAdd{compId})
}

// Marks all provided components
func markComponents(slice []CompId, comp ...Component) []CompId {
	for i := range comp {
		slice = append(slice, comp[i].CompId())
	}
	return slice
}

// Marks the provided components, excluding ones that are already set by the old mask
func markNewComponents(slice []CompId, oldMask archetypeMask, comp ...Component) []CompId {
	for i := range comp {
		compId := comp[i].CompId()
		if oldMask.hasComponent(compId) {
			continue // Skip: Component already set in oldMask
		}

		slice = append(slice, compId)
	}
	return slice
}

func markComponentMask(slice []CompId, mask archetypeMask) []CompId {
	// TODO: Optimization: Technically this only has to loop to the max registered compId, not the max possible. Also see optimization note in archEngine
	for compId := CompId(0); compId <= maxComponentId; compId++ {
		if mask.hasComponent(compId) {
			slice = append(slice, compId)
		}
	}

	return slice
}

func markComponentDiff(slice []CompId, newMask, oldMask archetypeMask) []CompId {
	mask := newMask.bitwiseClear(oldMask)
	return markComponentMask(slice, mask)
}
