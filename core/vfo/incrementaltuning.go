package vfo

import (
	"github.com/ftl/hellocontest/core/model"
)

// Default step sizes for the relative shift actions. Keyboard shortcuts use
// these; the remote interface may override them via an explicit amount.
const (
	DefaultFrequencyShift = model.Frequency(10) // Hz
	DefaultXITShift       = model.Frequency(10) // Hz
	DefaultRITShift       = model.Frequency(10) // Hz
)

type incrementalTuningState struct {
	intentActive bool
	actualActive bool
	offset       model.Frequency
}

func (v *VFO) availableOnVFO() bool {
	if v.id == model.VFO1 {
		return true
	}
	if v.online() {
		return v.client.IncrementalTuningPerVFO()
	} else {
		return v.offlineClient.IncrementalTuningPerVFO()
	}
}

func (v *VFO) setIncrementalTuning(kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	if v.online() {
		v.client.SetIncrementalTuning(v.id, kind, active, offset)
	} else {
		v.offlineClient.SetIncrementalTuning(kind, active, offset)
	}
}

func (v *VFO) IncrementalTuningActive(kind model.IncrementalTuningKind) bool {
	if !v.availableOnVFO() {
		return false
	}
	return v.state[kind].intentActive
}

func (v *VFO) SetIncrementalTuningActive(kind model.IncrementalTuningKind, active bool) {
	if !v.availableOnVFO() {
		return
	}
	v.state[kind].intentActive = active
	v.activate()
	v.emitIncrementalTuningActiveChanged(kind, active)
}

func (v *VFO) ShiftIncrementalTuning(kind model.IncrementalTuningKind, delta model.Frequency) {
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

func (v *VFO) ShiftAvailableIncrementalTuning(sign model.Frequency) {
	kind, ok := v.availableIncrementalTuningKind()
	if !ok {
		return
	}
	step := DefaultXITShift
	if kind == model.RIT {
		step = DefaultRITShift
	}
	v.ShiftIncrementalTuning(kind, sign*step)
}

func (v *VFO) availableIncrementalTuningKind() (model.IncrementalTuningKind, bool) {
	if !v.availableOnVFO() {
		return -1, false
	}
	for _, kind := range []model.IncrementalTuningKind{model.RIT, model.XIT} {
		if kind.Workmode() == v.workmode {
			return kind, true
		}
	}
	return -1, false
}

func (v *VFO) incrementalTuningVisible(kind model.IncrementalTuningKind) bool {
	if !v.availableOnVFO() {
		return false
	}
	return kind.Workmode() == v.workmode
}

func (v *VFO) WorkmodeChanged(vfo model.VFOID, workmode model.Workmode) {
	if vfo != v.id {
		return
	}
	v.workmode = workmode
	v.activate()
	for _, kind := range []model.IncrementalTuningKind{model.RIT, model.XIT} {
		v.emitIncrementalTuningVisibilityChanged(kind)
	}
}

func (v *VFO) activate() {
	if !v.availableOnVFO() {
		return
	}
	for _, kind := range []model.IncrementalTuningKind{model.RIT, model.XIT} {
		v.state[kind].actualActive = v.state[kind].intentActive && kind.Workmode() == v.workmode
		v.setIncrementalTuning(kind, v.state[kind].actualActive, v.state[kind].offset)
	}
}

func (v *VFO) VFOIncrementalTuningChanged(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	if vfo != v.id {
		return
	}
	v.state[kind].actualActive = active
	v.state[kind].offset = offset
	v.offlineClient.SetIncrementalTuning(kind, active, offset)
}

func (v *VFO) emitIncrementalTuningChanged(kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	model.Emit(v.listeners, func(listener model.VFOIncrementalTuningListener) {
		v.asyncRunner(func() {
			listener.VFOIncrementalTuningChanged(v.id, kind, active, offset)
		})
	})
}

func (v *VFO) emitIncrementalTuningActiveChanged(kind model.IncrementalTuningKind, active bool) {
	model.Emit(v.listeners, func(l model.VFOIncrementalTuningActiveListener) {
		v.asyncRunner(func() {
			l.VFOIncrementalTuningActiveChanged(v.id, kind, active)
		})
	})
}

func (v *VFO) emitIncrementalTuningVisibilityChanged(kind model.IncrementalTuningKind) {
	visible := v.incrementalTuningVisible(kind)
	model.Emit(v.listeners, func(l model.VFOIncrementalTuningVisibilityListener) {
		v.asyncRunner(func() {
			l.VFOIncrementalTuningVisibilityChanged(v.id, kind, visible)
		})
	})
}
