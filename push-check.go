package lifecycle

type PushComponentCheck struct {
	name    string
	isReady bool
}

func (component *PushComponentCheck) Name() string {
	return component.name
}

func (component *PushComponentCheck) Ready() bool {
	return component.isReady
}

func (component *PushComponentCheck) SetReady(isReady bool) {
	component.isReady = isReady
}
