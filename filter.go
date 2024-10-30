package ecs

import (
	"slices"
)

// Optional - Lets you view even if component is missing (func will return nil)
// With - Lets you add additional components that must be present
// Without - Lets you add additional components that must not be present
type Filter interface {
	Filter([]CompId) []CompId
}

type without struct {
	mask archetypeMask
}

// Creates a filter to ensure that entities will not have the specified components
func Without(comps ...any) without {
	return without{
		mask: buildArchMaskFromAny(comps...),
	}
}
func (w without) Filter(list []CompId) []CompId {
	return list // Dont filter anything. We need to exclude later on
	// return append(list, w.comps...)
}

type with struct {
	comps []CompId
}

// Creates a filter to ensure that entities have the specified components
func With(comps ...any) with {
	ids := make([]CompId, len(comps))
	for i := range comps {
		ids[i] = name(comps[i])
	}
	return with{
		comps: ids,
	}
}

func (w with) Filter(list []CompId) []CompId {
	return append(list, w.comps...)
}

type optional struct {
	comps []CompId
}

// Creates a filter to make the query still iterate even if a specific component is missing, in which case you'll get nil if the component isn't there when accessed
func Optional(comps ...any) optional {
	ids := make([]CompId, len(comps))
	for i := range comps {
		ids[i] = name(comps[i])
	}

	return optional{
		comps: ids,
	}
}

func (f optional) Filter(list []CompId) []CompId {
	for i := 0; i < len(list); i++ {
		for j := range f.comps {
			if list[i] == f.comps[j] {
				// If we have a match, we want to remove it from the list.
				list[i] = list[len(list)-1]
				list = list[:len(list)-1]

				// Because we just moved the last element to index i, we need to go back to process that element
				i--
				break
			}
		}
	}
	return list
}

type filterList struct {
	comps                     []CompId
	withoutArchMask           archetypeMask
	cachedArchetypeGeneration int // Denotes the world's archetype generation that was used to create the list of archIds. If the world has a new generation, we should probably regenerate
	archIds                   []archetypeId
}

func newFilterList(comps []CompId, filters ...Filter) filterList {
	var withoutArchMask archetypeMask
	for _, f := range filters {
		withoutFilter, isWithout := f.(without)
		if isWithout {
			withoutArchMask = withoutFilter.mask
		} else {
			comps = f.Filter(comps)
		}
	}

	return filterList{
		comps:           comps,
		withoutArchMask: withoutArchMask,
		archIds:         make([]archetypeId, 0),
	}
}
func (f *filterList) regenerate(world *World) {
	if world.engine.getGeneration() != f.cachedArchetypeGeneration {
		f.archIds = world.engine.FilterList(f.archIds, f.comps)

		if f.withoutArchMask != blankArchMask {
			f.archIds = slices.DeleteFunc(f.archIds, func(archId archetypeId) bool {
				return world.engine.dcr.archIdOverlapsMask(archId, f.withoutArchMask)
			})
		}

		f.cachedArchetypeGeneration = world.engine.getGeneration()
	}
}
