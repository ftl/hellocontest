package workmode

import "github.com/ftl/hellocontest/core"

func NewController() *Controller {
	return &Controller{
		workmode: core.SearchPounce,
	}
}

type Controller struct {
	view      View
	listeners []interface{}

	workmode core.Workmode
}

// View represents the visual part of the workmode handling.
type View interface {
	SetWorkmode(core.Workmode)
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
