package ui

import (
	"strconv"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
	"github.com/gotk3/gotk3/gtk"
)

type summaryDialog struct {
	dialog *gtk.Dialog
	parent gtk.IWidget

	controller SummaryController
	view       *summaryView
	style      *style.Style

	contestName  string
	cabrilloName string
	startTime    time.Time
	callsign     string
	myExchanges  string

	operatorMode string
	overlay      string
	powerMode    string
	assisted     bool

	workedModes   string
	workedBands   string
	operatingTime time.Duration
	breakTime     time.Duration
	breaks        int

	score core.Score

	openAfterExport bool
}

func setupSummaryDialog(parent gtk.IWidget, controller SummaryController) *summaryDialog {
	result := &summaryDialog{
		parent:     parent,
		controller: controller,
	}
	return result
}

func (d *summaryDialog) onDestroy() {
	d.dialog = nil
	d.view = nil
}

func (d *summaryDialog) Show() bool {
	d.view = newSummaryView(d.controller)

	d.view.contestNameEntry.SetText(d.contestName)
	d.view.cabrilloNameEntry.SetText(d.cabrilloName)
	d.view.startTimeEntry.SetText(core.FormatTimestamp(d.startTime))
	d.view.callsignEntry.SetText(d.callsign)
	d.view.myExchangesEntry.SetText(d.myExchanges)

	d.view.workedModesEntry.SetText(d.workedModes)
	d.view.workedBandsEntry.SetText(d.workedBands)
	d.view.operatingTimeEntry.SetText(core.FormatDuration(d.operatingTime))
	d.view.breakTimeEntry.SetText(core.FormatDuration(d.breakTime))
	d.view.breaksEntry.SetText(strconv.Itoa(d.breaks))

	d.view.scoreTable.ShowScore(d.score)

	d.view.openAfterExportCheckButton.SetActive(d.openAfterExport)

	dialog, _ := gtk.DialogNew()
	d.dialog = dialog
	d.dialog.SetDefaultSize(400, 400)
	d.dialog.SetTransientFor(nil)
	d.dialog.SetPosition(gtk.WIN_POS_CENTER)
	d.dialog.Connect("destroy", d.onDestroy)
	d.dialog.SetTitle("Summary")
	d.dialog.SetDefaultResponse(gtk.RESPONSE_OK)
	d.dialog.SetModal(true)
	contentArea, _ := d.dialog.GetContentArea()
	contentArea.Add(d.view.root)
	d.dialog.AddButton("Export", gtk.RESPONSE_OK)
	d.dialog.AddButton("Close", gtk.RESPONSE_CANCEL)

	d.dialog.ShowAll()
	result := d.dialog.Run() == gtk.RESPONSE_OK
	d.dialog.Close()
	d.dialog.Destroy()
	d.dialog = nil
	d.view = nil

	return result
}

func (d *summaryDialog) SetContestName(contestName string) {
	d.contestName = contestName
	if d.view != nil {
		d.view.contestNameEntry.SetText(contestName)
	}
}

func (d *summaryDialog) SetCabrilloName(cabrilloName string) {
	d.cabrilloName = cabrilloName
	if d.view != nil {
		d.view.cabrilloNameEntry.SetText(cabrilloName)
	}
}

func (d *summaryDialog) SetStartTime(startTime time.Time) {
	d.startTime = startTime
	if d.view != nil {
		d.view.startTimeEntry.SetText(core.FormatTimestamp(startTime))
	}
}

func (d *summaryDialog) SetCallsign(callsign string) {
	d.callsign = callsign
	if d.view != nil {
		d.view.callsignEntry.SetText(callsign)
	}
}

func (d *summaryDialog) SetMyExchanges(myExchanges string) {
	d.myExchanges = myExchanges
	if d.view != nil {
		d.view.myExchangesEntry.SetText(myExchanges)
	}
}

func (d *summaryDialog) SetOperatorMode(operatorMode string) {
	d.operatorMode = operatorMode
	if d.view != nil {
		d.view.operatorModeCombo.SetActiveID(operatorMode)
	}
}
func (d *summaryDialog) SetOverlay(overlay string) {
	d.overlay = overlay
	if d.view != nil {
		d.view.overlayCombo.SetActiveID(overlay)
	}
}
func (d *summaryDialog) SetPowerMode(powerMode string) {
	d.powerMode = powerMode
	if d.view != nil {
		d.view.powerModeCombo.SetActiveID(powerMode)
	}
}
func (d *summaryDialog) SetAssisted(assisted bool) {
	d.assisted = assisted
	if d.view != nil {
		d.view.assistedCheckButton.SetActive(assisted)
	}
}

func (d *summaryDialog) SetWorkedModes(workedModes string) {
	d.workedModes = workedModes
	if d.view != nil {
		d.view.workedModesEntry.SetText(workedModes)
	}
}

func (d *summaryDialog) SetWorkedBands(workedBands string) {
	d.workedBands = workedBands
	if d.view != nil {
		d.view.workedBandsEntry.SetText(workedBands)
	}
}

func (d *summaryDialog) SetOperatingTime(operatingTime time.Duration) {
	d.operatingTime = operatingTime
	if d.view != nil {
		d.view.operatingTimeEntry.SetText(core.FormatDuration(operatingTime))
	}
}

func (d *summaryDialog) SetBreakTime(breakTime time.Duration) {
	d.breakTime = breakTime
	if d.view != nil {
		d.view.breakTimeEntry.SetText(core.FormatDuration(breakTime))
	}
}

func (d *summaryDialog) SetBreaks(breaks int) {
	d.breaks = breaks
	if d.view != nil {
		d.view.breaksEntry.SetText(strconv.Itoa(breaks))
	}
}

func (d *summaryDialog) SetScore(score core.Score) {
	d.score = score
	if d.view != nil {
		d.view.scoreTable.ShowScore(score)
	}
}

func (d *summaryDialog) SetOpenAfterExport(open bool) {
	d.openAfterExport = open
	if d.view != nil {
		d.view.openAfterExportCheckButton.SetActive(open)
	}
}
