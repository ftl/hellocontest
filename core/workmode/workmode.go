package workmode

import "github.com/ftl/hellocontest/core"

func NewController(spPatterns, runPatterns []string) core.WorkmodeController {
	return &controller{
		workmode: core.SearchPounce,
		patterns: [][]string{
			spPatterns,
			runPatterns,
		},
	}
}

type controller struct {
	keyer core.KeyerController
	view  core.WorkmodeView

	workmode core.Workmode
	patterns [][]string
}

func (c *controller) SetView(view core.WorkmodeView) {
	c.view = view
	c.view.SetWorkmodeController(c)
	c.view.SetWorkmode(c.workmode)
}

func (c *controller) SetKeyer(keyer core.KeyerController) {
	c.keyer = keyer
	if c.keyer != nil {
		c.keyer.SetPatterns(c.patterns[c.workmode])
	}
}

func (c *controller) SetWorkmode(workmode core.Workmode) {
	if c.workmode == workmode {
		return
	}
	oldWorkmode := c.workmode
	c.workmode = workmode

	if c.view != nil {
		c.view.SetWorkmode(workmode)
	}

	if c.keyer != nil {
		for i := range c.patterns[oldWorkmode] {
			c.patterns[oldWorkmode][i] = c.keyer.GetPattern(i)
		}
		c.keyer.SetPatterns(c.patterns[c.workmode])
	}
}
