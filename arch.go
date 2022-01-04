package ecs

import (
	"fmt"
	"sort"
)

type LookupList struct {
	Lookup map[Id]int
	Ids []Id
}

type ArchEngine struct {
	archCounter ArchId
	reg map[string]Storage
	lookup map[ArchId]*LookupList
	dcr *DCR
}

type Storage interface {
	New(ArchId)
	Get(ArchId, any) bool // Gets Component Storage []Component
	Set(ArchId, any)      // Sets Component Storage []Component
	InternalGet(ArchId, int) (any, bool)  // Gets a component from component storage at an index
	InternalSet(ArchId, int, any) // Sets a component in component storage at an index
	InternalAppend(ArchId, any) // Appends a component to componenet Storage
	InternalDelete(ArchId, int) // Appends a component to componenet Storage
	// Len() int // Gets the current number of components in storage
	Has(ArchId) bool // Checks if the Storage has the ArchId in it
}

type Iterator[T any] interface {
	Get (idx int) T
}

type ValueIterator[T any] struct {
	ptr bool
	list []T
}
func (i ValueIterator[T]) Get(idx int) T {
	return i.list[idx]
}

// type PointerIterator[T] struct {
// 	list []&T
// }
// func (i *PointerIterator[T]) Get(idx int) T {
// 	return &i.list[idx]
// }

func GetStorageList[T any](s Storage, archId ArchId) []T {
	return s.(*ArchStorage[[]T, T]).list[archId]
}

func NewArchEngine() *ArchEngine {
	return &ArchEngine{
		// archCounter: 0,
		reg: make(map[string]Storage),
		lookup: make(map[ArchId]*LookupList),
		dcr: NewDCR(),
	}
}

// func (e *ArchEngine) NewArchId() ArchId {
// 	archId := e.archCounter
// 	e.archCounter++
// 	return archId
// }

func (e *ArchEngine) Print() {
	fmt.Println(e)
}

func (e ArchEngine) GetStorage(t any) Storage {
	n := name(t)
	iStorage, ok := e.reg[n]
	if !ok {
		panic("Attempting to read unregistered component")
	}
	return iStorage
}

func (e *ArchEngine) Read(archId ArchId, t any, val any) bool {
	n := name(t)
	iStorage, ok := e.reg[n]
	if !ok {
		panic("Attempting to read unregistered component")
		return false
	}

	iStorage.Get(archId, val)
	return true
}

func (e *ArchEngine) Read2(archId ArchId, t any) (Storage, bool) {
	n := name(t)
	iStorage, ok := e.reg[n]
	if !ok {
		panic("Attempting to read unregistered component")
		return nil, false
	}

	if iStorage.Has(archId) {
		return iStorage, true
	}
		return nil, false
}

func (e *ArchEngine) Write(archId ArchId, t any, val any) {
	n := name(t)
	iStorage, ok := e.reg[n]
	if !ok {
		panic("Attempting to read unregistered component")
	}

	iStorage.Set(archId, val)
}

// TODO - should this read all archetype's storage arrays?
func (e *ArchEngine) ReadAll(archId ArchId, id Id) Entity {
	// Lookup the id and it's index
	lookup, ok := e.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.Lookup[id]
	if !ok { panic("Entity ID doesn't exist in archetype ID!") }

	ent := NewEntity()
	for _, storage := range e.reg {
		comp, ok := storage.InternalGet(archId, index)
		if ok {
			ent.Add(comp)
		}
	}

	return ent
}

func (e *ArchEngine) DeleteAll(archId ArchId, id Id) {
	// Lookup the id and it's index
	lookup, ok := e.lookup[archId]
	if !ok { panic("LookupList is missing!") }
	index, ok := lookup.Lookup[id]
	if !ok { panic("Entity ID doesn't exist in archetype ID!") }

	for _, storage := range e.reg {
		storage.InternalDelete(archId, index)
	}

	// TODO - put this in its own function?
	// Delete id from LookupList
	// fmt.Println(len(lookup.Ids), len(lookup.Lookup), index, id)
	delete(lookup.Lookup, id)
	lastVal := lookup.Ids[len(lookup.Ids)-1]
	lookup.Ids[index] = lastVal
	lookup.Ids = lookup.Ids[:len(lookup.Ids)-1]

	// Reassign the lastVal (ID) to the index we moved it to
	lookup.Lookup[lastVal] = index
}

// func ArchWrite[T any](engine *ArchEngine, archId ArchId, val []T) {
// 	var t T
// 	n := name(t)
// 	iStorage, ok := engine.reg[n]
// 	if !ok {
// 		engine.reg[n] = NewArchStorage[[]T, T]()
// 		iStorage = engine.reg[n]
// 	}

// 	storage := iStorage.(*ArchStorage[[]T, T])

// 	storage.list[archId] = val
// }

// func ArchRead[T any](engine *ArchEngine, archId ArchId) ([]T, bool) {
// 	var t T
// 	n := name(t)
// 	iStorage, ok := engine.reg[n]
// 	if !ok {
// 		var ret []T
// 		return ret, false
// 	}

// 	storage := iStorage.(*ArchStorage[[]T, T])

// 	ret, ok := storage.list[archId]
// 	return ret, ok
// }

// func ArchReadAll(engine *ArchEngine, archId ArchId) []any {
// 	ret := make([]ArchComponent, 0)
// 	for _, storage := range e.reg {
// 		comp, ok := storage.list[archId]
// 		if !ok { continue }

// 		ret = append(ret, comp.(ArchComponent))
// 	}
// 	return ret
// }

func (e *ArchEngine) GetArchId(comp ...any) ArchId {
	archId := e.dcr.GetArchId(comp...)

	_, ok := e.lookup[archId]
	if !ok {
		// Need to build this archetype because it doesn't exist

		// All archetypes get the LookupList component
		lookup := &LookupList{
			Lookup: make(map[Id]int),
			Ids: make([]Id, 0),
		}
		// ArchWrite(e, archId, lookup)
		e.lookup[archId] = lookup


		// Add all component Lists to this archetype
		for i := range comp {
			n := name(comp[i])
			storage, ok := e.reg[n]
			if !ok { panic("Bug: Archetype doesn't have its component storage setup!") }
			storage.New(archId)
		}
	}

	return archId
}

// Returns the list of ArchIds that contain all components
// TODO - this can be optimized
func (e *ArchEngine) Filter(comp ...any) []ArchId {
	lists := make([]map[ArchId]bool, 0)
	for i := range comp {
		n := name(comp[i])
		lists = append(lists, e.dcr.archSet[n])
	}

	archIds := make([]ArchId, 0)
	for archId := range lists[0] {
		missing := false
		for i := range lists {
			_, exists := lists[i][archId]
			if !exists {
				missing = true
				break // at least one set was missing
			}
		}
		if !missing {
			archIds = append(archIds, archId)
		}
	}

	return archIds

	// archIds := make([]ArchId, len(lists[0]))
	// copy(archIds, lists[0]) // There's always at least 1 in the variadic
	// for i := 1; i < len(lists); i++ {
	// 	archIds = Intersect(archIds, lists[i])
	// }
	// return archIds
}

// func Intersect(a, b []ArchId) []ArchId {
// 	set := make(map[ArchId]bool)
// 	for _, id := range a {
// 		set[id] = true
// 	}

// 	ret := make([]ArchId, 0)
// 	for _, id := range b {
// 		_, ok := set[id]
// 		if ok {
// 			ret = append(ret, id)
// 		}
// 	}
// 	return ret
// }

// Uses component mask to generate an archetype ID or creates that archetype
// func (e *ArchEngine) GetArchId(comp ...interface{}) ArchId {
// 	// mask := e.GetArchMask(comp...)
// 	mask := ArchMask(0)
// 	for i := range comp {
// 		mask += e.componentRegistry[name(comp[i])]
// 	}

// 	archId, ok := e.archLookup[mask]
// 	if !ok {
// 		// Need to build this archetype because it doesn't exist
// 		archId = e.NewArchId()
// 		e.archLookup[mask] = archId

// 		// All archetypes get the LookupList component
// 		lookup := &LookupList{
// 			Lookup: make(map[Id]int),
// 			Ids: make([]Id, 0),
// 		}
// 		ArchWrite(e, archId, lookup)

// 		// Add all component Lists to this archetype
// 		for i := range comp {
// 			list := e.componentRegistry[name(comp[i])]
// 			ArchWrite(e, archId, list)
// 		}
// 	}

// 	return archId
// }

// Storage for ArchId's, Typically T will be a list (indexed by Id) of some component type
type ArchStorage[S Slice[T], T any] struct {
	list map[ArchId]S
}
func NewArchStorage[S Slice[T], T any]() *ArchStorage[S, T] {
	return &ArchStorage[S, T]{
		list: make(map[ArchId]S),
	}
}

func (s *ArchStorage[S, T]) New(archId ArchId) {
	_, ok := s.list[archId]
	if !ok {
		s.list[archId] = make(S, 0)
	}
}

// TODO - use this to replace new?
func (s *ArchStorage[S, T]) Get(archId ArchId, val any) bool {
	_, ok := s.list[archId]
	if !ok {
		return false
		// s.list[archId] = make(S, 0)
	}
	*(val.(*[]T)) = s.list[archId]
	return true
}

func (s *ArchStorage[S, T]) Set(archId ArchId, val any) {
	s.list[archId] = val.([]T)
	// *(val.(*[]T)) = s.list[archId]
}

// TODO - rename
// TODO - this is operating on the archetype's component slice. Should I just make a new interface for that to implement and put that there?
// TODO - handle index oob
func (s *ArchStorage[S, T]) InternalGet(archId ArchId, index int) (any, bool) {
	var ret T
	sl, ok := s.list[archId]
	if !ok {
		return nil, false
	}

	ret = sl[index]
	return ret, true
}

// TODO - rename
func (s *ArchStorage[S, T]) InternalSet(archId ArchId, index int, val any) {
	sl, ok := s.list[archId]
	if !ok { panic("Archetype doesn't have this component!") }

	sl[index] = val.(T)
}

func (s *ArchStorage[S, T]) InternalAppend(archId ArchId, comp any) {
	sl, ok := s.list[archId]
	if !ok { panic("Archetype doesn't have this component!") }

	s.list[archId] = append(sl, comp.(T))
}

func (s *ArchStorage[S, T]) InternalDelete(archId ArchId, index int) {
	sl, ok := s.list[archId]
	if !ok { return }

	// Move val at last position into hole
	lastVal := sl[len(sl)-1]
	sl[index] = lastVal
	sl = sl[:len(sl)-1]
}

func (s *ArchStorage[S, T]) Has(archId ArchId) bool {
	_, ok := s.list[archId]
	return ok
}

type CompId uint16

// Dynamic Component Registry
type DCR struct {
	archCounter ArchId
	compCounter CompId
	mapping map[string]CompId // Contains the CompId for the component name
	archSet map[string]map[ArchId]bool // Contains the set of ArchIds that have this component
	// componentStorageType map[string]any
	trie *node
}

func NewDCR() *DCR {
	r := &DCR{
		archCounter: 0,
		compCounter: 0,
		mapping: make(map[string]CompId),
		archSet: make(map[string]map[ArchId]bool),
	}
	r.trie = NewNode(r)
	return r
}

func (r *DCR) NewArchId() ArchId {
	archId := r.archCounter
	r.archCounter++
	return archId
}

// 1. Map all components to their component Id
// 2. Sort all component ids so that we can index the prefix tree
// 3. Walk the prefix tree to find the ArchId
func (r *DCR) GetArchId(comp ...any) ArchId {
	list := make([]CompId, len(comp))
	for i := range comp {
		list[i] = r.Register(comp[i])
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})

	cur := r.trie
	for _, idx := range list {
		cur = cur.Get(r, idx)
	}

	// Add this ArchId to every component's archList
	for _, c := range comp {
		n := name(c)
		r.archSet[n][cur.archId] = true

		// r.archList[n] = append(r.archList[n], cur.archId)
		// TODO - sort these to improve my filter speed?
		// sort.Slice(r.archList[n], func(i, j int) bool {
		// 	return r.archList[n][i] < r.archList[n][j]
		// })
	}
	return cur.archId
}

// Registers a component to a component Id and returns the Id
// If already registered, just return the Id and don't make a new one
func (r *DCR) Register(comp any) CompId {
	n := name(comp)

	id, ok := r.mapping[n]
	if !ok {
		// add empty ArchList
		r.archSet[n] = make(map[ArchId]bool)

		r.compCounter++
		r.mapping[n] = r.compCounter
		return r.compCounter
	}

	return id
}

type node struct {
	archId ArchId
//	parent *node
	child []*node
}

func NewNode(r *DCR) *node {
	return &node{
		archId: r.NewArchId(),
		child: make([]*node, 0),
	}
}

func (n *node) Get(r *DCR, id CompId) *node {
	if id < CompId(len(n.child)) {
		if n.child[id] == nil {
			n.child[id] = NewNode(r)
		}
		return n.child[id]
	}

	// Expand the slice to hold all required children
	n.child = append(n.child, make([]*node, 1 + int(id) - len(n.child))...)
	if n.child[id] == nil {
		n.child[id] = NewNode(r)
	}
	return n.child[id]
}
