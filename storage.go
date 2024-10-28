package ecs

type storage interface {
	ReadToEntity(*Entity, archetypeId, int) bool
	ReadToRawEntity(*RawEntity, archetypeId, int) bool
	Allocate(archetypeId, int) // Allocates the index, setting the data there to the zero value
	Delete(archetypeId, int)
	moveArchetype(archetypeId, int, archetypeId, int)
}

// --------------------------------------------------------------------------------
// - Lookup List
// --------------------------------------------------------------------------------
// TODO: Rename, this is kind of like an archetype header
type lookupList struct {
	index      *internalMap[Id, int] // A mapping from entity ids to array indices
	id         []Id                  // An array of every id in the arch list (essentially a reverse mapping from index to Id)
	holes      []int                 // List of indexes that have ben deleted
	mask       archetypeMask
	components []componentId // This is a list of all components that this archetype contains
}

func (l *lookupList) Len() int {
	return l.index.Len()
}

// Adds ourselves to the last available hole, else appends
// Returns the index
func (l *lookupList) addToEasiestHole(id Id) int {
	if len(l.holes) > 0 {
		lastHoleIndex := len(l.holes) - 1
		index := l.holes[lastHoleIndex]
		l.id[index] = id
		l.index.Put(id, index)

		l.holes = l.holes[:lastHoleIndex]
		return index
	} else {
		// Because the Id hasn't been added to this arch, we need to append it to the end
		l.id = append(l.id, id)
		index := len(l.id) - 1
		l.index.Put(id, index)
		return index
	}
}

// --------------------------------------------------------------------------------
// - ComponentSlice
// --------------------------------------------------------------------------------
type componentSlice[T any] struct {
	comp []T
}

// Note: This will panic if you write past the buffer by more than 1
func (s *componentSlice[T]) Write(index int, val T) {
	if index == len(s.comp) {
		// Case: index causes a single append (new element added)
		s.comp = append(s.comp, val)
	} else {
		// Case: index is inside the length
		// Edge: (Causes Panic): Index is greater than 1 plus length
		s.comp[index] = val
	}
}

// --------------------------------------------------------------------------------
// - ComponentSliceStorage
// --------------------------------------------------------------------------------
type componentSliceStorage[T any] struct {
	// TODO: Could these just increment rather than be a map lookup? I guess not every component type would have a storage slice for every archetype so we'd waste some memory. I guess at the very least we could use the faster lookup map
	slice map[archetypeId]*componentSlice[T]
}

func (ss *componentSliceStorage[T]) ReadToEntity(entity *Entity, archId archetypeId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return false
	}
	entity.Add(C(cSlice.comp[index]))
	return true
}

func (ss *componentSliceStorage[T]) ReadToRawEntity(entity *RawEntity, archId archetypeId, index int) bool {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return false
	}
	entity.Add(&cSlice.comp[index])
	return true
}

func (ss *componentSliceStorage[T]) Allocate(archId archetypeId, index int) {
	cSlice, ok := ss.slice[archId]
	if !ok {
		cSlice = &componentSlice[T]{
			comp: make([]T, 0, DefaultAllocation),
		}
		ss.slice[archId] = cSlice
	}

	var val T
	cSlice.Write(index, val)
}

func (ss *componentSliceStorage[T]) moveArchetype(oldArchId archetypeId, oldIndex int, newArchId archetypeId, newIndex int) {
	oldSlice := ss.slice[oldArchId]
	newSlice := ss.slice[newArchId]

	val := oldSlice.comp[oldIndex]
	newSlice.Write(newIndex, val)
}

// Delete is somewhat special because it deletes the index of the archId for the componentSlice
// but then plugs the hole by pushing the last element of the componentSlice into index
func (ss *componentSliceStorage[T]) Delete(archId archetypeId, index int) {
	cSlice, ok := ss.slice[archId]
	if !ok {
		return
	}

	lastVal := cSlice.comp[len(cSlice.comp)-1]
	cSlice.comp[index] = lastVal
	cSlice.comp = cSlice.comp[:len(cSlice.comp)-1]
}
