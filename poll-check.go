package lifecycle

import (
	"sync/atomic"
	"time"
)

type PollComponentCheck struct {
	name     string
	isReady  *atomic.Bool
	isActive *atomic.Bool

	pollDelay time.Duration
	checkFn   func() bool
}

func (component *PollComponentCheck) Name() string {
	return component.name
}

func (component *PollComponentCheck) Ready() bool {
	return component.isReady.Load()
}

func (component *PollComponentCheck) Start() {
	component.isActive.Store(true)

	for component.isActive.Load() {
		nextIsReady := component.checkFn()
		component.isReady.Store(nextIsReady)

		time.Sleep(component.pollDelay)
	}
}

func (component *PollComponentCheck) Stop() {
	component.isActive.Store(false)
}
