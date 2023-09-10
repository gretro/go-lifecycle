package lifecycle

import "sync/atomic"

type PushComponentCheck struct {
	name    string
	isReady *atomic.Bool
}

func (component *PushComponentCheck) Name() string {
	return component.name
}

func (component *PushComponentCheck) Ready() bool {
	return component.isReady.Load()
}

func (component *PushComponentCheck) SetReady(isReady bool) {
	component.isReady.Store(isReady)
}
