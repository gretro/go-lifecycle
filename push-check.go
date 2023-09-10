package lifecycle

import "sync/atomic"

// PushComponentCheck performs a readiness check based on a manual input.
type PushComponentCheck struct {
	name    string
	isReady *atomic.Bool
}

// Name is the name of the component being checked for
func (component *PushComponentCheck) Name() string {
	return component.name
}

// Ready returns true if the last readiness check set was true
func (component *PushComponentCheck) Ready() bool {
	return component.isReady.Load()
}

// SetReady records the readiness check to be persisted
func (component *PushComponentCheck) SetReady(isReady bool) {
	component.isReady.Store(isReady)
}
