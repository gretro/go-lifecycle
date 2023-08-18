package lifecycle

import "time"

type PollComponentCheck struct {
	name     string
	isReady  bool
	isActive bool

	pollDelay time.Duration
	checkFn   func() bool
}

func (component *PollComponentCheck) Name() string {
	return component.name
}

func (component *PollComponentCheck) Ready() bool {
	return component.isReady
}

func (component *PollComponentCheck) Start() {
	component.isActive = true

	for component.isActive {
		component.isReady = component.checkFn()

		time.Sleep(component.pollDelay)
	}
}

func (component *PollComponentCheck) Stop() {
	component.isActive = false
}
