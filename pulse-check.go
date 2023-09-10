package lifecycle

import (
	"sync/atomic"
	"time"
)

// PulseComponentCheck performs readiness check based on a timeout. Each pulse marks
// the component as being ready, until a given duration.
type PulseComponentCheck struct {
	name       string
	expiration time.Duration

	lastPulse *atomic.Pointer[time.Time]
}

// Name is the name of the component being checked for
func (component *PulseComponentCheck) Name() string {
	return component.name
}

// Ready returns true if the last pulse recorded was before the expiration
func (component *PulseComponentCheck) Ready() bool {
	lastPulse := component.lastPulse.Load()

	if lastPulse == nil {
		return false
	}

	return time.Since(*lastPulse) <= component.expiration
}

// RecordPulse records a pulse from the component and marks the component as being
// alive until the state expires
func (component *PulseComponentCheck) RecordPulse() {
	now := time.Now()
	component.lastPulse.Store(&now)
}
