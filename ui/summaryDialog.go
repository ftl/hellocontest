package ui

import (
	"strconv"
	"time"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/export/summary"
)

var _ summary.View = (*summaryDialog)(nil)

type summaryDialog struct {
	parent     *qtlib.QMainWindow
	controller SummaryController

	dialog *qtlib.QDialog
	view   *summaryView

	// Cached state pushed by the controller before Show is called.
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

	qtcsEnabled bool
	score       core.Score

	openAfterExport bool
}

func newSummaryDialog(parent *qtlib.QMainWindow, controller SummaryController) *summaryDialog {
	return &summaryDialog{parent: parent, controller: controller}
}

func (d *summaryDialog) Show() bool {
	d.view = newSummaryView(d.controller)

	d.view.doIgnore(func() {
		d.view.contestName.SetText(d.contestName)
		d.view.cabrilloName.SetText(d.cabrilloName)
		d.view.startTime.SetText(core.FormatTimestamp(d.startTime))
		d.view.callsign.SetText(d.callsign)
		d.view.myExchanges.SetText(d.myExchanges)

		d.view.workedModes.SetText(d.workedModes)
		d.view.workedBands.SetText(d.workedBands)
		d.view.operatingTime.SetText(core.FormatDuration(d.operatingTime))
		d.view.breakTime.SetText(core.FormatDuration(d.breakTime))
		d.view.breaks.SetText(strconv.Itoa(d.breaks))

		selectComboValue(d.view.operatorModeCombo, d.operatorMode)
		selectComboValue(d.view.overlayCombo, d.overlay)
		selectComboValue(d.view.powerModeCombo, d.powerMode)
		d.view.assisted.SetChecked(d.assisted)

		d.view.scoreTable.SetQTCsEnabled(d.qtcsEnabled)
		d.view.scoreTable.ShowScore(d.score)

		d.view.openAfterExport.SetChecked(d.openAfterExport)
	})

	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetWindowTitle("Summary")
	d.dialog.SetModal(true)
	d.dialog.SetMinimumSize(qtlib.NewQSize2(700, 500))
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	root := qtlib.NewQVBoxLayout(d.dialog.QWidget)
	root.AddWidget(d.view.root)

	buttons := qtlib.NewQDialogButtonBox4(
		qtlib.QDialogButtonBox__Ok | qtlib.QDialogButtonBox__Close,
	)
	buttons.Button(qtlib.QDialogButtonBox__Ok).SetText("Export")
	buttons.OnAccepted(func() { d.dialog.Accept() })
	buttons.OnRejected(func() { d.dialog.Reject() })
	root.AddWidget(buttons.QWidget)

	accepted := d.dialog.Exec() == int(qtlib.QDialog__Accepted)
	d.dialog.DeleteLater()
	d.dialog = nil
	d.view = nil
	return accepted
}

func selectComboValue(combo *qtlib.QComboBox, value string) {
	idx := combo.FindText(value)
	if idx < 0 {
		idx = 0
	}
	combo.SetCurrentIndex(idx)
}

func (d *summaryDialog) SetContestName(value string) {
	d.contestName = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.contestName.SetText(value) })
	}
}

func (d *summaryDialog) SetCabrilloName(value string) {
	d.cabrilloName = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.cabrilloName.SetText(value) })
	}
}

func (d *summaryDialog) SetStartTime(t time.Time) {
	d.startTime = t
	if d.view != nil {
		d.view.doIgnore(func() { d.view.startTime.SetText(core.FormatTimestamp(t)) })
	}
}

func (d *summaryDialog) SetCallsign(value string) {
	d.callsign = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.callsign.SetText(value) })
	}
}

func (d *summaryDialog) SetMyExchanges(value string) {
	d.myExchanges = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.myExchanges.SetText(value) })
	}
}

func (d *summaryDialog) SetOperatorMode(value string) {
	d.operatorMode = value
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.operatorModeCombo, value) })
	}
}

func (d *summaryDialog) SetOverlay(value string) {
	d.overlay = value
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.overlayCombo, value) })
	}
}

func (d *summaryDialog) SetPowerMode(value string) {
	d.powerMode = value
	if d.view != nil {
		d.view.doIgnore(func() { selectComboValue(d.view.powerModeCombo, value) })
	}
}

func (d *summaryDialog) SetAssisted(value bool) {
	d.assisted = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.assisted.SetChecked(value) })
	}
}

func (d *summaryDialog) SetWorkedModes(value string) {
	d.workedModes = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.workedModes.SetText(value) })
	}
}

func (d *summaryDialog) SetWorkedBands(value string) {
	d.workedBands = value
	if d.view != nil {
		d.view.doIgnore(func() { d.view.workedBands.SetText(value) })
	}
}

func (d *summaryDialog) SetOperatingTime(t time.Duration) {
	d.operatingTime = t
	if d.view != nil {
		d.view.doIgnore(func() { d.view.operatingTime.SetText(core.FormatDuration(t)) })
	}
}

func (d *summaryDialog) SetBreakTime(t time.Duration) {
	d.breakTime = t
	if d.view != nil {
		d.view.doIgnore(func() { d.view.breakTime.SetText(core.FormatDuration(t)) })
	}
}

func (d *summaryDialog) SetBreaks(n int) {
	d.breaks = n
	if d.view != nil {
		d.view.doIgnore(func() { d.view.breaks.SetText(strconv.Itoa(n)) })
	}
}

func (d *summaryDialog) SetQTCsEnabled(enabled bool) {
	d.qtcsEnabled = enabled
	if d.view != nil {
		d.view.scoreTable.SetQTCsEnabled(enabled)
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
		d.view.doIgnore(func() { d.view.openAfterExport.SetChecked(open) })
	}
}
