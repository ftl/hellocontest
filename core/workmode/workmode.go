package workmode

import (
	"math"

	"github.com/ftl/hellocontest/core"
)

func NewController() *Controller {
	return &Controller{
		workmode: core.SearchPounce,
	}
}

type Controller struct {
	view      View
	listeners []interface{}

	workmode            core.Workmode
	operationModeSprint bool

	lastQSONumber int
}

// View represents the visual part of the workmode handling.
type View interface {
	SetWorkmode(core.Workmode)
	SetOperationModeHint(hint string)
}

type WorkmodeChangedListener interface {
	WorkmodeChanged(workmode core.Workmode)
}

type WorkmodeChangedListenerFunc func(workmode core.Workmode)

func (f WorkmodeChangedListenerFunc) WorkmodeChanged(workmode core.Workmode) {
	f(workmode)
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

func (c *Controller) RowAdded(qso core.QSO) {
	isNew := qso.MyNumber > core.QSONumber(c.lastQSONumber)
	c.lastQSONumber = int(math.Max(float64(c.lastQSONumber), float64(qso.MyNumber)))

	if !c.operationModeSprint || !isNew {
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
	if c.workmode == workmode {
		return
	}
	c.workmode = workmode
	c.emitWorkmodeChanged(c.workmode)
}

func (c *Controller) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitWorkmodeChanged(workmode core.Workmode) {
	if c.view != nil {
		c.view.SetWorkmode(workmode)
	}
	for _, listener := range c.listeners {
		if workmodeChangedListener, ok := listener.(WorkmodeChangedListener); ok {
			workmodeChangedListener.WorkmodeChanged(workmode)
		}
	}
}
