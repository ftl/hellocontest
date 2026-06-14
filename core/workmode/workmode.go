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
	c.emitWorkmodeChanged(c.workmode)
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitWorkmodeChanged(workmode core.Workmode) {
	if c.view != nil {
		c.view.SetWorkmode(workmode)
	}
	for _, listener := range c.listeners {
		if workmodeChangedListener, ok := listener.(WorkmodeChangedListener); ok {
			workmodeChangedListener.WorkmodeChanged(core.VFO1, workmode)
		}
	}
}
