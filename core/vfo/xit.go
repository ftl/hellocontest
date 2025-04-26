package vfo

import (
	"log"

	"github.com/ftl/hellocontest/core"
)

type XITControl struct {
	active      bool
	activeOnVFO bool
	workmode    core.Workmode

	offset core.Frequency

	vfo *VFO
}

func (x *XITControl) XITActive() bool {
	return x.active
}

func (x *XITControl) SetXITActive(active bool) {
	x.active = active
	x.activateOnVFO()
}

func (x *XITControl) WorkmodeChanged(workmode core.Workmode) {
	x.workmode = workmode
	if x.active {
		x.activateOnVFO()
	}
}

func (x *XITControl) activateOnVFO() {
	x.activeOnVFO = x.active && (x.workmode == core.SearchPounce)
	x.vfo.SetXIT(x.activeOnVFO, x.offset)
}

func (x *XITControl) VFOXITChanged(active bool, offset core.Frequency) {
	x.activeOnVFO = active
	x.offset = offset

	shouldBeActive := x.active && (x.workmode == core.SearchPounce)
	if shouldBeActive && !x.activeOnVFO {
		log.Printf("XITControl: XIT turned off by user")
	}
}
