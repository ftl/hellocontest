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
	state     [2]incrementalTuningState
	available [2]bool
	workmode  core.Workmode

	vfo *VFO
}

func (x *IncrementalTuningControl) IncrementalTuningActive(kind core.IncrementalTuningKind) bool {
	return x.state[kind].intentActive
}

func (x *IncrementalTuningControl) SetIncrementalTuningActive(kind core.IncrementalTuningKind, active bool) {
	x.state[kind].intentActive = active
	x.activate()
	x.emitIncrementalTuningActiveChanged(kind, active)
}

func (x *IncrementalTuningControl) ShiftOffset(kind core.IncrementalTuningKind, delta core.Frequency) {
	x.state[kind].offset += delta
	x.vfo.SetIncrementalTuning(kind, x.state[kind].actualActive, x.state[kind].offset)
}

func (x *IncrementalTuningControl) ToggleIncrementalTuning() {
	kind, ok := x.AvailableIncrementalTuningKind()
	if !ok {
		return
	}
	x.SetIncrementalTuningActive(kind, !x.state[kind].intentActive)
}

func (x *IncrementalTuningControl) ShiftAvailableIncrementalTuning(sign core.Frequency) {
	kind, ok := x.AvailableIncrementalTuningKind()
	if !ok {
		return
	}
	step := core.DefaultXITShift
	if kind == core.RIT {
		step = core.DefaultRITShift
	}
	x.ShiftOffset(kind, sign*step)
}

func (x *IncrementalTuningControl) AvailableIncrementalTuningKind() (core.IncrementalTuningKind, bool) {
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		if x.available[kind] && kind.Workmode() == x.workmode {
			return kind, true
		}
	}
	return 0, false
}

func (x *IncrementalTuningControl) visible(kind core.IncrementalTuningKind) bool {
	return x.available[kind] && kind.Workmode() == x.workmode
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
		x.vfo.SetIncrementalTuning(kind, x.state[kind].actualActive, x.state[kind].offset)
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
	x.available[kind] = available
	x.emitIncrementalTuningVisibilityChanged(kind)
}

func (x *IncrementalTuningControl) emitIncrementalTuningActiveChanged(kind core.IncrementalTuningKind, active bool) {
	for _, listener := range x.vfo.listeners {
		if l, ok := listener.(core.IncrementalTuningActiveListener); ok {
			x.vfo.asyncRunner(func() {
				l.IncrementalTuningActiveChanged(x.vfo.id, kind, active)
			})
		}
	}
}

func (x *IncrementalTuningControl) emitIncrementalTuningVisibilityChanged(kind core.IncrementalTuningKind) {
	visible := x.visible(kind)
	for _, listener := range x.vfo.listeners {
		if l, ok := listener.(core.IncrementalTuningVisibilityListener); ok {
			x.vfo.asyncRunner(func() {
				l.IncrementalTuningVisibilityChanged(x.vfo.id, kind, visible)
			})
		}
	}
}
