package vfo

import (
	"github.com/ftl/hellocontest/core"
)

type incrementalTuningState struct {
	intentActive bool
	actualActive bool
	offset       core.Frequency
}

func (v *VFO) availableOnVFO() bool {
	if v.id == core.VFO1 {
		return true
	}
	if v.online() {
		return v.client.IncrementalTuningPerVFO()
	} else {
		return v.offlineClient.IncrementalTuningPerVFO()
	}
}

func (v *VFO) setIncrementalTuning(kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	if v.online() {
		v.client.SetIncrementalTuning(v.id, kind, active, offset)
	} else {
		v.offlineClient.SetIncrementalTuning(kind, active, offset)
	}
}

func (v *VFO) IncrementalTuningActive(kind core.IncrementalTuningKind) bool {
	if !v.availableOnVFO() {
		return false
	}
	return v.state[kind].intentActive
}

func (v *VFO) SetIncrementalTuningActive(kind core.IncrementalTuningKind, active bool) {
	if !v.availableOnVFO() {
		return
	}
	v.state[kind].intentActive = active
	v.activate()
	v.emitIncrementalTuningActiveChanged(kind, active)
}

func (v *VFO) ShiftIncrementalTuning(kind core.IncrementalTuningKind, delta core.Frequency) {
	if !v.availableOnVFO() {
		return
	}
	v.state[kind].offset += delta
	v.setIncrementalTuning(kind, v.state[kind].actualActive, v.state[kind].offset)
}

func (v *VFO) ToggleAvailableIncrementalTuning() {
	kind, ok := v.availableIncrementalTuningKind()
	if !ok {
		return
	}
	v.SetIncrementalTuningActive(kind, !v.state[kind].intentActive)
}

func (v *VFO) ShiftAvailableIncrementalTuning(sign core.Frequency) {
	kind, ok := v.availableIncrementalTuningKind()
	if !ok {
		return
	}
	step := core.DefaultXITShift
	if kind == core.RIT {
		step = core.DefaultRITShift
	}
	v.ShiftIncrementalTuning(kind, sign*step)
}

func (v *VFO) availableIncrementalTuningKind() (core.IncrementalTuningKind, bool) {
	if !v.availableOnVFO() {
		return -1, false
	}
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		if kind.Workmode() == v.workmode {
			return kind, true
		}
	}
	return -1, false
}

func (v *VFO) incrementalTuningVisible(kind core.IncrementalTuningKind) bool {
	if !v.availableOnVFO() {
		return false
	}
	return kind.Workmode() == v.workmode
}

func (v *VFO) WorkmodeChanged(vfo core.VFOID, workmode core.Workmode) {
	if vfo != v.id {
		return
	}
	v.workmode = workmode
	v.activate()
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		v.emitIncrementalTuningVisibilityChanged(kind)
	}
}

func (v *VFO) activate() {
	for _, kind := range []core.IncrementalTuningKind{core.RIT, core.XIT} {
		v.state[kind].actualActive = v.state[kind].intentActive && kind.Workmode() == v.workmode
		v.setIncrementalTuning(kind, v.state[kind].actualActive, v.state[kind].offset)
	}
}

func (v *VFO) VFOIncrementalTuningChanged(vfo core.VFOID, kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	if vfo != v.id {
		return
	}
	v.state[kind].actualActive = active
	v.state[kind].offset = offset
	v.offlineClient.SetIncrementalTuning(kind, active, offset)
}

func (v *VFO) IncrementalTuningAvailabilityChanged(vfo core.VFOID, kind core.IncrementalTuningKind, available bool) {
	if vfo != v.id {
		return
	}
	v.emitIncrementalTuningVisibilityChanged(kind)
}

func (v *VFO) emitIncrementalTuningChanged(kind core.IncrementalTuningKind, active bool, offset core.Frequency) {
	core.Emit(v.listeners, func(listener core.VFOIncrementalTuningListener) {
		v.asyncRunner(func() {
			listener.VFOIncrementalTuningChanged(v.id, kind, active, offset)
		})
	})
}

func (v *VFO) emitIncrementalTuningActiveChanged(kind core.IncrementalTuningKind, active bool) {
	core.Emit(v.listeners, func(l core.IncrementalTuningActiveListener) {
		v.asyncRunner(func() {
			l.IncrementalTuningActiveChanged(v.id, kind, active)
		})
	})
}

func (v *VFO) emitIncrementalTuningVisibilityChanged(kind core.IncrementalTuningKind) {
	visible := v.incrementalTuningVisible(kind)
	core.Emit(v.listeners, func(l core.IncrementalTuningVisibilityListener) {
		v.asyncRunner(func() {
			l.IncrementalTuningVisibilityChanged(v.id, kind, visible)
		})
	})
}
