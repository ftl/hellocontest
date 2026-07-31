package workmode

import (
	"github.com/ftl/hellocontest/core"
)

func NewController() *Controller {
	return &Controller{
		workmode: core.SearchPounce,
	}
}

type Controller struct {
	view      View
	listeners []any

	workmode            core.Workmode
	operationModeSprint bool
	vfo2Enabled         bool
	focusedVFO          core.VFOID
}

// View represents the visual part of the workmode handling.
type View interface {
	SetWorkmode(core.Workmode)
	SetOperationModeHint(hint string)
}

type WorkmodeChangedListener interface {
	WorkmodeChanged(vfo core.VFOID, workmode core.Workmode)
}

type WorkmodeChangedListenerFunc func(vfo core.VFOID, workmode core.Workmode)

func (f WorkmodeChangedListenerFunc) WorkmodeChanged(vfo core.VFOID, workmode core.Workmode) {
	f(vfo, workmode)
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.SetWorkmode(c.workmode)
	c.view.SetOperationModeHint(c.operationModeHint())
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.operationModeSprint = contest.OperationModeSprint

	if c.view != nil {
		c.view.SetOperationModeHint(c.operationModeHint())
	}
}

func (c *Controller) operationModeHint() string {
	switch {
	case c.operationModeSprint:
		return "Sprint"
	default:
		return ""
	}
}

func (c *Controller) QSOAdded(qso core.QSO) {
	if !c.operationModeSprint {
		return
	}

	c.SetWorkmode(c.nextWorkmode())
}

func (c *Controller) nextWorkmode() core.Workmode {
	switch c.workmode {
	case core.SearchPounce:
		return core.Run
	case core.Run:
		return core.SearchPounce
	default:
		return core.SearchPounce
	}
}

func (c *Controller) Workmode() core.Workmode {
	return c.workmode
}

func (c *Controller) SetWorkmode(workmode core.Workmode) {
	if workmode == core.UnknownWorkmode {
		workmode = core.SearchPounce
	}
	if c.workmode == workmode {
		return
	}
	c.workmode = workmode
	if c.view != nil {
		c.view.SetWorkmode(c.workmode)
	}
	c.emitAllWorkmodes()
}

// RadioChanged implements core.RadioChangedListener.
func (c *Controller) RadioChanged(_ string, singleVFO bool) {
	c.vfo2Enabled = !singleVFO
	c.emitAllWorkmodes()
}

// FocusedVFOChanged implements core.FocusedVFOListener.
func (c *Controller) FocusedVFOChanged(vfo core.VFOID) {
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
	c.emitAllWorkmodes()
}

// EffectiveWorkmode returns the workmode for a given VFO.
// In SO2V (vfo2Enabled), VFO2 is always S&P; VFO1 uses the global workmode.
// Otherwise both VFOs use the global workmode.
func (c *Controller) EffectiveWorkmode(vfo core.VFOID) core.Workmode {
	if c.vfo2Enabled && vfo == core.VFO2 {
		return core.SearchPounce
	}
	return c.workmode
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

// emitAllWorkmodes notifies listeners about the effective workmode for each VFO.
// In single-VFO mode, only VFO1 is emitted. In SO2V, both VFOs are emitted.
func (c *Controller) emitAllWorkmodes() {
	c.emitWorkmodeChanged(core.VFO1, c.EffectiveWorkmode(core.VFO1))
	if c.vfo2Enabled {
		c.emitWorkmodeChanged(core.VFO2, c.EffectiveWorkmode(core.VFO2))
	}
}

func (c *Controller) emitWorkmodeChanged(vfo core.VFOID, workmode core.Workmode) {
	for _, listener := range c.listeners {
		if workmodeChangedListener, ok := listener.(WorkmodeChangedListener); ok {
			workmodeChangedListener.WorkmodeChanged(vfo, workmode)
		}
	}
}
