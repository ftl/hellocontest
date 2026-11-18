package entry

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core/model"
)

const (
	commandPrefix     = "@"   // goes to a callsign, a numbered marker, or the CQ marker
	cqCommand         = "@cq" // goes to the frequency of the CQ marker
	textMarkerCommand = ":"   // followed by a text, marks the frequency with that text
)

// View represents the visual part of the QSO data entry.
type View interface {
	SetMyExchange(int, string)

	SetFrequency(model.VFOID, model.Frequency)
	SetBand(vfo model.VFOID, text string)
	SetMode(vfo model.VFOID, text string)
	SetTXState(vfo model.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration)

	SetCallsign(model.VFOID, string)
	SetTheirExchange(vfo model.VFOID, index int, text string)

	SetSerialClaim(model.VFOID, model.QSONumber, bool)
	SetActiveVFO(model.VFOID)
	SetActiveField(model.VFOID, model.EntryField)
	SelectText(model.VFOID, model.EntryField, string)
	SetDuplicateMarker(model.VFOID, bool)
	SetEditingMarker(model.VFOID, bool)
	ShowMessage(model.VFOID, ...any)
	ClearMessage(model.VFOID)
	SetVFOEnabled(model.VFOID, bool)
	SetVFOWorkmode(model.VFOID, model.Workmode)
	SetTXVFO(model.VFOID)
}

type input struct {
	callsign      string
	theirReport   string
	theirNumber   string
	theirExchange []string
	band          string
	mode          string
}

// Logbook functionality used for QSO entry.
type Logbook interface {
	NextQSONumber() model.QSONumber
	LastBand() model.Band
	LastMode() model.Mode
	LastExchange() []string
	AddQSO(model.QSO)
	UpdateQSO(model.QSO)
	FindDuplicateQSOs(model.Callsign, model.Band, model.Mode) []model.QSO
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendMacro(index int)
	SendQuestion(q string)
	GetText(workmode model.Workmode, index int) (string, error)
	SendText(text string, args ...any)
	Repeat()
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	InputChanged(vfo model.VFOID, call string, band model.Band, mode model.Mode, exchange []string)
}

type Bandmap interface {
	Add(model.Spot)
	SelectByCallsign(model.Callsign)
	GotoCQMarker()
	GotoNumberedMarker(int)
	MarkWithNextNumber(model.Frequency, model.Band, model.MarkerInput)
	MarkWithNumber(int, model.Frequency, model.Band, model.MarkerInput)
	MarkWithText(string, model.Frequency, model.Band)
}

// NewController returns a new entry controller.
func NewController(settings model.Settings, clock model.Clock, logbook Logbook, qsoList QSOList, bandmap Bandmap, asyncRunner model.AsyncRunner) *Controller {
	result := &Controller{
		clock:       clock,
		view:        new(nullView),
		logbook:     logbook,
		qsoList:     qsoList,
		callinfo:    new(nullCallinfo),
		asyncRunner: asyncRunner,
		bandmap:     bandmap,
		esmView:     new(nullESMView),
		vfoSwitcher: new(nullVFOSwitcher),

		vfos:                 make([]model.VFO, model.VFOCount),
		input:                make([]input, model.VFOCount),
		selectedFrequency:    make([]model.Frequency, model.VFOCount),
		selectedBand:         make([]model.Band, model.VFOCount),
		selectedMode:         make([]model.Mode, model.VFOCount),
		activeField:          make([]model.EntryField, model.VFOCount),
		errorField:           make([]model.EntryField, model.VFOCount),
		currentCallinfoFrame: make([]model.CallinfoFrame, model.VFOCount),
		claims:               newSerialClaims(),
		esmState:             make([]model.ESMState, model.VFOCount),
		esmMessage:           make([]string, model.VFOCount),
		esmMacroIndex:        make([]int, model.VFOCount),
		lastNumber:           make([]model.QSONumber, model.VFOCount),

		stationCallsign: settings.Station().Callsign.String(),
		vfoWorkmode:     make([]model.Workmode, model.VFOCount),
	}
	for vfo := range len(result.vfos) {
		result.vfos[vfo] = new(nullVFO)
		result.activeField[vfo] = model.CallsignField
	}
	result.updateExchangeFields(settings.Contest())
	return result
}

// VFOSwitcher is implemented by something that can command the rig to make a given VFO the current one.
type VFOSwitcher interface {
	SetCurrentVFO(model.VFOID)
}

type nullVFOSwitcher struct{}

func (n *nullVFOSwitcher) SetCurrentVFO(model.VFOID) {}

type editSnapshot struct {
	focusedVFO     model.VFOID
	input          input
	myReport       string
	myNumber       string
	myExchange     []string
	claimedSerial  model.QSONumber
	claimSnapshot  model.QSONumber
	claimCommitted bool
	activeField    model.EntryField
	errorField     model.EntryField
	callinfoFrame  model.CallinfoFrame
	esmState       []model.ESMState
	esmMessage     []string
	esmMacroIndex  []int
}

type Controller struct {
	clock           model.Clock
	view            View
	logbook         Logbook
	qsoList         QSOList
	keyer           Keyer
	callinfo        Callinfo
	vfos            []model.VFO
	vfoSwitcher     VFOSwitcher
	ignoreVFOChange bool
	bandmap         Bandmap
	esmView         ESMView

	asyncRunner model.AsyncRunner
	listeners   []any

	stationCallsign string
	workmode        model.Workmode
	vfoWorkmode     []model.Workmode

	myExchangeFields         []model.ExchangeField
	theirExchangeFields      []model.ExchangeField
	myReportExchangeField    model.ExchangeField
	myNumberExchangeField    model.ExchangeField
	theirReportExchangeField model.ExchangeField
	theirNumberExchangeField model.ExchangeField
	generateSerialExchange   bool
	generateReport           bool
	defaultExchangeValues    []string
	currentCallinfoFrame     []model.CallinfoFrame

	input              []input
	myReport           string
	myNumber           string
	myExchange         []string
	lastNumber         []model.QSONumber
	lastExchange       []string
	focusedVFO         model.VFOID
	txVFO              model.VFOID
	vfo2Enabled        bool
	activeField        []model.EntryField
	errorField         []model.EntryField
	selectedFrequency  []model.Frequency
	selectedBand       []model.Band
	selectedMode       []model.Mode
	claims             SerialClaims
	editing            bool
	editQSO            model.QSO
	editSnapshot       *editSnapshot
	ignoreQSOSelection bool

	ptt            bool
	parrotActive   bool
	parrotTimeLeft time.Duration

	esmEnabled    bool
	esmState      []model.ESMState
	esmMessage    []string
	esmMacroIndex []int
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitFocusedVFOChanged(vfo model.VFOID) {
	for _, l := range c.listeners {
		if listener, ok := l.(model.FocusedVFOListener); ok {
			listener.FocusedVFOChanged(vfo)
		}
	}
}

func (c *Controller) emitCallsignEntered(callsign string) {
	for _, l := range c.listeners {
		if listener, ok := l.(model.CallsignEnteredListener); ok {
			listener.CallsignEntered(c.focusedVFO, callsign)
		}
	}
}

func (c *Controller) emitCallsignLogged(callsign string, frequency model.Frequency) {
	for _, l := range c.listeners {
		if listener, ok := l.(model.CallsignLoggedListener); ok {
			listener.CallsignLogged(callsign, frequency)
		}
	}
}

func (c *Controller) SetView(view View) {
	if view == nil {
		panic("entry.Controller.SetView must not be called with nil")
	}
	if _, ok := c.view.(*nullView); !ok {
		panic("entry.Controller.SetView was already called")
	}

	c.view = view
	c.view.SetVFOEnabled(model.VFO2, c.vfo2Enabled)
	c.view.SetActiveVFO(c.focusedVFO)
	for vfo := range model.VFOCount {
		c.view.SetVFOWorkmode(model.VFOID(vfo), c.vfoWorkmode[vfo])
	}
	c.Clear()
}

func (c *Controller) LogbookLoaded() {
	lastBand := c.logbook.LastBand()
	lastMode := c.logbook.LastMode()
	for vfo := range model.VFOCount {
		c.selectedBand[vfo] = lastBand
		c.selectedMode[vfo] = lastMode
	}
	c.Clear()
	c.showInput()
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) CallinfoFrameChanged(vfo model.VFOID, frame model.CallinfoFrame) {
	c.currentCallinfoFrame[vfo] = frame
	// TODO what do we need to update here?
}

func (c *Controller) SetVFO(id model.VFOID, vfo model.VFO) {
	if vfo == nil {
		c.vfos[id] = new(nullVFO)
	} else {
		c.vfos[id] = vfo
	}
	vfo.Notify(c)
}

// ShiftFrequency changes the frequency of the currently focused VFO by delta.
func (c *Controller) ShiftFrequency(delta model.Frequency) {
	c.vfos[c.focusedVFO].ShiftFrequency(delta)
}

func (c *Controller) GotoNextField() model.EntryField {
	return c.gotoNextField(false)
}

func (c *Controller) GotoNextFieldSkippingReport() model.EntryField {
	return c.gotoNextField(true)
}

func (c *Controller) gotoNextField(skipReport bool) model.EntryField {
	switch c.activeField[c.focusedVFO] {
	case model.CallsignField:
		c.leaveCallsignField()
	}

	transitions := map[model.EntryField]model.EntryField{
		model.BandField: model.CallsignField,
		model.ModeField: model.CallsignField,
	}
	if len(c.theirExchangeFields) > 0 {
		transitions[model.CallsignField] = model.TheirExchangeField(1)
	}
	for _, field := range c.myExchangeFields {
		transitions[field.Field] = model.CallsignField
	}
	for i, field := range c.theirExchangeFields {
		if i == len(c.theirExchangeFields)-1 {
			transitions[field.Field] = model.CallsignField
		} else {
			transitions[field.Field] = field.Field.NextExchangeField()
		}
	}

	nextField := nextFieldOf(transitions, c.activeField[c.focusedVFO])
	reportField := c.theirReportExchangeField.Field
	if skipReport && reportField != "" && nextField == reportField {
		nextField = nextFieldOf(transitions, nextField)
	}

	c.SetActiveField(nextField)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
	return c.activeField[c.focusedVFO]
}

func nextFieldOf(transitions map[model.EntryField]model.EntryField, field model.EntryField) model.EntryField {
	next := transitions[field]
	if next == "" {
		return model.CallsignField
	}
	return next
}

func (c *Controller) GotoNextPlaceholder() {
	c.SetActiveField(model.CallsignField)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
	c.view.SelectText(c.focusedVFO, c.activeField[c.focusedVFO], model.FilterPlaceholder)
}

func (c *Controller) leaveCallsignField() {
	callsign, err := model.ParseCallsign(c.input[c.focusedVFO].callsign)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(c.input[c.focusedVFO].theirExchange) == len(c.currentCallinfoFrame[c.focusedVFO].PredictedExchange) {
		for i, field := range c.theirExchangeFields {
			if !c.isPredictable(field.Field) {
				continue
			}
			if c.input[c.focusedVFO].theirExchange[i] == "" {
				c.setTheirExchangePrediction(i, c.currentCallinfoFrame[c.focusedVFO].PredictedExchange[i])
			}
		}
	}

	_, found := c.isDuplicate(c.focusedVFO, callsign)
	if !found {
		c.view.SetDuplicateMarker(c.focusedVFO, false)
		return
	}
	if c.editing {
		c.view.SetDuplicateMarker(c.focusedVFO, c.editQSO.Callsign != callsign)
		return
	}

	c.view.SetDuplicateMarker(c.focusedVFO, true)
}

// isPredictable returns true if the exchange for the given field is predictable.
func (c *Controller) isPredictable(field model.EntryField) bool {
	switch field {
	case c.theirReportExchangeField.Field:
		return false
	case c.theirNumberExchangeField.Field:
		if len(c.theirNumberExchangeField.Properties) == 1 {
			return false
		}
	}
	return true
}

func (c *Controller) RefreshPrediction() {
	c.callinfoInputChanged(c.focusedVFO, c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], []string{})

	if len(c.input[c.focusedVFO].theirExchange) == len(c.currentCallinfoFrame[c.focusedVFO].PredictedExchange) {
		for i, field := range c.theirExchangeFields {
			if !c.isPredictable(field.Field) {
				continue
			}
			c.setTheirExchangePrediction(i, c.currentCallinfoFrame[c.focusedVFO].PredictedExchange[i])
		}
	}
}

func (c *Controller) RefreshView() {
	c.showInput()
}

func (c *Controller) showQSO(qso model.QSO) {
	c.input[model.VFO1].callsign = qso.Callsign.String()
	c.input[model.VFO1].theirReport = qso.TheirReport.String()
	c.input[model.VFO1].theirNumber = qso.TheirNumber.String()
	c.input[model.VFO1].theirExchange = ensureLen(qso.TheirExchange, len(c.theirExchangeFields))
	c.myReport = qso.MyReport.String()
	c.myNumber = qso.MyNumber.String()
	c.myExchange = ensureLen(qso.MyExchange, len(c.myExchangeFields))
	c.input[model.VFO1].band = qso.Band.String()
	c.input[model.VFO1].mode = qso.Mode.String()

	c.selectedFrequency[model.VFO1] = qso.Frequency
	c.selectedBand[model.VFO1] = qso.Band
	c.selectedMode[model.VFO1] = qso.Mode

	c.showInput()
}

func ensureLen(a []string, l int) []string {
	if len(a) < l {
		return append(a, make([]string, l-len(a))...)
	}
	if len(a) > l {
		return a[:l]
	}
	return a
}

func (c *Controller) showInput() {
	for vfo := range model.VFOCount {
		v := model.VFOID(vfo)
		c.view.SetCallsign(v, c.input[vfo].callsign)
		for i, value := range c.input[vfo].theirExchange {
			c.view.SetTheirExchange(v, i+1, value)
		}
		c.view.SetFrequency(v, c.selectedFrequency[vfo])
		c.view.SetBand(v, c.input[vfo].band)
		c.view.SetMode(v, c.input[vfo].mode)
	}
	// myExchange remains shared (single row in the UI).
	for i, value := range c.myExchange {
		c.view.SetMyExchange(i+1, value)
	}
}

// setTheirExchangePrediction replaces the value of the given field with the given predicted value,
// if the given value is not empty.
func (c *Controller) setTheirExchangePrediction(i int, value string) {
	if value == "" {
		return
	}
	c.input[c.focusedVFO].theirExchange[i] = value
	c.view.SetTheirExchange(c.focusedVFO, i+1, value)
}

func (c *Controller) isDuplicate(vfo model.VFOID, callsign model.Callsign) (model.QSO, bool) {
	qsos := c.logbook.FindDuplicateQSOs(callsign, c.selectedBand[vfo], c.selectedMode[vfo])
	if len(qsos) == 0 {
		return model.QSO{}, false
	}
	return qsos[len(qsos)-1], true
}

// SetVFOSwitcher wires the focus actions to a backend that can command the rig.
// If never called, focus actions still update internal state but do not retune the rig.
func (c *Controller) SetVFOSwitcher(switcher VFOSwitcher) {
	if switcher == nil {
		c.vfoSwitcher = new(nullVFOSwitcher)
		return
	}
	c.vfoSwitcher = switcher
}

func (c *Controller) CurrentVFOChanged(vfo model.VFOID) {
	if c.ignoreVFOChange {
		return
	}
	if vfo == model.VFO2 && !c.vfo2Enabled {
		return
	}
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
	c.emitFocusedVFOChanged(c.focusedVFO)
	c.refreshMyNumberInputs()
	c.view.SetActiveVFO(c.focusedVFO)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
}

// SetFocusedVFO is the single funnel for changing focused VFO.
func (c *Controller) SetFocusedVFO(vfo model.VFOID) {
	if vfo == model.VFO2 && !c.vfo2Enabled {
		return
	}
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
	c.ignoreVFOChange = true
	c.vfoSwitcher.SetCurrentVFO(c.focusedVFO)
	c.ignoreVFOChange = false
	c.emitFocusedVFOChanged(c.focusedVFO)
	c.refreshMyNumberInputs()
	c.view.SetActiveVFO(c.focusedVFO)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
}

// setFocusedVFOSilent updates focusedVFO without commanding the rig. Used by edit mode.
func (c *Controller) setFocusedVFOSilent(vfo model.VFOID) {
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
}

// ToggleFocusedVFO flips between VFO1 and VFO2. No-op if VFO2 is disabled.
func (c *Controller) ToggleFocusedVFO() {
	if !c.vfo2Enabled {
		c.SetFocusedVFO(model.VFO1)
		return
	}
	if c.focusedVFO == model.VFO1 {
		c.SetFocusedVFO(model.VFO2)
	} else {
		c.SetFocusedVFO(model.VFO1)
	}
}

// FocusVFO1 sets the focused VFO to VFO1.
func (c *Controller) FocusVFO1() {
	c.SetFocusedVFO(model.VFO1)
}

// FocusVFO2 sets the focused VFO to VFO2. No-op if VFO2 is disabled.
func (c *Controller) FocusVFO2() {
	c.SetFocusedVFO(model.VFO2)
}

// LogVFO synthesises focus on vfo and then logs the QSO from that row.
func (c *Controller) LogVFO(vfo model.VFOID) {
	c.SetFocusedVFO(vfo)
	c.Log()
}

// ClearVFO synthesises focus on vfo and then clears that row.
func (c *Controller) ClearVFO(vfo model.VFOID) {
	c.SetFocusedVFO(vfo)
	c.Clear()
}

// RadioChanged toggles VFO2 availability based on the connected radio's single-VFO flag.
// Implements model.RadioChangedListener.
func (c *Controller) RadioChanged(_ string, singleVFO bool) {
	c.vfo2Enabled = !singleVFO
	if !c.vfo2Enabled && c.focusedVFO == model.VFO2 {
		c.releaseSerialClaimFor(model.VFO2)
		c.input[model.VFO2] = input{}
		c.setFocusedVFOSilent(model.VFO1)
	}
	c.view.SetVFOEnabled(model.VFO2, c.vfo2Enabled)
}

// canTransmit reports whether the keyer is currently allowed to transmit. False during edit mode.
func (c *Controller) canTransmit() bool {
	return !c.editing
}

// SetTXVFO commands the rig to switch the TX VFO, suppressing any rig-side
// VFO change echo that might arrive from hamlib polling.

// claimSerialFor reserves the next unclaimed serial for the given VFO if it has none yet.
// Sticky: subsequent calls while a claim exists are no-ops.
func (c *Controller) claimSerialFor(vfo model.VFOID) {
	c.claims.ClaimNext(vfo, c.logbook.NextQSONumber())
	c.refreshMyNumberInputs()
}

// releaseSerialClaimFor frees the claim slot for vfo.
// If the serial was committed (sent over air), it is burned and cannot be reused.
func (c *Controller) releaseSerialClaimFor(vfo model.VFOID) {
	c.claims.Release(vfo)
	c.refreshMyNumberInputs()
}

// SerialSent is called when the keyer transmits a message containing a serial number.
// It claims the serial for the focused VFO (if not already claimed) and commits it.
func (c *Controller) SerialSent() {
	c.claimSerialFor(c.focusedVFO)
	c.claims.Commit(c.focusedVFO)
	c.refreshMyNumberInputs()
}

// IsSerialCommitted reports whether the serial claim for vfo has been committed (sent over air).
func (c *Controller) IsSerialCommitted(vfo model.VFOID) bool {
	return c.claims.committed[vfo]
}

// refreshMyNumberInputs syncs myNumber (and exchange serial slot) with the
// current displayed serial value for the focused VFO. Reads NextQSONumber once.
// Also pushes each VFO's serial claim to the view.
func (c *Controller) refreshMyNumberInputs() {
	base := c.logbook.NextQSONumber()
	c.writeMyNumberInput(c.claims.DisplayedSerial(c.focusedVFO, base))
	for vfo := range model.VFOCount {
		c.view.SetSerialClaim(model.VFOID(vfo), c.claims.claimed[vfo], c.claims.committed[vfo])
	}
}

func (c *Controller) refreshMyNumberInput(vfo model.VFOID) {
	base := c.logbook.NextQSONumber()
	c.writeMyNumberInput(c.claims.DisplayedSerial(vfo, base))
}

func (c *Controller) writeMyNumberInput(serial model.QSONumber) {
	value := serial.String()
	c.myNumber = value
	i := c.myNumberExchangeField.Field.ExchangeIndex() - 1
	if i < 0 || !c.generateSerialExchange {
		return
	}
	if i >= len(c.myExchange) {
		return
	}
	c.myExchange[i] = value
}

// enterEditMode snapshots VFO1's state, force-focuses VFO1 silently, marks editing,
// and claims the QSO's existing serial for VFO1 for the duration of the edit.
func (c *Controller) enterEditMode(qso model.QSO) {
	c.editSnapshot = &editSnapshot{
		focusedVFO:     c.focusedVFO,
		input:          c.input[model.VFO1],
		myReport:       c.myReport,
		myNumber:       c.myNumber,
		myExchange:     append([]string(nil), c.myExchange...),
		claimedSerial:  c.claims.claimed[model.VFO1],
		claimSnapshot:  c.claims.snapshot[model.VFO1],
		claimCommitted: c.claims.committed[model.VFO1],
		activeField:    c.activeField[model.VFO1],
		errorField:     c.errorField[model.VFO1],
		callinfoFrame:  c.currentCallinfoFrame[model.VFO1],
		esmState:       append([]model.ESMState(nil), c.esmState...),
		esmMessage:     append([]string(nil), c.esmMessage...),
		esmMacroIndex:  append([]int(nil), c.esmMacroIndex...),
	}
	c.setFocusedVFOSilent(model.VFO1)
	c.editing = true
	c.editQSO = qso
	c.claims.claimed[model.VFO1] = qso.MyNumber
	c.claims.snapshot[model.VFO1] = c.logbook.NextQSONumber()
	c.claims.committed[model.VFO1] = false
	// TODO step 6: c.view.SetVFOEnabled(model.VFO2, false)
}

// leaveEditMode restores the pre-edit state captured by enterEditMode. No-op if not editing.
func (c *Controller) leaveEditMode() {
	if c.editSnapshot == nil {
		return
	}
	snap := c.editSnapshot
	c.input[model.VFO1] = snap.input
	c.myReport = snap.myReport
	c.myNumber = snap.myNumber
	c.myExchange = append(c.myExchange[:0], snap.myExchange...)
	c.claims.claimed[model.VFO1] = snap.claimedSerial
	c.claims.snapshot[model.VFO1] = snap.claimSnapshot
	c.claims.committed[model.VFO1] = snap.claimCommitted
	c.activeField[model.VFO1] = snap.activeField
	c.errorField[model.VFO1] = snap.errorField
	c.currentCallinfoFrame[model.VFO1] = snap.callinfoFrame
	copy(c.esmState, snap.esmState)
	copy(c.esmMessage, snap.esmMessage)
	copy(c.esmMacroIndex, snap.esmMacroIndex)
	c.editing = false
	c.editQSO = model.QSO{}
	c.editSnapshot = nil
	c.setFocusedVFOSilent(snap.focusedVFO)
	// TODO step 6: c.view.SetVFOEnabled(model.VFO2, true)
}

func (c *Controller) SetActiveField(field model.EntryField) {
	c.activeField[c.focusedVFO] = field
	c.updateESM()
}

func (c *Controller) SelectMatch(index int) {
	c.selectCallsign(c.currentCallinfoFrame[c.focusedVFO].GetMatch(index))
}

func (c *Controller) SelectBestMatchOnFrequency() {
	c.selectCallsign(c.currentCallinfoFrame[c.focusedVFO].BestMatchOnFrequency().Callsign.String())
}

func (c *Controller) selectCallsign(callsign string) {
	if callsign == "" {
		return
	}
	c.SetActiveField(model.CallsignField)
	c.Enter(callsign)
	c.view.SetCallsign(c.focusedVFO, c.input[c.focusedVFO].callsign)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
}

func (c *Controller) Enter(text string) {
	switch c.activeField[c.focusedVFO] {
	case model.CallsignField:
		c.input[c.focusedVFO].callsign = text
		c.enterCallsign(text)
	case model.BandField:
		c.input[c.focusedVFO].band = text
		c.bandSelected(text)
	case model.ModeField:
		c.input[c.focusedVFO].mode = text
		c.modeSelected(text)
	}

	i := c.activeField[c.focusedVFO].ExchangeIndex() - 1
	switch {
	case c.activeField[c.focusedVFO].IsMyExchange():
		c.myExchange[i] = text
	case c.activeField[c.focusedVFO].IsTheirExchange():
		c.input[c.focusedVFO].theirExchange[i] = text
		c.enterTheirExchange(c.activeField[c.focusedVFO])
	}

	c.updateESM()
}

func (c *Controller) frequencyEntered(frequency model.Frequency) {
	// log.Printf("Frequency selected: %s", frequency)
	c.vfos[c.focusedVFO].SetFrequency(frequency)
}

func (c *Controller) bandEntered(band model.Band) {
	c.input[c.focusedVFO].band = band.String()
	c.vfos[c.focusedVFO].SetBand(band)
}

func (c *Controller) VFOFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	c.asyncRunner(func() {
		if vfo == model.VFO1 && c.editing {
			return
		}
		if c.selectedFrequency[vfo] == frequency {
			return
		}
		c.selectedFrequency[vfo] = frequency

		c.view.SetFrequency(vfo, frequency)
	})
}

func (c *Controller) VFOTuned(vfo model.VFOID, frequency model.Frequency) {
	c.asyncRunner(func() {
		if vfo == model.VFO1 && c.editing {
			return
		}
		// the operator left the frequency of the QSO in progress, therefore the input
		// belongs to the previous frequency and gets cleared
		if !model.IsFrequencyJump(c.selectedFrequency[vfo], frequency) {
			return
		}
		c.selectedFrequency[vfo] = frequency
		c.clearInput(vfo)
	})
}

// clearInput resets the input for the given VFO without changing the focused VFO.
func (c *Controller) clearInput(vfo model.VFOID) {
	c.claims.Release(vfo)
	c.vfos[vfo].Refresh()
	c.activeField[vfo] = model.CallsignField

	lastExchange := c.logbook.LastExchange()
	c.fillExchangeDefaults(vfo, lastExchange)

	c.refreshMyNumberInputs()
	c.showInput()
	c.view.SetFrequency(vfo, c.selectedFrequency[vfo])
	c.view.SetDuplicateMarker(vfo, false)
	c.view.ClearMessage(vfo)
	c.callinfoInputChanged(vfo, "", model.NoBand, model.NoMode, []string{})

	// Only push UI focus if this is focused VFO. For non-focused VFOs
	// we reset internal state only — SetActiveField would steal focus.
	if vfo == c.focusedVFO {
		c.view.SetActiveField(vfo, c.activeField[vfo])
	}
}

// callinfoInputChanged notifies the callinfo subsystem about input changes on a specific VFO.
func (c *Controller) callinfoInputChanged(vfo model.VFOID, call string, band model.Band, mode model.Mode, exchange []string) {
	if c.callinfo == nil {
		return
	}
	c.callinfo.InputChanged(vfo, call, band, mode, exchange)
}

func (c *Controller) bandSelected(s string) {
	if band, err := model.ParseBand(s); err == nil {
		// log.Printf("Band selected: %v", band)
		c.selectedBand[c.focusedVFO] = band
		c.vfos[c.focusedVFO].SetBand(band)
		c.enterCallsign(c.input[c.focusedVFO].callsign)
	}
}

func (c *Controller) VFOBandChanged(vfo model.VFOID, band model.Band) {
	c.asyncRunner(func() {
		if vfo == model.VFO1 && c.editing {
			return
		}
		if band == model.NoBand || band == c.selectedBand[vfo] {
			return
		}
		c.selectedBand[vfo] = band
		c.input[vfo].band = band.String()

		c.view.SetBand(vfo, c.input[vfo].band)
	})
}

func (c *Controller) modeSelected(s string) {
	if mode, err := model.ParseMode(s); err == nil {
		// log.Printf("Mode selected: %v", mode)
		c.selectedMode[c.focusedVFO] = mode
		c.vfos[c.focusedVFO].SetMode(mode)
		if c.generateReport {
			c.generateReportForMode(mode)
		}
		c.enterCallsign(c.input[c.focusedVFO].callsign)
	}
}

func (c *Controller) generateReportForMode(mode model.Mode) {
	generatedReport := defaultReportForMode(mode)
	myIndex := c.myReportExchangeField.Field.ExchangeIndex()
	if myIndex > 0 {
		c.myReport = generatedReport
		c.myExchange[myIndex-1] = generatedReport
		c.view.SetMyExchange(myIndex, c.myReport)
	}
	theirIndex := c.theirReportExchangeField.Field.ExchangeIndex()
	if theirIndex > 0 {
		c.input[c.focusedVFO].theirReport = generatedReport
		c.input[c.focusedVFO].theirExchange[theirIndex-1] = generatedReport
		c.view.SetTheirExchange(c.focusedVFO, theirIndex, c.myReport)
	}
}

func defaultReportForMode(mode model.Mode) string {
	switch mode {
	case model.ModeCW, model.ModeDigital, model.ModeRTTY:
		return "599"
	case model.ModeSSB, model.ModeFM:
		return "59"
	default:
		return ""
	}
}

func (c *Controller) VFOModeChanged(vfo model.VFOID, mode model.Mode) {
	c.asyncRunner(func() {
		if vfo == model.VFO1 && c.editing {
			return
		}
		if mode == model.NoMode || mode == c.selectedMode[vfo] {
			return
		}
		c.selectedMode[vfo] = mode
		c.input[vfo].mode = mode.String()

		c.view.SetMode(vfo, c.input[vfo].mode)
	})
}

func (c *Controller) VFOPTTChanged(vfo model.VFOID, active bool) {
	c.asyncRunner(func() {
		if vfo != model.VFO1 {
			c.view.SetTXState(vfo, active, false, 0)
			return
		}
		c.ptt = active
		c.updateTXState()
	})
}

func (c *Controller) TXVFOChanged(vfo model.VFOID) {
	c.asyncRunner(func() {
		c.txVFO = vfo
		c.view.SetTXVFO(vfo)
	})
}

func (c *Controller) ParrotActive(active bool) {
	c.parrotActive = active
	c.updateTXState()
	if active {
		c.clearInput(model.VFO1)
	}
}

func (c *Controller) ParrotTimeLeft(timeLeft time.Duration) {
	c.parrotTimeLeft = timeLeft
	c.updateTXState()
}

func (c *Controller) updateTXState() {
	c.view.SetTXState(model.VFO1, c.ptt, c.parrotActive, c.parrotTimeLeft)
}

func (c *Controller) SendQuestion() {
	if c.keyer == nil {
		return
	}
	if !c.canTransmit() {
		return
	}

	switch {
	case c.activeField[c.focusedVFO].IsTheirExchange():
		c.keyer.SendQuestion("nr")
	default:
		callsign := c.input[c.focusedVFO].callsign
		if callsign != "" {
			c.keyer.SendText(callsign)
		} else {
			c.keyer.SendQuestion("")
		}
	}
}

func (c *Controller) RepeatLastTransmission() {
	if c.keyer == nil {
		return
	}
	if !c.canTransmit() {
		return
	}

	c.keyer.Repeat()
}

func (c *Controller) enterCallsign(s string) {
	c.emitCallsignEntered(c.input[c.focusedVFO].callsign)
	c.callinfoInputChanged(c.focusedVFO, c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], c.input[c.focusedVFO].theirExchange)

	callsign, err := model.ParseCallsign(s)
	if err != nil {
		return
	}

	// Sticky per-keystroke serial claim. Skip while editing — editQSO's serial owns the slot.
	if !c.editing {
		c.claimSerialFor(c.focusedVFO)
	}

	qso, found := c.isDuplicate(c.focusedVFO, callsign)
	if !found {
		c.view.ClearMessage(c.focusedVFO)
		return
	}

	c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, qso.MyNumber.String()), model.CallsignField)
}

func (c *Controller) enterTheirExchange(field model.EntryField) {
	if c.callinfo == nil {
		return
	}
	c.callinfoInputChanged(c.focusedVFO, c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], c.input[c.focusedVFO].theirExchange)
	c.clearErrorOnField(field)
}

func (c *Controller) QSOSelected(qso model.QSO) {
	if c.ignoreQSOSelection {
		return
	}

	log.Printf("QSO selected: %v", qso)
	c.enterEditMode(qso)

	c.showQSO(qso)
	c.view.SetActiveField(c.focusedVFO, model.CallsignField)
	c.view.SetEditingMarker(model.VFO1, true)
	c.callinfoInputChanged(c.focusedVFO, qso.Callsign.String(), qso.Band, qso.Mode, qso.TheirExchange)
}

func (c *Controller) EnterPressed() {
	if c.parseCallsignCommand() {
		c.input[c.focusedVFO].callsign = ""
		c.enterCallsign(c.input[c.focusedVFO].callsign)
		c.view.SetCallsign(c.focusedVFO, c.input[c.focusedVFO].callsign)
		return
	}

	if c.esmEnabled && !c.editing {
		c.NextESMStep()
	} else {
		c.Log()
	}
}

func (c *Controller) CurrentQSOState() (model.Callsign, model.QSODataState) {
	callEmpty := (c.input[c.focusedVFO].callsign == "")

	call, err := model.ParseCallsign(c.input[c.focusedVFO].callsign)
	callOK := (err == nil)

	theirExchange := make([]string, len(c.theirExchangeFields))
	_, err = c.parseTheirExchange(theirExchange, nil, nil)
	exchangeOK := (err == nil)

	switch {
	case callEmpty, !callOK:
		return model.NoCallsign, model.QSODataEmpty
	case callOK && !exchangeOK:
		return call, model.QSODataInvalid
	case callOK && exchangeOK:
		return call, model.QSODataValid
	default:
		log.Printf("invalid QSO state: %s, %+v", c.input[c.focusedVFO].callsign, c.input[c.focusedVFO].theirExchange)
		return model.NoCallsign, model.QSODataEmpty
	}
}

func (c *Controller) Log() {
	var err error
	qso := model.QSO{}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Callsign, err = model.ParseCallsign(c.input[c.focusedVFO].callsign)
	if err != nil {
		c.showErrorOnField(err, model.CallsignField)
		return
	}

	qso.Frequency = c.selectedFrequency[c.focusedVFO]

	qso.Band, err = model.ParseBand(c.input[c.focusedVFO].band)
	if err != nil {
		c.showErrorOnField(err, model.BandField)
		return
	}

	qso.Mode, err = model.ParseMode(c.input[c.focusedVFO].mode)
	if err != nil {
		c.showErrorOnField(err, model.ModeField)
		return
	}

	// handle their exchange
	qso.TheirExchange = make([]string, len(c.theirExchangeFields))
	fieldWithError, err := c.parseTheirExchange(qso.TheirExchange, &qso.TheirReport, &qso.TheirNumber)
	if err != nil {
		c.showErrorOnField(err, fieldWithError.Field)
		return
	}

	// handle my exchange
	myNumber, err := strconv.Atoi(c.myNumber)
	if err == nil {
		qso.MyNumber = model.QSONumber(myNumber)
	}
	qso.MyExchange = make([]string, len(c.myExchangeFields))
	for i, field := range c.myExchangeFields {
		value := c.myExchange[i]
		qso.MyExchange[i] = value

		// TODO parse the value using the conval validators and show an error on the field

		switch field.Field {
		case c.myReportExchangeField.Field:
			qso.MyReport, err = model.ParseRST(value)
			if err != nil {
				c.showErrorOnField(err, field.Field)
				return
			}
		case c.myNumberExchangeField.Field:
			myNumber, err := strconv.Atoi(value)
			if err == nil {
				qso.MyExchange[i] = fmt.Sprintf("%03d", myNumber)
				qso.MyNumber = model.QSONumber(myNumber)
			} else if len(field.Properties) == 1 {
				c.showErrorOnField(err, field.Field)
				return
			}
		}
	}
	if c.editing {
		qso.Workmode = c.editQSO.Workmode
		c.logbook.UpdateQSO(qso)
	} else {
		qso.Workmode = c.workmode
		c.logbook.AddQSO(qso)
		c.lastNumber[c.focusedVFO] = qso.MyNumber
	}

	// NextQSONumber may have advanced; refresh the other VFO's serial preview.
	c.refreshMyNumberInputs()

	c.emitCallsignLogged(qso.Callsign.String(), qso.Frequency)

	if c.workmode == model.SearchPounce {
		spot := model.Spot{
			Call:      qso.Callsign,
			Frequency: qso.Frequency,
			Band:      qso.Band,
			Mode:      qso.Mode,
			Time:      qso.Time,
			Source:    model.WorkedSpot,
		}
		c.bandmap.Add(spot)
	}

	c.Clear()
}

func (c *Controller) parseCallsignCommand() bool {
	if c.activeField[c.focusedVFO] != model.CallsignField {
		return false
	}

	if f, ok := parseKilohertz(c.input[c.focusedVFO].callsign); ok {
		c.frequencyEntered(f)
		return true
	}

	if b, err := model.ParseBand(c.input[c.focusedVFO].callsign); err == nil {
		c.bandEntered(b)
		return true
	}

	if isCQCommand(c.input[c.focusedVFO].callsign) {
		c.bandmap.GotoCQMarker()
		return true
	}

	if number, ok := parseMarkerCommand(c.input[c.focusedVFO].callsign); ok {
		c.bandmap.GotoNumberedMarker(number)
		return true
	}

	if text, ok := parseTextMarkerCommand(c.input[c.focusedVFO].callsign); ok {
		c.markFromText(text)
		return true
	}

	if call, ok := parseBandmapCallsign(c.input[c.focusedVFO].callsign); ok {
		c.bandmap.SelectByCallsign(call)
		return true
	}

	// a command without a target does nothing, it must not reach the ESM and start a
	// transmission
	return isCommand(c.input[c.focusedVFO].callsign)
}

func isCommand(s string) bool {
	text := strings.TrimSpace(s)
	return strings.HasPrefix(text, commandPrefix) || strings.HasPrefix(text, textMarkerCommand)
}

func parseKilohertz(s string) (model.Frequency, bool) {
	kHz, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return model.Frequency(kHz * 1000), true
}

func isCQCommand(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), cqCommand)
}

func parseMarkerCommand(s string) (int, bool) {
	text, found := strings.CutPrefix(strings.TrimSpace(s), commandPrefix)
	if !found {
		return 0, false
	}
	return parseMarkerNumber(text)
}

func parseTextMarkerCommand(s string) (string, bool) {
	text, found := strings.CutPrefix(strings.TrimSpace(s), textMarkerCommand)
	if !found {
		return "", false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	return text, true
}

func parseBandmapCallsign(s string) (model.Callsign, bool) {
	text, found := strings.CutPrefix(strings.TrimSpace(s), commandPrefix)
	if !found {
		return model.Callsign{}, false
	}

	call, err := model.ParseCallsign(text)
	if err != nil {
		log.Printf("invalid bandmap callsign: %v", err)
		return model.Callsign{}, false
	}
	return call, true
}

// parseTheirExchange parses their exchange fields and writes the values into the slice/pointers. Parsing errors are returned.
// The arguments may be nil, this can be used to just validate the input.
func (c *Controller) parseTheirExchange(theirExchange []string, theirReport *model.RST, theirNumber *model.QSONumber) (model.ExchangeField, error) {
	for i, field := range c.theirExchangeFields {
		value := c.input[c.focusedVFO].theirExchange[i]
		if value == "" && !field.EmptyAllowed {
			return field, fmt.Errorf("%s is missing", field.Short) // TODO use field.Name
		}

		// TODO parse the value using the conval validators and show an error on the field

		if len(theirExchange) > i {
			theirExchange[i] = value
		}

		switch field.Field {
		case c.theirReportExchangeField.Field:
			rst, err := model.ParseRST(value)
			if err != nil {
				return field, err
			}
			if theirReport != nil {
				*theirReport = rst
			}
		case c.theirNumberExchangeField.Field:
			n, err := strconv.Atoi(value)
			onlySerial := true
			for _, p := range field.Properties {
				if p != conval.SerialNumberProperty && p != conval.EmptyProperty {
					onlySerial = false
					break
				}
			}
			if err == nil {
				if onlySerial && len(theirExchange) > i {
					theirExchange[i] = fmt.Sprintf("%03d", n)
				}
				if theirNumber != nil {
					*theirNumber = model.QSONumber(n)
				}
			} else if onlySerial {
				return field, err
			}
		default:
			if len(c.currentCallinfoFrame[c.focusedVFO].PredictedExchange) == len(theirExchange) && len(theirExchange) > i && theirExchange[i] == "" {
				c.setTheirExchangePrediction(i, c.currentCallinfoFrame[c.focusedVFO].PredictedExchange[i])
				return field, fmt.Errorf("check their exchange")
			}
		}
	}
	return model.ExchangeField{}, nil
}

func (c *Controller) showErrorOnField(err error, field model.EntryField) {
	c.SetActiveField(field)
	c.errorField[c.focusedVFO] = field
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
	c.view.ShowMessage(c.focusedVFO, err)
}

func (c *Controller) clearErrorOnField(field model.EntryField) {
	if c.errorField[c.focusedVFO] != field {
		return
	}
	c.view.ClearMessage(c.focusedVFO)
}

func (c *Controller) Clear() {
	// If we are in edit mode, exiting the modal is the whole job of Clear.
	// leaveEditMode restores the pre-edit state across all the per-VFO fields.
	if c.editing {
		c.leaveEditMode()
		c.showInput()
		c.view.SetFrequency(c.focusedVFO, c.selectedFrequency[c.focusedVFO])
		c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
		c.view.SetDuplicateMarker(c.focusedVFO, false)
		c.view.SetEditingMarker(model.VFO1, false)
		c.view.ClearMessage(c.focusedVFO)
		c.selectLastQSO()
		c.callinfoInputChanged(c.focusedVFO, "", model.NoBand, model.NoMode, []string{})
		return
	}

	// Release before the wipe so claim slots reflect "nothing pending" while we rebuild.
	// The companion refresh writes the serial display below, after the input zero/fill pass.
	c.claims.Release(c.focusedVFO)

	c.vfos[c.focusedVFO].Refresh()

	c.activeField[c.focusedVFO] = model.CallsignField

	lastExchange := c.logbook.LastExchange()
	c.fillExchangeDefaults(c.focusedVFO, lastExchange)
	// Non-focused VFOs that have no in-progress QSO are also initialized so their default
	// report follows the current contest/mode.
	for vfo := range model.VFOCount {
		v := model.VFOID(vfo)
		if v == c.focusedVFO {
			continue
		}
		if c.claims.claimed[v] != 0 || c.input[v].callsign != "" {
			continue
		}
		c.fillExchangeDefaults(v, lastExchange)
	}

	// Refresh serial displays for both VFOs (other VFO's preview may shift after the release).
	c.refreshMyNumberInputs()

	c.updateESM()

	c.showInput()
	c.view.SetFrequency(c.focusedVFO, c.selectedFrequency[c.focusedVFO])
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
	c.view.SetDuplicateMarker(c.focusedVFO, false)
	c.view.SetEditingMarker(model.VFO1, false)
	c.view.ClearMessage(c.focusedVFO)
	c.selectLastQSO()
	c.callinfoInputChanged(c.focusedVFO, "", model.NoBand, model.NoMode, []string{})
}

// fillExchangeDefaults resets vfo's input fields (callsign/band/mode/exchange) and seeds
// the default exchange values (mode-derived report, last-exchange carry-over). Idempotent.
func (c *Controller) fillExchangeDefaults(vfo model.VFOID, lastExchange []string) {
	c.input[vfo].callsign = ""
	if c.selectedBand[vfo] != model.NoBand {
		c.input[vfo].band = c.selectedBand[vfo].String()
	}
	generatedReport := ""
	if c.selectedMode[vfo] != model.NoMode {
		c.input[vfo].mode = c.selectedMode[vfo].String()
		generatedReport = defaultReportForMode(c.selectedMode[vfo])
	}

	c.myReport = ""
	c.myNumber = ""
	c.input[vfo].theirReport = ""
	c.input[vfo].theirNumber = ""
	c.input[vfo].theirExchange = make([]string, len(c.theirExchangeFields))
	c.myExchange = make([]string, len(c.myExchangeFields))
	for i, value := range c.defaultExchangeValues {
		if value == "" && i < len(lastExchange) {
			value = lastExchange[i]
		}
		if i >= len(c.myExchangeFields) {
			continue
		}
		c.myExchange[i] = value
		if i == c.myReportExchangeField.Field.ExchangeIndex()-1 {
			if c.generateReport {
				value = generatedReport
			}
			c.myReport = value
			c.myExchange[i] = value
			c.input[vfo].theirExchange[i] = value
			c.input[vfo].theirReport = value
		}
	}
}

func (c *Controller) Activate() {
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
}

func (c *Controller) EditLastQSO() {
	c.activeField[c.focusedVFO] = model.CallsignField
	c.qsoList.SelectLastQSO()
}

func (c *Controller) StopTX() {
	c.keyer.Stop()
}

func (c *Controller) selectLastQSO() {
	c.ignoreQSOSelection = true
	c.qsoList.SelectLastQSO()
	c.ignoreQSOSelection = false
}

func (c *Controller) CurrentValues() model.KeyerValues {
	myNumber, _ := strconv.Atoi(c.myNumber)

	myXchanges := make([]string, 0, len(c.myExchange))
	for i, field := range c.myExchangeFields {
		switch field.Field {
		case c.myReportExchangeField.Field, c.myNumberExchangeField.Field:
			continue
		default:
			myXchanges = append(myXchanges, c.myExchange[i])
		}
	}

	values := model.KeyerValues{}
	values.MyReport, _ = model.ParseRST(c.myReport)
	values.MyNumber = model.QSONumber(myNumber)
	values.MyXchange = strings.Join(myXchanges, " ")
	values.MyExchange = strings.Join(c.myExchange, " ")
	values.MyExchanges = c.myExchange
	values.TheirCall = c.input[c.focusedVFO].callsign
	values.LastNumber = c.lastNumber[c.focusedVFO]

	return values
}

func (c *Controller) CurrentVFOState() (model.Frequency, model.Band, model.Mode) {
	return c.selectedFrequency[c.focusedVFO], c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO]
}

func (c *Controller) FocusedVFO() (string, bool) {
	return c.vfos[c.focusedVFO].Name(), c.vfo2Enabled
}

func (c *Controller) StationChanged(station model.Station) {
	c.stationCallsign = station.Callsign.String()
}

func (c *Controller) ContestChanged(contest model.Contest) {
	c.updateExchangeFields(contest)
}

func (c *Controller) WorkmodeChanged(vfo model.VFOID, workmode model.Workmode) {
	c.vfoWorkmode[vfo] = workmode
	if vfo == c.focusedVFO {
		c.workmode = workmode
		c.updateESM()
	}
	c.view.SetVFOWorkmode(vfo, workmode)
}

func (c *Controller) updateExchangeFields(contest model.Contest) {
	c.myExchangeFields = contest.MyExchangeFields
	c.myReportExchangeField = contest.MyReportExchangeField
	c.myNumberExchangeField = contest.MyNumberExchangeField
	c.theirExchangeFields = contest.TheirExchangeFields
	c.theirReportExchangeField = contest.TheirReportExchangeField
	c.theirNumberExchangeField = contest.TheirNumberExchangeField
	c.generateSerialExchange = contest.GenerateSerialExchange
	c.generateReport = contest.GenerateReport
	c.defaultExchangeValues = contest.ExchangeValues

	c.myExchange = make([]string, len(contest.MyExchangeFields))
	for vfo := range model.VFOCount {
		c.input[vfo].theirExchange = make([]string, len(contest.TheirExchangeFields))
	}

	c.Clear()
}

func (c *Controller) MarkInBandmap() {
	callsign := strings.TrimSpace(c.input[c.focusedVFO].callsign)
	frequency := c.selectedFrequency[c.focusedVFO]
	band := c.selectedBand[c.focusedVFO]

	// an empty callsign marks the frequency with the next available number
	if callsign == "" {
		c.bandmap.MarkWithNextNumber(frequency, band, c.currentMarkerInput())
		return
	}

	// a number marks the frequency with exactly that number
	if number, ok := parseMarkerNumber(callsign); ok {
		// the callsign field holds the number of the marker, therefore the marker takes
		// only the exchange
		input := c.currentMarkerInput()
		input.Call = ""
		c.bandmap.MarkWithNumber(number, frequency, band, input)
		c.clearCallsignInput()
		return
	}

	if text, ok := parseTextMarkerCommand(callsign); ok {
		c.markFromText(text)
		return
	}

	call, err := model.ParseCallsign(callsign)
	if err != nil {
		// an incomplete callsign is no spot for the other operators, but the operator
		// parks the QSO at a marker and continues it later
		c.bandmap.MarkWithNextNumber(frequency, band, c.currentMarkerInput())
		return
	}
	spot := model.Spot{
		Call:      call,
		Frequency: c.selectedFrequency[c.focusedVFO],
		Band:      c.selectedBand[c.focusedVFO],
		Mode:      c.selectedMode[c.focusedVFO],
		Time:      c.clock.Now(),
		Source:    model.ManualSpot,
	}
	c.bandmap.Add(spot)
}

func (c *Controller) MarkWithNumberInBandmap(number int) {
	c.bandmap.MarkWithNumber(number, c.selectedFrequency[c.focusedVFO], c.selectedBand[c.focusedVFO], c.currentMarkerInput())
}

func (c *Controller) currentMarkerInput() model.MarkerInput {
	// the operator comes back to the marker later and continues the QSO there
	return model.MarkerInput{
		Call:          strings.TrimSpace(c.input[c.focusedVFO].callsign),
		TheirExchange: slices.Clone(c.input[c.focusedVFO].theirExchange),
	}
}

func parseMarkerNumber(s string) (int, bool) {
	number, err := strconv.Atoi(s)
	if err != nil || number < 1 {
		return 0, false
	}
	return number, true
}

func (c *Controller) markFromText(text string) {
	frequency := c.selectedFrequency[c.focusedVFO]
	band := c.selectedBand[c.focusedVFO]

	// a text that is a number marks the frequency with that number
	if number, ok := parseMarkerNumber(text); ok {
		c.bandmap.MarkWithNumber(number, frequency, band, c.currentMarkerInput())
	} else {
		c.bandmap.MarkWithText(text, frequency, band)
	}
	c.clearCallsignInput()
}

func (c *Controller) clearCallsignInput() {
	c.input[c.focusedVFO].callsign = ""
	c.enterCallsign(c.input[c.focusedVFO].callsign)
	c.view.SetCallsign(c.focusedVFO, c.input[c.focusedVFO].callsign)
}

func (c *Controller) MarkerSelected(vfo model.VFOID, marker model.BandmapMarker) {
	// the bandmap decides which VFO goes to the marker, the focus follows that decision
	c.SetFocusedVFO(vfo)

	c.Clear()
	// setting the frequency also moves the rig to the band of the marker. A marker names
	// no mode, therefore the mode of the VFO stays as it is.
	c.frequencyEntered(marker.Frequency)
	c.SetActiveField(model.CallsignField)
	c.restoreMarkerInput(marker.Input)
	c.Activate()
}

func (c *Controller) restoreMarkerInput(input model.MarkerInput) {
	if input.IsEmpty() {
		return
	}

	// the operator continues the QSO that they left behind on the frequency of the marker

	c.input[c.focusedVFO].callsign = input.Call
	c.input[c.focusedVFO].theirExchange = ensureLen(input.TheirExchange, len(c.theirExchangeFields))
	c.showInput()
	c.enterCallsign(input.Call)
	c.updateESM()
}

func (c *Controller) EntrySelected(vfo model.VFOID, entry model.BandmapEntry) {
	// the bandmap decides which VFO works the spot, the focus follows that decision
	c.SetFocusedVFO(vfo)

	c.Clear()
	// setting the frequency also moves the rig to the band of the spot
	c.frequencyEntered(entry.Frequency)
	if entry.Mode != model.NoMode && entry.Mode != c.selectedMode[c.focusedVFO] {
		c.vfos[c.focusedVFO].SetMode(entry.Mode)
	}
	c.SetActiveField(model.CallsignField)
	c.Enter(entry.Call.String())
	c.view.SetCallsign(c.focusedVFO, c.input[c.focusedVFO].callsign)
	// the spot list may live in another window, therefore the callsign field of the
	// target VFO explicitly takes the keyboard focus
	c.Activate()
}
