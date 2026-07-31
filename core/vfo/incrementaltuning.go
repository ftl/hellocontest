package vfo

import (
	"github.com/ftl/hellocontest/core"
)

type incrementalTuningState struct {
	intentActive bool
	actualActive bool
	offset       core.Frequency
}

type IncrementalTuningControl struct {
	state    [2]incrementalTuningState
	workmode core.Workmode

	vfo *VFO
}

func (x *IncrementalTuningControl) availableOnVFO() bool {
	if x.vfo.id == core.VFO1 {
		return true
	}
	if x.vfo.online() {
		return x.vfo.client.IncrementalTuningPerVFO()
	} else {
		return x.vfo.offlineClient.IncrementalTuningPerVFO()
	}
}

func (x *IncrementalTuningControl) setIncrementalTuning(kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	if x.vfo.online() {
		x.vfo.client.SetIncrementalTuning(x.vfo.id, kind, active, offset)
	} else {
		x.vfo.offlineClient.SetIncrementalTuning(kind, active, offset)
	}
}

func (x *IncrementalTuningControl) IncrementalTuningActive(kind core.IncrementalTuningKind) bool {
	if !x.availableOnVFO() {
		return false
	}
	return x.state[kind].intentActive
}

func (x *IncrementalTuningControl) SetIncrementalTuningActive(kind core.IncrementalTuningKind, active bool) {
	if !x.availableOnVFO() {
		return
	}
	x.state[kind].intentActive = active
	x.activate()
	x.emitIncrementalTuningActiveChanged(kind, active)
}

func (x *IncrementalTuningControl) ShiftIncrementalTuning(kind core.IncrementalTuningKind, delta core.Frequency) {
	if !x.availableOnVFO() {
		return
	}
	x.state[kind].offset += delta
	x.vfo.setIncrementalTuning(kind, x.state[kind].actualActive, x.state[kind].offset)
}

func (x *IncrementalTuningControl) ToggleAvailableIncrementalTuning() {
	kind, ok := x.availableIncrementalTuningKind()
	if !ok {
		return
	}
	x.SetIncrementalTuningActive(kind, !x.state[kind].intentActive)
}

func (x *IncrementalTuningControl) ShiftAvailableIncrementalTuning(sign core.Frequency) {
	kind, ok := x.availableIncrementalTuningKind()
	if !ok {
		return
	}
	step := core.DefaultXITShift
	if kind == core.RIT {
		step = core.DefaultRITShift
	}
	x.ShiftIncrementalTuning(kind, sign*step)
}

func (x *IncrementalTuningControl) availableIncrementalTuningKind() (core.IncrementalTuningKind, bool) {
	if !x.availableOnVFO() {
		return -1, false
	}
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		if kind.Workmode() == x.workmode {
			return kind, true
		}
	}
	return -1, false
}

func (x *IncrementalTuningControl) visible(kind core.IncrementalTuningKind) bool {
	if !x.availableOnVFO() {
		return false
	}
	return kind.Workmode() == x.workmode
}

func (x *IncrementalTuningControl) WorkmodeChanged(vfo core.VFOID, workmode core.Workmode) {
	if vfo != x.vfo.id {
		return
	}
	x.workmode = workmode
	x.activate()
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		x.emitIncrementalTuningVisibilityChanged(kind)
	}
}

func (x *IncrementalTuningControl) activate() {
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		x.state[kind].actualActive = x.state[kind].intentActive && kind.Workmode() == x.workmode
		x.setIncrementalTuning(kind, x.state[kind].actualActive, x.state[kind].offset)
	}
}

func (x *IncrementalTuningControl) VFOIncrementalTuningChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	if vfo != x.vfo.id {
		return
	}
	x.state[kind].actualActive = active
	x.state[kind].offset = offset
}

func (x *IncrementalTuningControl) IncrementalTuningAvailabilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, available bool) {
	if vfo != x.vfo.id {
		return
	}
	x.emitIncrementalTuningVisibilityChanged(kind)
}

func (x *IncrementalTuningControl) emitIncrementalTuningChanged(kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	core.Emit(x.vfo.listeners, func(listener core.VFOIncrementalTuningListener) {
		x.vfo.asyncRunner(func() {
			listener.VFOIncrementalTuningChanged(x.vfo.id, kind, active, offset)
		})
	})
}

func (x *IncrementalTuningControl) emitIncrementalTuningActiveChanged(kind core.IncrementalTuningKind, active bool) {
	core.Emit(x.vfo.listeners, func(l core.IncrementalTuningActiveListener) {
		x.vfo.asyncRunner(func() {
			l.IncrementalTuningActiveChanged(x.vfo.id, kind, active)
		})
	})
}

func (x *IncrementalTuningControl) emitIncrementalTuningVisibilityChanged(kind core.IncrementalTuningKind) {
	visible := x.visible(kind)
	core.Emit(x.vfo.listeners, func(l core.IncrementalTuningVisibilityListener) {
		x.vfo.asyncRunner(func() {
			l.IncrementalTuningVisibilityChanged(x.vfo.id, kind, visible)
		})
	})
}
