package system

import (
	"time"

	"github.com/unitoftime/ecs"
)

type SystemName string

type System interface {
	GetName() SystemName
}

type RealtimeSystem interface {
	RunRealtime(delta time.Duration)
}

type FixedSystem interface {
	RunFixed(delta time.Duration)
}

type StepSystem interface {
	RunStep(step int32)
}

// Ordering and execution flow

/* Optional interface that can be implemented by the system to indicate that execution engine must run this system before other systems */
type RunBeforeSystem interface {
	/* Get list of systems that should be runned before this system  */
	GetRunBefore() []SystemName
}

/* Optional interface that can be implemented by the system to indicate that execution engine must run this system after other systems */
type RunAfterSystem interface {
	/* Get list of systems that should be runned after this system  */
	GetRunAfter() []SystemName
}

/*
Optional interface that can be implemented by the system to indicate that execution engine must ensure read access to the specified componets list.
During execution of this system, other systems can also read same components.

If this or `WriteComponentsSystem` interfaces not implemented for the system, execution engine will ensure write access to all the components and lock the entire world during execution of this system.
*/
type ReadComponentsSystem interface {
	/* Get list of components with read access. During execution of this system, other systems can also read same components. */
	GetReadComponents() []ecs.Component
}

/*
Optional interface that can be implemented by the system to indicate that execution engine must ensure write access to the specified componets list.
During execution of this system, no other system can read or write to specified list of components.

If this or `ReadComponentsSystem` interfaces not implemented for the system, execution engine will ensure write access to all the components and lock the entire world during execution of this system.
*/
type WriteComponentsSystem interface {
	/* Get list of components with write access. During execution of this system, no other system can read or write to specified list of components */
	GetWriteComponents() []ecs.Component
}
