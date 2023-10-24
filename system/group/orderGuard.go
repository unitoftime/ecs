package group

import (
	"sync"

	"github.com/unitoftime/ecs/system"
)

type orderGuard struct {
	/* Waiter for each system to wait for all the systems that must be runned before this system */
	runAfterWait map[system.SystemName]*sync.WaitGroup
	/* After system run, it releases waiter that are waiting for its completion */
	runBeforeReleases map[system.SystemName][]system.SystemName
	/* For how much systems, system must wait before start. Used in reset to reset the `runAfterWait` waiters */
	initialWaitersCount map[system.SystemName]int
}

func newOrderGuard(systems []system.System) orderGuard {
	var runBeforeReleases map[system.SystemName][]system.SystemName = map[system.SystemName][]system.SystemName{}
	var initialWaitersCount map[system.SystemName]int = map[system.SystemName]int{}

	for _, s := range systems {
		runAfter, isRunAfter := s.(system.RunAfterSystem)
		runBefore, isRunBefore := s.(system.RunBeforeSystem)

		if isRunAfter {
			for _, runAfterSystem := range runAfter.GetRunAfter() {
				beforeReleasesList, ok := runBeforeReleases[runAfterSystem]
				if !ok {
					beforeReleasesList = []system.SystemName{}
				}

				beforeReleasesList = append(beforeReleasesList, s.GetName())
				runBeforeReleases[runAfterSystem] = beforeReleasesList
			}

			initialWaitersCount[s.GetName()] = len(runAfter.GetRunAfter())
		}

		if isRunBefore {
			for _, runBeforeSystem := range runBefore.GetRunBefore() {
				beforeReleasesList, ok := runBeforeReleases[s.GetName()]
				if !ok {
					beforeReleasesList = []system.SystemName{}
				}

				beforeReleasesList = append(beforeReleasesList, runBeforeSystem)
				runBeforeReleases[s.GetName()] = beforeReleasesList

				count, ok := initialWaitersCount[runBeforeSystem]
				if !ok {
					count = 0
				}
				initialWaitersCount[runBeforeSystem] = count + 1
			}
		}
	}

	return orderGuard{
		runAfterWait:        map[system.SystemName]*sync.WaitGroup{},
		runBeforeReleases:   runBeforeReleases,
		initialWaitersCount: initialWaitersCount,
	}
}

func (g *orderGuard) Reset() {
	for systemName, initialCount := range g.initialWaitersCount {
		g.runAfterWait[systemName].Add(initialCount)
	}
}

func (g *orderGuard) Lock(s system.System) {
	if waiter, ok := g.runAfterWait[s.GetName()]; ok {
		waiter.Wait()
	}
}

func (g *orderGuard) Release(s system.System) {
	if releases, ok := g.runBeforeReleases[s.GetName()]; ok {
		for _, release := range releases {
			g.runAfterWait[release].Done()
		}
	}
}
