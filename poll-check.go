package lifecycle

import (
	"sync/atomic"
	"time"
)

// PollComponentCheck is a component check where the reporting mechanism will be polled
// every X amount of time.
type PollComponentCheck struct {
	name     string
	isReady  *atomic.Bool
	isActive *atomic.Bool

	pollDelay time.Duration
	checkFn   func() bool
}

// Name is the name of the component being checked for
func (component *PollComponentCheck) Name() string {
	return component.name
}

// Ready returns true if the component was ready the last time it was polled
func (component *PollComponentCheck) Ready() bool {
	return component.isReady.Load()
}

// Start will poll the component every X amount of time. This is a blocking method.
func (component *PollComponentCheck) Start() {
	component.isActive.Store(true)

	for component.isActive.Load() {
		nextIsReady := component.checkFn()
		component.isReady.Store(nextIsReady)

		time.Sleep(component.pollDelay)
	}
}

// Stop will break the polling if it was previously started.
func (component *PollComponentCheck) Stop() {
	component.isActive.Store(false)
}
