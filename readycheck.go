package lifecycle

import (
	"sync"
	"sync/atomic"
	"time"
)

type ReadyCheck struct {
	componentsMutex *sync.RWMutex

	components []ComponentCheck
}

func NewReadyCheck() *ReadyCheck {
	return &ReadyCheck{
		componentsMutex: &sync.RWMutex{},
		components:      make([]ComponentCheck, 0),
	}
}

func (rdy *ReadyCheck) StartPolling() {
	for _, component := range rdy.components {
		if poll, ok := component.(*PollComponentCheck); ok {
			go poll.Start()
		}
	}
}

func (rdy *ReadyCheck) StopPolling() {
	for _, component := range rdy.components {
		if poll, ok := component.(*PollComponentCheck); ok {
			poll.Stop()
		}
	}
}

func (rdy *ReadyCheck) Ready() bool {
	rdy.componentsMutex.RLock()
	defer rdy.componentsMutex.RUnlock()

	for _, component := range rdy.components {
		isReady := component.Ready()
		if !isReady {
			return false
		}
	}

	return true
}

func (rdy *ReadyCheck) Explain() map[string]bool {
	rdy.componentsMutex.RLock()
	defer rdy.componentsMutex.RUnlock()

	explanation := make(map[string]bool, len(rdy.components))

	for _, component := range rdy.components {
		explanation[component.Name()] = component.Ready()
	}

	return explanation
}

type ComponentCheck interface {
	Name() string
	Ready() bool
}

func (rdy *ReadyCheck) RegisterPollComponent(name string, checkFn func() bool, pollDelay time.Duration) *PollComponentCheck {
	rdy.componentsMutex.Lock()
	defer rdy.componentsMutex.Unlock()

	pollComponent := &PollComponentCheck{
		name:     name,
		isReady:  &atomic.Bool{},
		isActive: &atomic.Bool{},

		checkFn:   checkFn,
		pollDelay: pollDelay,
	}

	rdy.components = append(rdy.components, pollComponent)
	return pollComponent
}

func (rdy *ReadyCheck) RegisterPushComponent(name string) *PushComponentCheck {
	rdy.componentsMutex.Lock()
	defer rdy.componentsMutex.Unlock()

	pushComponent := &PushComponentCheck{
		name:    name,
		isReady: &atomic.Bool{},
	}

	rdy.components = append(rdy.components, pushComponent)
	return pushComponent
}

func (rdy *ReadyCheck) RegisterComponent(name string, component ComponentCheck) {
	rdy.componentsMutex.Lock()
	defer rdy.componentsMutex.Unlock()

	rdy.components = append(rdy.components, component)
}
