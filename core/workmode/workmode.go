package workmode

import "github.com/ftl/hellocontest/core"

func NewController(spPatterns, runPatterns []string) *Controller {
	return &Controller{
		workmode: core.SearchPounce,
		patterns: [][]string{
			spPatterns,
			runPatterns,
		},
	}
}

type Controller struct {
	view      View
	keyer     Keyer
	listeners []interface{}

	workmode core.Workmode
	patterns [][]string
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

// Keyer functionality used by the workmode controller.
type Keyer interface {
	SetPatterns([]string)
	GetPattern(index int) string
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.view.SetWorkmode(c.workmode)
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
	if c.keyer != nil {
		c.keyer.SetPatterns(c.patterns[c.workmode])
	}
}

func (c *Controller) SetWorkmode(workmode core.Workmode) {
	if c.workmode == workmode {
		return
	}
	oldWorkmode := c.workmode
	c.workmode = workmode

	if c.keyer != nil {
		for i := range c.patterns[oldWorkmode] {
			c.patterns[oldWorkmode][i] = c.keyer.GetPattern(i)
		}
		c.keyer.SetPatterns(c.patterns[c.workmode])
	}

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
