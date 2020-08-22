package callinfo

import "github.com/ftl/hellocontest/core"

func NewController() core.CallinfoController {
	return &callinfo{}
}

type callinfo struct {
	view core.CallinfoView
}

func (c *callinfo) SetView(view core.CallinfoView) {
	c.view = view
}

func (c *callinfo) Show() {
	if c.view == nil {
		return
	}
	c.view.Show()
}

func (c *callinfo) Hide() {
	if c.view == nil {
		return
	}
	c.view.Hide()
}
