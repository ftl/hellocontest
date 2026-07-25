package vfo

import (
	"github.com/ftl/hellocontest/core"
)

type IncrementalTuningActiveListener interface {
	IncrementalTuningActiveChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool)
}

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

func kindForWorkmode(kind core.IncrementalTuningKind) core.Workmode {
	if kind == core.RIT {
		return core.Run
	}
	return core.SearchPounce
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

func (x *IncrementalTuningControl) emitIncrementalTuningActiveChanged(kind core.IncrementalTuningKind, active bool) {
	for _, listener := range x.vfo.listeners {
		if l, ok := listener.(IncrementalTuningActiveListener); ok {
			x.vfo.asyncRunner(func() {
				l.IncrementalTuningActiveChanged(x.vfo.id, kind, active)
			})
		}
	}
}

func (x *IncrementalTuningControl) WorkmodeChanged(vfo core.VFOID, workmode core.Workmode) {
	if vfo != x.vfo.id {
		return
	}
	x.workmode = workmode
	x.activate()
}

func (x *IncrementalTuningControl) activate() {
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		x.state[kind].actualActive = x.state[kind].intentActive && kindForWorkmode(kind) == x.workmode
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
