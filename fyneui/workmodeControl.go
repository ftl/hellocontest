package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/ftl/hellocontest/core"
)

const (
	workmodeSearchPounce = "Search & Pounce"
	workmodeRun          = "Run"
)

type WorkmodeController interface {
	SetWorkmode(core.Workmode)
}

type workmodeControl struct {
	container  *fyne.Container
	controller WorkmodeController

	workmodeRadioGroup *widget.RadioGroup
	operationModeLabel *widget.Label
}

func setupWorkmodeControl() *workmodeControl {
	result := &workmodeControl{}

	label := widget.NewLabel("Workmode:")

	result.workmodeRadioGroup = widget.NewRadioGroup(
		[]string{workmodeSearchPounce, workmodeRun},
		result.onWorkmodeChanged,
	)
	result.workmodeRadioGroup.Horizontal = true

	result.operationModeLabel = widget.NewLabel("")
	result.operationModeLabel.Hidden = true

	result.container = container.NewHBox(
		label,
		result.workmodeRadioGroup,
		layout.NewSpacer(),
		result.operationModeLabel,
	)

	return result
}

func (c *workmodeControl) onWorkmodeChanged(workmodeLabel string) {
	var workmode core.Workmode
	switch workmodeLabel {
	case workmodeSearchPounce:
		workmode = core.SearchPounce
	case workmodeRun:
		workmode = core.Run
	default:
		log.Printf("unknown workmode %s", workmodeLabel)
		return
	}
	c.controller.SetWorkmode(workmode)
}

func (c *workmodeControl) SetWorkmodeController(controller WorkmodeController) {
	c.controller = controller
}

func (c *workmodeControl) SetWorkmode(workmode core.Workmode) {
	switch workmode {
	case core.SearchPounce:
		c.workmodeRadioGroup.Selected = workmodeSearchPounce
	case core.Run:
		c.workmodeRadioGroup.Selected = workmodeRun
	default:
		c.workmodeRadioGroup.Selected = ""
	}
	c.workmodeRadioGroup.Refresh()
}

func (c *workmodeControl) SetOperationModeHint(hint string) {
	c.operationModeLabel.SetText(hint)
	c.operationModeLabel.Hidden = (hint == "")
	c.operationModeLabel.Refresh()
}
