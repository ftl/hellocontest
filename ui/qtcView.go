package ui

import (
	"fmt"
	"strconv"

	qtlib "github.com/mappu/miqt/qt6"

	"github.com/ftl/hellocontest/core/model"
)

const (
	qtcSeriesName    = "series"
	qtcTimestampName = "timestamp"
	qtcCallsignName  = "callsign"
	qtcExchangeName  = "exchange"
)

// QTCController is the callback surface for the QTC dialog.
type QTCController interface {
	StartAction()
	ConfirmStart()
	HeaderAction()
	ConfirmHeader()
	DataAction()
	ConfirmData()

	DoAction()
	DoConfirm()
	RepeatLastTransmission()

	Enter(string)
	GotoNextField()
	SetActiveField(model.QTCField)

	CompleteQTCSeries()
	AbortQTCSeries()

	Stop()
	DoubleStop()
}

type qtcView struct {
	root       *qtlib.QWidget
	controller QTCController
	mode       model.QTCMode
	vfoName    string

	theirCallLabel    *qtlib.QLabel
	theirCallMessage  string
	startHeading      *qtlib.QLabel
	startActionButton *qtlib.QPushButton
	confirmStartBtn   *qtlib.QPushButton

	headerHeading      *qtlib.QLabel
	seriesEntry        *qtlib.QLineEdit
	headerActionButton *qtlib.QPushButton
	confirmHeaderBtn   *qtlib.QPushButton

	dataHeading      *qtlib.QLabel
	qtcTable         *qtlib.QTableWidget
	qtcTimeEntry     *qtlib.QLineEdit
	qtcCallEntry     *qtlib.QLineEdit
	qtcExchangeEntry *qtlib.QLineEdit
	dataActionButton *qtlib.QPushButton
	confirmDataBtn   *qtlib.QPushButton

	qtcs []model.QTC
}

func newQTCView(controller QTCController, mode model.QTCMode, vfoName string) *qtcView {
	v := &qtcView{controller: controller, mode: mode, vfoName: vfoName}

	var sendText string
	var modeText string
	if mode == model.ProvideQTC {
		sendText = "Send"
		modeText = "Offer QTC"
	} else {
		sendText = "AGN"
		modeText = "Request QTC"
	}

	v.root = qtlib.NewQWidget2()
	root := qtlib.NewQVBoxLayout(v.root)

	// Their-call header line
	v.theirCallLabel = qtlib.NewQLabel3("")
	v.theirCallLabel.SetAlignment(qtlib.AlignCenter)
	root.AddWidget(v.theirCallLabel.QWidget)

	// Phase 1: Start
	v.startHeading = qtlib.NewQLabel3("1. Start")
	v.startHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	root.AddWidget(v.startHeading.QWidget)
	startRow := qtlib.NewQWidget2()
	startLayout := qtlib.NewQHBoxLayout(startRow)
	startLayout.SetContentsMargins(0, 0, 0, 0)
	startLayout.AddWidget(qtlib.NewQLabel3(modeText).QWidget)
	startLayout.AddStretch()
	v.startActionButton = qtlib.NewQPushButton3("Send")
	v.startActionButton.OnClicked(controller.StartAction)
	startLayout.AddWidget(v.startActionButton.QWidget)
	v.confirmStartBtn = qtlib.NewQPushButton3("QRV")
	v.confirmStartBtn.OnClicked(controller.ConfirmStart)
	startLayout.AddWidget(v.confirmStartBtn.QWidget)
	root.AddWidget(startRow)

	// Phase 2: Header
	v.headerHeading = qtlib.NewQLabel3("2. Header")
	v.headerHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	root.AddWidget(v.headerHeading.QWidget)
	headerRow := qtlib.NewQWidget2()
	headerLayout := qtlib.NewQHBoxLayout(headerRow)
	headerLayout.SetContentsMargins(0, 0, 0, 0)
	headerLayout.AddWidget(qtlib.NewQLabel3("Series/QTC Count:").QWidget)
	v.seriesEntry = qtlib.NewQLineEdit2()
	setObjectName(v.seriesEntry, qtcSeriesName)
	v.seriesEntry.SetReadOnly(mode == model.ProvideQTC)
	v.attachEntryHandlers(v.seriesEntry, model.QTCHeaderField)
	headerLayout.AddWidget(v.seriesEntry.QWidget)
	v.headerActionButton = qtlib.NewQPushButton3(sendText)
	v.headerActionButton.OnClicked(controller.HeaderAction)
	headerLayout.AddWidget(v.headerActionButton.QWidget)
	v.confirmHeaderBtn = qtlib.NewQPushButton3("R")
	v.confirmHeaderBtn.OnClicked(controller.ConfirmHeader)
	headerLayout.AddWidget(v.confirmHeaderBtn.QWidget)
	root.AddWidget(headerRow)

	// Phase 3: QTCs
	v.dataHeading = qtlib.NewQLabel3("3. QTCs")
	v.dataHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	root.AddWidget(v.dataHeading.QWidget)

	v.qtcTable = qtlib.NewQTableWidget3(0, 5)
	v.qtcTable.SetHorizontalHeaderLabels([]string{"#", "Time", "Call", "Exch.", "Cfm."})
	v.qtcTable.SetEditTriggers(qtlib.QAbstractItemView__NoEditTriggers)
	v.qtcTable.SetSelectionMode(qtlib.QAbstractItemView__SingleSelection)
	v.qtcTable.SetFocusPolicy(qtlib.NoFocus)
	v.qtcTable.VerticalHeader().SetVisible(false)
	v.qtcTable.HorizontalHeader().SetHighlightSections(false)
	v.qtcTable.HorizontalHeader().SetDefaultAlignment(qtlib.AlignLeft | qtlib.AlignVCenter)
	root.AddWidget(v.qtcTable.QWidget)

	dataRow := qtlib.NewQWidget2()
	dataLayout := qtlib.NewQHBoxLayout(dataRow)
	dataLayout.SetContentsMargins(0, 0, 0, 0)
	if mode == model.ReceiveQTC {
		v.qtcTimeEntry = qtlib.NewQLineEdit2()
		setObjectName(v.qtcTimeEntry, qtcTimestampName)
		v.qtcTimeEntry.SetPlaceholderText("Time")
		v.attachEntryHandlers(v.qtcTimeEntry, model.QTCTimestampField)
		dataLayout.AddWidget(v.qtcTimeEntry.QWidget)

		v.qtcCallEntry = qtlib.NewQLineEdit2()
		setObjectName(v.qtcCallEntry, qtcCallsignName)
		v.qtcCallEntry.SetPlaceholderText("Call")
		v.attachEntryHandlers(v.qtcCallEntry, model.QTCCallsignField)
		dataLayout.AddWidget(v.qtcCallEntry.QWidget)

		v.qtcExchangeEntry = qtlib.NewQLineEdit2()
		setObjectName(v.qtcExchangeEntry, qtcExchangeName)
		v.qtcExchangeEntry.SetPlaceholderText("Exch.")
		v.attachEntryHandlers(v.qtcExchangeEntry, model.QTCExchangeField)
		dataLayout.AddWidget(v.qtcExchangeEntry.QWidget)
	} else {
		dataLayout.AddStretch()
	}
	v.dataActionButton = qtlib.NewQPushButton3(sendText)
	v.dataActionButton.OnClicked(controller.DataAction)
	dataLayout.AddWidget(v.dataActionButton.QWidget)
	v.confirmDataBtn = qtlib.NewQPushButton3("R")
	v.confirmDataBtn.OnClicked(controller.ConfirmData)
	dataLayout.AddWidget(v.confirmDataBtn.QWidget)
	root.AddWidget(dataRow)

	return v
}

func (v *qtcView) attachEntryHandlers(edit *qtlib.QLineEdit, field model.QTCField) {
	edit.OnTextChanged(func(text string) {
		v.controller.Enter(text)
	})
	edit.OnFocusInEvent(func(super func(ev *qtlib.QFocusEvent), ev *qtlib.QFocusEvent) {
		super(ev)
		v.controller.SetActiveField(field)
	})
	edit.OnKeyPressEvent(func(super func(ev *qtlib.QKeyEvent), ev *qtlib.QKeyEvent) {
		key := ev.Key()
		switch key {
		case int(qtlib.Key_Tab), int(qtlib.Key_Space):
			v.controller.GotoNextField()
		case int(qtlib.Key_Return), int(qtlib.Key_Enter):
			v.controller.DoConfirm()
		case int(qtlib.Key_Question):
			v.controller.DoAction()
		case int(qtlib.Key_Equal):
			v.controller.RepeatLastTransmission()
		default:
			super(ev)
		}
	})
}

func (v *qtcView) setHeader(theirCall model.Callsign, header model.QTCHeader) {
	var format string
	if v.mode == model.ProvideQTC {
		format = "Sending QTCs to %s"
	} else {
		format = "Receiving QTCs from %s"
	}
	if v.vfoName != "" {
		format += " on " + v.vfoName
	}
	v.theirCallMessage = fmt.Sprintf(format, theirCall.String())
	v.theirCallLabel.SetText(v.theirCallMessage)
	v.theirCallLabel.SetStyleSheet("")
	if header.QTCCount > 0 {
		v.seriesEntry.SetText(fmt.Sprintf("%d/%d", header.SeriesNumber, header.QTCCount))
	} else {
		v.seriesEntry.SetText("")
	}
}

func (v *qtcView) setFieldError(_ model.QTCField, message string) {
	v.theirCallLabel.SetStyleSheet(QTCFieldErrorStyle)
	v.theirCallLabel.SetText(message)
}

func (v *qtcView) clearFieldError() {
	v.theirCallLabel.SetStyleSheet("")
	v.theirCallLabel.SetText(v.theirCallMessage)
}

func (v *qtcView) setQTCs(qtcs []model.QTC) {
	v.qtcs = qtcs
	v.qtcTable.SetRowCount(0)
	for i, q := range qtcs {
		v.appendQTCRow(i, q)
	}
}

func (v *qtcView) setQTC(index int, qtc model.QTC) {
	if index < 0 {
		return
	}
	if index >= len(v.qtcs) {
		v.qtcs = append(v.qtcs, qtc)
		v.appendQTCRow(index, qtc)
		return
	}
	v.qtcs[index] = qtc
	v.fillQTCRow(index, qtc)
}

func (v *qtcView) appendQTCRow(index int, qtc model.QTC) {
	v.qtcTable.InsertRow(index)
	v.fillQTCRow(index, qtc)
}

func (v *qtcView) fillQTCRow(index int, qtc model.QTC) {
	v.qtcTable.SetItem(index, 0, qtlib.NewQTableWidgetItem2(strconv.Itoa(index+1)))
	v.qtcTable.SetItem(index, 1, qtlib.NewQTableWidgetItem2(qtc.QTCTime.String()))
	v.qtcTable.SetItem(index, 2, qtlib.NewQTableWidgetItem2(qtc.QTCCallsign.String()))
	v.qtcTable.SetItem(index, 3, qtlib.NewQTableWidgetItem2(qtc.QTCNumber.String()))
	cfm := ""
	if qtc.Confirmed {
		cfm = "✓"
	}
	v.qtcTable.SetItem(index, 4, qtlib.NewQTableWidgetItem2(cfm))
}

func (v *qtcView) clearExchangeEntry() {
	if v.qtcTimeEntry != nil {
		v.qtcTimeEntry.SetText("")
	}
	if v.qtcCallEntry != nil {
		v.qtcCallEntry.SetText("")
	}
	if v.qtcExchangeEntry != nil {
		v.qtcExchangeEntry.SetText("")
	}
}

func (v *qtcView) enableWidgets(phase model.QTCWorkflowPhase) {
	v.startActionButton.SetEnabled(phase == model.QTCStart)
	v.confirmStartBtn.SetEnabled(phase == model.QTCStart)

	v.seriesEntry.SetEnabled(phase == model.QTCExchangeHeader)
	v.headerActionButton.SetEnabled(phase == model.QTCExchangeHeader)
	v.confirmHeaderBtn.SetEnabled(phase == model.QTCExchangeHeader)

	v.qtcTable.SetEnabled(phase == model.QTCExchangeData)
	if v.qtcTimeEntry != nil {
		v.qtcTimeEntry.SetEnabled(phase == model.QTCExchangeData)
	}
	if v.qtcCallEntry != nil {
		v.qtcCallEntry.SetEnabled(phase == model.QTCExchangeData)
	}
	if v.qtcExchangeEntry != nil {
		v.qtcExchangeEntry.SetEnabled(phase == model.QTCExchangeData)
	}
	v.dataActionButton.SetEnabled(phase == model.QTCExchangeData)
	v.confirmDataBtn.SetEnabled(phase == model.QTCExchangeData)
}

func (v *qtcView) highlightPhase(phase model.QTCWorkflowPhase) {
	v.startHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	v.headerHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	v.dataHeading.SetStyleSheet(QTCPhaseInactiveStyle)
	switch phase {
	case model.QTCStart:
		v.startHeading.SetStyleSheet(QTCPhaseActiveStyle)
	case model.QTCExchangeHeader:
		v.headerHeading.SetStyleSheet(QTCPhaseActiveStyle)
	case model.QTCExchangeData:
		v.dataHeading.SetStyleSheet(QTCPhaseActiveStyle)
	}
}

func (v *qtcView) focusStart() {
	v.highlightPhase(model.QTCStart)
	v.enableWidgets(model.QTCStart)
	v.confirmStartBtn.SetFocus()
}

func (v *qtcView) focusHeader() {
	v.highlightPhase(model.QTCExchangeHeader)
	v.enableWidgets(model.QTCExchangeHeader)
	v.confirmHeaderBtn.SetFocus()
}

func (v *qtcView) focusData() {
	v.highlightPhase(model.QTCExchangeData)
	v.enableWidgets(model.QTCExchangeData)
	v.confirmDataBtn.SetFocus()
}

func (v *qtcView) focusNone() {
	v.highlightPhase(model.QTCNone)
	v.enableWidgets(model.QTCNone)
	v.qtcTable.ClearSelection()
}

func (v *qtcView) focusEntry(field model.QTCField) {
	switch field {
	case model.QTCHeaderField:
		v.seriesEntry.SetFocus()
	case model.QTCTimestampField:
		if v.qtcTimeEntry != nil {
			v.qtcTimeEntry.SetFocus()
		}
	case model.QTCCallsignField:
		if v.qtcCallEntry != nil {
			v.qtcCallEntry.SetFocus()
		}
	case model.QTCExchangeField:
		if v.qtcExchangeEntry != nil {
			v.qtcExchangeEntry.SetFocus()
		}
	}
}

func (v *qtcView) focusQTC(index int) {
	if index < 0 || index >= v.qtcTable.RowCount() {
		return
	}
	v.qtcTable.SetCurrentCell(index, 0)
	v.confirmDataBtn.SetFocus()
}
