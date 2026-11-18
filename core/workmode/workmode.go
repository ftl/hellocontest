package workmode

import (
	"github.com/ftl/hellocontest/core/model"
)

func NewController() *Controller {
	return &Controller{
		workmode: model.SearchPounce,
	}
}

type Controller struct {
	view      View
	listeners []any

	workmode            model.Workmode
	operationModeSprint bool
	vfo2Enabled         bool
	focusedVFO          model.VFOID
}

// View represents the visual part of the workmode handling.
type View interface {
	SetWorkmode(model.Workmode)
	SetOperationModeHint(hint string)
}

type WorkmodeChangedListener interface {
	WorkmodeChanged(vfo model.VFOID, workmode model.Workmode)
}

type WorkmodeChangedListenerFunc func(vfo model.VFOID, workmode model.Workmode)

func (f WorkmodeChangedListenerFunc) WorkmodeChanged(vfo model.VFOID, workmode model.Workmode) {
	f(vfo, workmode)
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.SetWorkmode(c.workmode)
	c.view.SetOperationModeHint(c.operationModeHint())
}

func (c *Controller) ContestChanged(contest model.Contest) {
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

func (c *Controller) QSOAdded(qso model.QSO) {
	if !c.operationModeSprint {
		return
	}

	c.SetWorkmode(c.nextWorkmode())
}

func (c *Controller) nextWorkmode() model.Workmode {
	switch c.workmode {
	case model.SearchPounce:
		return model.Run
	case model.Run:
		return model.SearchPounce
	default:
		return model.SearchPounce
	}
}

func (c *Controller) Workmode() model.Workmode {
	return c.workmode
}

func (c *Controller) SetWorkmode(workmode model.Workmode) {
	if workmode == model.UnknownWorkmode {
		workmode = model.SearchPounce
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

// RadioChanged implements model.RadioChangedListener.
func (c *Controller) RadioChanged(_ string, singleVFO bool) {
	c.vfo2Enabled = !singleVFO
	c.emitAllWorkmodes()
}

// FocusedVFOChanged implements model.FocusedVFOListener.
func (c *Controller) FocusedVFOChanged(vfo model.VFOID) {
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
	c.emitAllWorkmodes()
}

// EffectiveWorkmode returns the workmode for a given VFO.
// In SO2V (vfo2Enabled), VFO2 is always S&P; VFO1 uses the global workmode.
// Otherwise both VFOs use the global workmode.
func (c *Controller) EffectiveWorkmode(vfo model.VFOID) model.Workmode {
	if c.vfo2Enabled && vfo == model.VFO2 {
		return model.SearchPounce
	}
	return c.workmode
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

// emitAllWorkmodes notifies listeners about the effective workmode for each VFO.
// In single-VFO mode, only VFO1 is emitted. In SO2V, both VFOs are emitted.
func (c *Controller) emitAllWorkmodes() {
	c.emitWorkmodeChanged(model.VFO1, c.EffectiveWorkmode(model.VFO1))
	if c.vfo2Enabled {
		c.emitWorkmodeChanged(model.VFO2, c.EffectiveWorkmode(model.VFO2))
	}
}

func (c *Controller) emitWorkmodeChanged(vfo model.VFOID, workmode model.Workmode) {
	for _, listener := range c.listeners {
		if workmodeChangedListener, ok := listener.(WorkmodeChangedListener); ok {
			workmodeChangedListener.WorkmodeChanged(vfo, workmode)
		}
	}
}
