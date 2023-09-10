package lifecycle

import (
	"sync"
	"sync/atomic"
	"time"
)

// ReadyCheck is an utility that allows you to record the readiness status of multiple components and report them
// when necessary.
type ReadyCheck struct {
	componentsMutex *sync.RWMutex

	components []ComponentCheck
}

// NewReadyCheck creates a new instance of [ReadyCheck]
func NewReadyCheck() *ReadyCheck {
	return &ReadyCheck{
		componentsMutex: &sync.RWMutex{},
		components:      make([]ComponentCheck, 0),
	}
}

// StartPolling starts polling from poll components
func (rdy *ReadyCheck) StartPolling() {
	for _, component := range rdy.components {
		if poll, ok := component.(*PollComponentCheck); ok {
			go poll.Start()
		}
	}
}

// StopPolling stops polling from poll components
func (rdy *ReadyCheck) StopPolling() {
	for _, component := range rdy.components {
		if poll, ok := component.(*PollComponentCheck); ok {
			poll.Stop()
		}
	}
}

// Ready returns true if all components are considered ready
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

// Explain returns a map detailling which component is considered ready or not
func (rdy *ReadyCheck) Explain() map[string]bool {
	rdy.componentsMutex.RLock()
	defer rdy.componentsMutex.RUnlock()

	explanation := make(map[string]bool, len(rdy.components))

	for _, component := range rdy.components {
		explanation[component.Name()] = component.Ready()
	}

	return explanation
}

// ComponentCheck abstracts away how a check is performed and allows each component to report its readiness status
// when prompted
type ComponentCheck interface {
	Name() string
	Ready() bool
}

// RegisterPollComponent creates a new [PollComponentCheck] with the given [checkFn] and [pollDelay] and registers it
func (rdy *ReadyCheck) RegisterPollComponent(name string, checkFn func() bool, pollDelay time.Duration) *PollComponentCheck {
	pollComponent := &PollComponentCheck{
		name:     name,
		isReady:  &atomic.Bool{},
		isActive: &atomic.Bool{},

		checkFn:   checkFn,
		pollDelay: pollDelay,
	}

	rdy.RegisterComponent(name, pollComponent)

	return pollComponent
}

// RegisterPushComponent creates a new [PushComponentCheck] and registers it
func (rdy *ReadyCheck) RegisterPushComponent(name string) *PushComponentCheck {
	pushComponent := &PushComponentCheck{
		name:    name,
		isReady: &atomic.Bool{},
	}

	rdy.RegisterComponent(name, pushComponent)

	return pushComponent
}

// RegisterPulseComponent creates a new [PulseComponentCheck] and registers it
func (rdy *ReadyCheck) RegisterPulseComponent(name string, exp time.Duration) *PulseComponentCheck {
	pulseComponent := &PulseComponentCheck{
		name:       name,
		expiration: exp,
		lastPulse:  &atomic.Pointer[time.Time]{},
	}

	rdy.RegisterComponent(name, pulseComponent)
	return pulseComponent
}

// RegisterComponent registers any given [ComponentCheck] interface
func (rdy *ReadyCheck) RegisterComponent(name string, component ComponentCheck) {
	rdy.componentsMutex.Lock()
	defer rdy.componentsMutex.Unlock()

	rdy.components = append(rdy.components, component)
}
