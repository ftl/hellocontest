package ui

import (
	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/qtc"
)

var _ qtc.View = (*qtcDialog)(nil)

type qtcDialog struct {
	parent     *qtlib.QMainWindow
	controller QTCController

	dialog *qtlib.QDialog
	view   *qtcView
	logBtn *qtlib.QPushButton

	stopKeyHandler *stopKeyHandler

	activePhase core.QTCWorkflowPhase
	activeField core.QTCField
	activeQTC   int

	completed bool
}

func newQTCDialog(parent *qtlib.QMainWindow, controller QTCController) *qtcDialog {
	return &qtcDialog{parent: parent, controller: controller}
}

func (d *qtcDialog) QuestionQTCCount(max int) (int, bool) {
	// TODO: implement a small modal QInputDialog-style picker.
	if max > 10 {
		max = 10
	}
	return max, true
}

func (d *qtcDialog) Show(mode core.QTCMode, series core.QTCSeries, vfoName string) {
	d.completed = false

	d.dialog = qtlib.NewQDialog(d.parent.QWidget)
	d.dialog.SetWindowTitle("QTC")
	d.dialog.SetModal(true)
	d.dialog.SetMinimumSize(qtlib.NewQSize2(600, 500))
	d.dialog.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)

	d.stopKeyHandler = newStopKeyHandler(d.dialog.QWidget)
	d.stopKeyHandler.SetStopKeyController(d.controller)

	d.view = newQTCView(d.controller, mode, vfoName)

	root := qtlib.NewQVBoxLayout(d.dialog.QWidget)
	root.AddWidget(d.view.root)

	buttons := qtlib.NewQDialogButtonBox4(
		qtlib.QDialogButtonBox__Ok | qtlib.QDialogButtonBox__Cancel,
	)
	d.logBtn = buttons.Button(qtlib.QDialogButtonBox__Ok)
	d.logBtn.SetText("Log")
	buttons.OnAccepted(func() {
		d.completed = true
		d.controller.CompleteQTCSeries()
		d.dialog.Accept()
	})
	buttons.OnRejected(func() {
		if d.completed {
			d.dialog.Reject()
			return
		}
		d.controller.AbortQTCSeries()
		d.dialog.Reject()
	})
	root.AddWidget(buttons.QWidget)

	d.view.setHeader(series.TheirCallsign, series.Header)
	d.view.setQTCs(series.QTCs)
	d.applyActivePhase()

	d.dialog.Exec()
	d.dialog.DeleteLater()
	d.dialog = nil
	d.view = nil
	d.logBtn = nil
	d.stopKeyHandler = nil
}

func (d *qtcDialog) Close() {
	if d.dialog == nil {
		return
	}
	d.completed = true
	d.dialog.Done(int(qtlib.QDialog__Accepted))
}

func (d *qtcDialog) UpdateQTC(index int, q core.QTC) {
	if d.view == nil {
		return
	}
	d.view.setQTC(index, q)
}

func (d *qtcDialog) ClearDataInputs() {
	if d.view == nil {
		return
	}
	d.view.clearExchangeEntry()
}

func (d *qtcDialog) ShowFieldError(field core.QTCField, message string) {
	if d.view == nil {
		return
	}
	d.view.setFieldError(field, message)
}

func (d *qtcDialog) ClearFieldError() {
	if d.view == nil {
		return
	}
	d.view.clearFieldError()
}

func (d *qtcDialog) SetActivePhase(phase core.QTCWorkflowPhase) {
	d.activePhase = phase
	d.applyActivePhase()
}

func (d *qtcDialog) applyActivePhase() {
	if d.view == nil {
		return
	}
	switch d.activePhase {
	case core.QTCStart:
		d.view.focusStart()
	case core.QTCExchangeHeader:
		d.view.focusHeader()
	case core.QTCExchangeData:
		d.view.focusData()
	case core.QTCFinish:
		d.view.focusNone()
		if d.logBtn != nil {
			d.logBtn.SetFocus()
		}
	default:
		d.view.focusNone()
	}
}

func (d *qtcDialog) SetActiveField(field core.QTCField) {
	d.activeField = field
	if d.view == nil {
		return
	}
	d.view.focusEntry(field)
}

func (d *qtcDialog) SetActiveQTC(index int) {
	d.activeQTC = index
	if d.view == nil {
		return
	}
	d.view.focusQTC(index)
}
