package group

import (
	"fmt"
	"sort"
	"sync"

	"github.com/unitoftime/ecs"
	"github.com/unitoftime/ecs/system"
)

type componentSystemRequiredAccess struct {
	componentId ecs.ComponentId
	read        bool
}

// User by groupd to prevent race conditions for systems which operate on same compoennts
type componentsGuard struct {
	registeredSystems       map[system.SystemName]struct{}
	components              []ecs.ComponentId
	componentsLocks         map[ecs.ComponentId]*sync.RWMutex
	requiredSystemsAccesses map[system.SystemName][]componentSystemRequiredAccess
	exclusiveSystemsAccess  map[system.SystemName]struct{}
	/* Lock to use between exclusive systems. We need it because not all the components can be knows to the guard */
	exclusiveSystemsLock *sync.Mutex
}

type ComponentsGuard interface {
	InitForGroup(grp Group)
	Lock(s system.System)
	Release(s system.System)
}

func NewComponentsGuard() ComponentsGuard {
	return &componentsGuard{
		registeredSystems:       map[system.SystemName]struct{}{},
		components:              []ecs.ComponentId{},
		componentsLocks:         make(map[ecs.ComponentId]*sync.RWMutex),
		requiredSystemsAccesses: map[system.SystemName][]componentSystemRequiredAccess{},
		exclusiveSystemsAccess:  map[system.SystemName]struct{}{},
		exclusiveSystemsLock:    &sync.Mutex{},
	}
}

func (g *componentsGuard) InitForGroup(grp Group) {
	systems := grp.GetAllSystems()

	for _, s := range systems {
		if _, ok := g.registeredSystems[s.GetName()]; ok {
			panic(fmt.Sprintf("System with name %s is already registered in other gorup. System must have only one group.", s.GetName()))
		}
		g.registeredSystems[s.GetName()] = struct{}{}

		readComponents, isReadComponents := s.(system.ReadComponentsSystem)
		writeComponents, isWriteComponents := s.(system.WriteComponentsSystem)

		if isReadComponents {
			for _, component := range readComponents.GetReadComponents() {
				componentsLock, ok := g.componentsLocks[component.Id()]
				if !ok {
					componentsLock = &sync.RWMutex{}
				}
				g.componentsLocks[component.Id()] = componentsLock

				requiredAccesses, ok := g.requiredSystemsAccesses[s.GetName()]
				if !ok {
					requiredAccesses = []componentSystemRequiredAccess{}
				}

				requiredAccesses = append(requiredAccesses, componentSystemRequiredAccess{
					componentId: component.Id(),
					read:        true,
				})
				g.requiredSystemsAccesses[s.GetName()] = requiredAccesses
			}
		}

		if isWriteComponents {
			for _, component := range writeComponents.GetWriteComponents() {
				componentsLock, ok := g.componentsLocks[component.Id()]
				if !ok {
					componentsLock = &sync.RWMutex{}
				}
				g.componentsLocks[component.Id()] = componentsLock

				requiredAccesses, ok := g.requiredSystemsAccesses[s.GetName()]
				if !ok {
					requiredAccesses = []componentSystemRequiredAccess{}
				}

				requiredAccesses = append(requiredAccesses, componentSystemRequiredAccess{
					componentId: component.Id(),
					read:        false,
				})
				g.requiredSystemsAccesses[s.GetName()] = requiredAccesses
			}
		}

		if !isReadComponents && !isWriteComponents {
			g.exclusiveSystemsAccess[s.GetName()] = struct{}{}
		}
	}

	g.components = []ecs.ComponentId{}
	for component := range g.componentsLocks {
		g.components = append(g.components, component)
	}
	sort.Slice(g.components, func(i, j int) bool {
		return g.components[i] < g.components[j]
	})

	// By sorttin access lists and running locks in order we can ensure that there will be no deadlocks
	// TODO: There can be better way to do this
	for k, requiredAccesses := range g.requiredSystemsAccesses {
		sort.Slice(requiredAccesses, func(i, j int) bool {
			return requiredAccesses[i].componentId < requiredAccesses[j].componentId
		})
		g.requiredSystemsAccesses[k] = requiredAccesses
	}
}

func (g *componentsGuard) Lock(s system.System) {
	if _, ok := g.exclusiveSystemsAccess[s.GetName()]; ok {
		// Exclusive system. Get all the write locks to all the components
		for _, component := range g.components {
			g.componentsLocks[component].Lock()
		}

		g.exclusiveSystemsLock.Lock()
	} else {
		// Just lock for the components that are required by this system
		for _, requiredAccess := range g.requiredSystemsAccesses[s.GetName()] {
			if requiredAccess.read {
				g.componentsLocks[requiredAccess.componentId].RLock()
			} else {
				g.componentsLocks[requiredAccess.componentId].Lock()
			}
		}
	}
}

func (g *componentsGuard) Release(s system.System) {
	if _, ok := g.exclusiveSystemsAccess[s.GetName()]; ok {
		g.exclusiveSystemsLock.Unlock()

		// Exclusive system. Release all the write locks to all the components
		for _, component := range g.components {
			g.componentsLocks[component].Unlock()
		}
	} else {
		// Just release locks for the components that are required by this system
		for _, requiredAccess := range g.requiredSystemsAccesses[s.GetName()] {
			if requiredAccess.read {
				g.componentsLocks[requiredAccess.componentId].RUnlock()
			} else {
				g.componentsLocks[requiredAccess.componentId].Unlock()
			}
		}
	}
}
