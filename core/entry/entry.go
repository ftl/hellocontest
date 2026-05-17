package entry

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/conval"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
	"github.com/ftl/hellocontest/core/ticker"
)

const (
	jumpThreshold core.Frequency = 250 // Hz
)

// View represents the visual part of the QSO data entry.
type View interface {
	SetUTC(string)
	SetMyCall(string)
	SetMyExchange(int, string)

	SetFrequency(core.VFOID, core.Frequency)
	SetBand(vfo core.VFOID, text string)
	SetMode(vfo core.VFOID, text string)
	SetXITActive(vfo core.VFOID, active bool)
	SetXIT(vfo core.VFOID, active bool, offset core.Frequency)
	SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration)

	SetCallsign(string)
	SetTheirExchange(int, string)

	SetActiveField(core.EntryField)
	SelectText(core.EntryField, string)
	SetDuplicateMarker(bool)
	SetEditingMarker(bool)
	ShowMessage(...any)
	ClearMessage()
}

type input struct {
	callsign      string
	theirReport   string
	theirNumber   string
	theirExchange []string
	myReport      string
	myNumber      string
	myExchange    []string
	band          string
	mode          string
}

// Logbook functionality used for QSO entry.
type Logbook interface {
	NextQSONumber() core.QSONumber
	LastBand() core.Band
	LastMode() core.Mode
	LastExchange() []string
	AddQSO(core.QSO)
	UpdateQSO(core.QSO)
	FindDuplicateQSOs(core.Callsign, core.Band, core.Mode) []core.QSO
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
	GetText(workmode core.Workmode, index int) (string, error)
	SendText(text string, args ...any)
	Repeat()
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	InputChanged(call string, band core.Band, mode core.Mode, exchange []string)
}

type Bandmap interface {
	Add(core.Spot)
	SelectByCallsign(core.Callsign)
}

// NewController returns a new entry controller.
func NewController(settings core.Settings, clock core.Clock, logbook Logbook, qsoList QSOList, bandmap Bandmap, asyncRunner core.AsyncRunner) *Controller {
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

		vfos:                 make([]core.VFO, core.VFOCount),
		input:                make([]input, core.VFOCount),
		selectedFrequency:    make([]core.Frequency, core.VFOCount),
		selectedBand:         make([]core.Band, core.VFOCount),
		selectedMode:         make([]core.Mode, core.VFOCount),
		activeField:          make([]core.EntryField, core.VFOCount),
		errorField:           make([]core.EntryField, core.VFOCount),
		currentCallinfoFrame: make([]core.CallinfoFrame, core.VFOCount),
		claimedSerial:        make([]core.QSONumber, core.VFOCount),
		claimSnapshot:        make([]core.QSONumber, core.VFOCount),
		esmState:             make([]core.ESMState, core.VFOCount),
		esmMessage:           make([]string, core.VFOCount),

		stationCallsign: settings.Station().Callsign.String(),
	}
	for vfo := range len(result.vfos) {
		result.vfos[vfo] = new(nullVFO)
	}
	result.refreshTicker = ticker.New(clock, result.refreshUTC)
	result.updateExchangeFields(settings.Contest())
	return result
}

// VFOSwitcher is implemented by something that can command the rig to make a given VFO the current one.
type VFOSwitcher interface {
	SetCurrentVFO(core.VFOID)
}

type nullVFOSwitcher struct{}

func (n *nullVFOSwitcher) SetCurrentVFO(core.VFOID) {}

type editSnapshot struct {
	focusedVFO    core.VFOID
	input         input
	claimedSerial core.QSONumber
	claimSnapshot core.QSONumber
	activeField   core.EntryField
	errorField    core.EntryField
	callinfoFrame core.CallinfoFrame
	esmState      []core.ESMState
	esmMessage    []string
}

type Controller struct {
	clock       core.Clock
	view        View
	logbook     Logbook
	qsoList     QSOList
	keyer       Keyer
	callinfo    Callinfo
	vfos        []core.VFO
	vfoSwitcher VFOSwitcher
	bandmap     Bandmap
	esmView     ESMView

	asyncRunner   core.AsyncRunner
	refreshTicker *ticker.Ticker
	listeners     []any

	stationCallsign string
	workmode        core.Workmode

	myExchangeFields         []core.ExchangeField
	theirExchangeFields      []core.ExchangeField
	myReportExchangeField    core.ExchangeField
	myNumberExchangeField    core.ExchangeField
	theirReportExchangeField core.ExchangeField
	theirNumberExchangeField core.ExchangeField
	generateSerialExchange   bool
	generateReport           bool
	defaultExchangeValues    []string
	currentCallinfoFrame     []core.CallinfoFrame

	input               []input
	focusedVFO          core.VFOID
	vfo2Enabled         bool
	activeField         []core.EntryField
	errorField          []core.EntryField
	selectedFrequency   []core.Frequency
	selectedBand        []core.Band
	selectedMode        []core.Mode
	claimedSerial       []core.QSONumber
	claimSnapshot       []core.QSONumber
	editing             bool
	editQSO             core.QSO
	editSnapshot        *editSnapshot
	ignoreQSOSelection  bool
	ignoreFrequencyJump bool

	ptt            bool
	parrotActive   bool
	parrotTimeLeft time.Duration

	esmEnabled bool
	esmState   []core.ESMState
	esmMessage []string
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitCallsignEntered(callsign string) {
	for _, l := range c.listeners {
		if listener, ok := l.(core.CallsignEnteredListener); ok {
			listener.CallsignEntered(callsign)
		}
	}
}

func (c *Controller) emitCallsignLogged(callsign string, frequency core.Frequency) {
	for _, l := range c.listeners {
		if listener, ok := l.(core.CallsignLoggedListener); ok {
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
	c.Clear()
	c.refreshUTC()
}

func (c *Controller) LogbookLoaded() {
	c.selectedBand[core.VFO1] = c.logbook.LastBand()
	c.selectedMode[core.VFO1] = c.logbook.LastMode()
	c.Clear()
	c.showInput()
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) notifyCallinfoInputChanged(call string, band core.Band, mode core.Mode, exchange []string) {
	if c.callinfo == nil {
		return
	}
	c.callinfo.InputChanged(call, band, mode, exchange)
}

func (c *Controller) CallinfoFrameChanged(frame core.CallinfoFrame) {
	c.currentCallinfoFrame[c.focusedVFO] = frame
	// TODO what do we need to update here?
}

func (c *Controller) SetVFO(id core.VFOID, vfo core.VFO) {
	if vfo == nil {
		c.vfos[id] = new(nullVFO)
	} else {
		c.vfos[id] = vfo
	}
	vfo.Notify(c)
}

func (c *Controller) GotoNextField() core.EntryField {
	switch c.activeField[c.focusedVFO] {
	case core.CallsignField:
		c.leaveCallsignField()
	}

	transitions := map[core.EntryField]core.EntryField{
		core.BandField: core.CallsignField,
		core.ModeField: core.CallsignField,
	}
	if len(c.theirExchangeFields) > 0 {
		transitions[core.CallsignField] = core.TheirExchangeField(1)
	}
	for _, field := range c.myExchangeFields {
		transitions[field.Field] = core.CallsignField
	}
	for i, field := range c.theirExchangeFields {
		if i == len(c.theirExchangeFields)-1 {
			transitions[field.Field] = core.CallsignField
		} else {
			transitions[field.Field] = field.Field.NextExchangeField()
		}
	}

	nextField := transitions[c.activeField[c.focusedVFO]]
	if nextField == "" {
		nextField = core.CallsignField
	}

	c.SetActiveField(nextField)
	c.view.SetActiveField(c.activeField[c.focusedVFO])
	return c.activeField[c.focusedVFO]
}

func (c *Controller) GotoNextPlaceholder() {
	c.SetActiveField(core.CallsignField)
	c.view.SetActiveField(c.activeField[c.focusedVFO])
	c.view.SelectText(c.activeField[c.focusedVFO], core.FilterPlaceholder)
}

func (c *Controller) leaveCallsignField() {
	callsign, err := core.ParseCallsign(c.input[c.focusedVFO].callsign)
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
		c.view.SetDuplicateMarker(false)
		return
	}
	if c.editing {
		c.view.SetDuplicateMarker(c.editQSO.Callsign != callsign)
		return
	}

	c.view.SetDuplicateMarker(true)
}

// isPredictable returns true if the exchange for the given field is predictable.
func (c *Controller) isPredictable(field core.EntryField) bool {
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
	c.notifyCallinfoInputChanged(c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], []string{})

	if len(c.input[c.focusedVFO].theirExchange) == len(c.currentCallinfoFrame[c.focusedVFO].PredictedExchange) {
		for i, field := range c.theirExchangeFields {
			if !c.isPredictable(field.Field) {
				continue
			}
			c.setTheirExchangePrediction(i, c.currentCallinfoFrame[c.focusedVFO].PredictedExchange[i])
		}
	}
}

func (c *Controller) StartAutoRefresh() {
	c.refreshTicker.Start()
}

func (c *Controller) refreshUTC() {
	c.asyncRunner(func() {
		utc := c.clock.Now().UTC()
		c.view.SetUTC(utc.Format("15:04"))
	})
}

func (c *Controller) RefreshView() {
	c.showInput()
}

func (c *Controller) showQSO(qso core.QSO) {
	c.input[core.VFO1].callsign = qso.Callsign.String()
	c.input[core.VFO1].theirReport = qso.TheirReport.String()
	c.input[core.VFO1].theirNumber = qso.TheirNumber.String()
	c.input[core.VFO1].theirExchange = ensureLen(qso.TheirExchange, len(c.theirExchangeFields))
	c.input[core.VFO1].myReport = qso.MyReport.String()
	c.input[core.VFO1].myNumber = qso.MyNumber.String()
	c.input[core.VFO1].myExchange = ensureLen(qso.MyExchange, len(c.myExchangeFields))
	c.input[core.VFO1].band = qso.Band.String()
	c.input[core.VFO1].mode = qso.Mode.String()

	c.selectedFrequency[core.VFO1] = qso.Frequency
	c.selectedBand[core.VFO1] = qso.Band
	c.selectedMode[core.VFO1] = qso.Mode

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
	c.view.SetCallsign(c.input[core.VFO1].callsign)
	for i, value := range c.input[core.VFO1].theirExchange {
		c.view.SetTheirExchange(i+1, value)
	}
	for i, value := range c.input[core.VFO1].myExchange {
		c.view.SetMyExchange(i+1, value)
	}
	c.view.SetFrequency(core.VFO1, c.selectedFrequency[core.VFO1])
	c.view.SetBand(core.VFO1, c.input[core.VFO1].band)
	c.view.SetMode(core.VFO1, c.input[core.VFO1].mode)
}

// setTheirExchangePrediction replaces the value of the given field with the given predicted value,
// if the given value is not empty.
func (c *Controller) setTheirExchangePrediction(i int, value string) {
	if value == "" {
		return
	}
	c.input[c.focusedVFO].theirExchange[i] = value
	c.view.SetTheirExchange(i+1, value)
}

func (c *Controller) isDuplicate(vfo core.VFOID, callsign core.Callsign) (core.QSO, bool) {
	qsos := c.logbook.FindDuplicateQSOs(callsign, c.selectedBand[vfo], c.selectedMode[vfo])
	if len(qsos) == 0 {
		return core.QSO{}, false
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

// SetFocusedVFO is the single funnel for changing focused VFO. It commands the
// rig to make `vfo` the current TX VFO via vfoSwitcher.
func (c *Controller) SetFocusedVFO(vfo core.VFOID) {
	if vfo == core.VFO2 && !c.vfo2Enabled {
		return
	}
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
	c.vfoSwitcher.SetCurrentVFO(vfo)
}

// setFocusedVFOSilent updates focusedVFO without commanding the rig. Used by edit mode.
func (c *Controller) setFocusedVFOSilent(vfo core.VFOID) {
	if c.focusedVFO == vfo {
		return
	}
	c.focusedVFO = vfo
}

// ToggleFocusedVFO flips between VFO1 and VFO2. No-op if VFO2 is disabled.
func (c *Controller) ToggleFocusedVFO() {
	if !c.vfo2Enabled {
		c.SetFocusedVFO(core.VFO1)
		return
	}
	if c.focusedVFO == core.VFO1 {
		c.SetFocusedVFO(core.VFO2)
	} else {
		c.SetFocusedVFO(core.VFO1)
	}
}

// FocusVFO1 sets the focused VFO to VFO1.
func (c *Controller) FocusVFO1() {
	c.SetFocusedVFO(core.VFO1)
}

// FocusVFO2 sets the focused VFO to VFO2. No-op if VFO2 is disabled.
func (c *Controller) FocusVFO2() {
	c.SetFocusedVFO(core.VFO2)
}

// LogVFO synthesises focus on vfo and then logs the QSO from that row.
func (c *Controller) LogVFO(vfo core.VFOID) {
	c.SetFocusedVFO(vfo)
	c.Log()
}

// ClearVFO synthesises focus on vfo and then clears that row.
func (c *Controller) ClearVFO(vfo core.VFOID) {
	c.SetFocusedVFO(vfo)
	c.Clear()
}

// RadioChanged toggles VFO2 availability based on the connected radio's single-VFO flag.
// Implements core.RadioChangedListener.
func (c *Controller) RadioChanged(_ string, singleVFO bool) {
	c.vfo2Enabled = !singleVFO
	if !c.vfo2Enabled && c.focusedVFO == core.VFO2 {
		c.releaseSerialClaimFor(core.VFO2)
		c.input[core.VFO2] = input{}
		c.setFocusedVFOSilent(core.VFO1)
	}
}

// canTransmit reports whether the keyer is currently allowed to transmit. False during edit mode.
func (c *Controller) canTransmit() bool {
	return !c.editing
}

// nextUnclaimedSerial walks forward from the logbook's NextQSONumber, skipping any serial
// currently claimed by the other VFO, and returns the first free serial for forVFO.
func (c *Controller) nextUnclaimedSerial(forVFO core.VFOID) core.QSONumber {
	candidate := c.logbook.NextQSONumber()
	otherVFO := core.VFO1
	if forVFO == core.VFO1 {
		otherVFO = core.VFO2
	}
	if c.claimedSerial[otherVFO] != 0 && c.claimedSerial[otherVFO] == candidate {
		candidate++
	}
	return candidate
}

// displayedSerialFor returns the serial currently visible to the operator on vfo's row:
// the claimed value if any, otherwise the next unclaimed preview.
func (c *Controller) displayedSerialFor(vfo core.VFOID) core.QSONumber {
	if c.claimedSerial[vfo] != 0 {
		return c.claimedSerial[vfo]
	}
	return c.nextUnclaimedSerial(vfo)
}

// claimSerialFor reserves the next unclaimed serial for the given VFO if it has none yet.
// Sticky: subsequent calls while a claim exists are no-ops.
func (c *Controller) claimSerialFor(vfo core.VFOID) {
	if c.claimedSerial[vfo] != 0 {
		return
	}
	c.claimedSerial[vfo] = c.nextUnclaimedSerial(vfo)
	c.claimSnapshot[vfo] = c.logbook.NextQSONumber()
	c.refreshMyNumberInputs()
}

// releaseSerialClaimFor frees the claim slot for vfo. The serial itself is reusable only if
// NextQSONumber has not advanced since the claim was taken.
func (c *Controller) releaseSerialClaimFor(vfo core.VFOID) {
	c.claimedSerial[vfo] = 0
	c.claimSnapshot[vfo] = 0
	c.refreshMyNumberInputs()
}

// refreshMyNumberInputs syncs each VFO's input.myNumber (and exchange serial slot) with the
// current displayed serial value. Reads NextQSONumber once.
func (c *Controller) refreshMyNumberInputs() {
	base := c.logbook.NextQSONumber()
	for vfo := range core.VFOCount {
		c.writeMyNumberInput(core.VFOID(vfo), c.displayedSerialWithBase(core.VFOID(vfo), base))
	}
}

func (c *Controller) refreshMyNumberInput(vfo core.VFOID) {
	c.writeMyNumberInput(vfo, c.displayedSerialFor(vfo))
}

func (c *Controller) displayedSerialWithBase(vfo core.VFOID, base core.QSONumber) core.QSONumber {
	if c.claimedSerial[vfo] != 0 {
		return c.claimedSerial[vfo]
	}
	other := core.VFO1
	if vfo == core.VFO1 {
		other = core.VFO2
	}
	if c.claimedSerial[other] != 0 && c.claimedSerial[other] == base {
		return base + 1
	}
	return base
}

func (c *Controller) writeMyNumberInput(vfo core.VFOID, serial core.QSONumber) {
	value := serial.String()
	c.input[vfo].myNumber = value
	i := c.myNumberExchangeField.Field.ExchangeIndex() - 1
	if i < 0 || !c.generateSerialExchange {
		return
	}
	if i >= len(c.input[vfo].myExchange) {
		return
	}
	c.input[vfo].myExchange[i] = value
}

// enterEditMode snapshots VFO1's state, force-focuses VFO1 silently, marks editing,
// and claims the QSO's existing serial for VFO1 for the duration of the edit.
func (c *Controller) enterEditMode(qso core.QSO) {
	c.editSnapshot = &editSnapshot{
		focusedVFO:    c.focusedVFO,
		input:         c.input[core.VFO1],
		claimedSerial: c.claimedSerial[core.VFO1],
		claimSnapshot: c.claimSnapshot[core.VFO1],
		activeField:   c.activeField[core.VFO1],
		errorField:    c.errorField[core.VFO1],
		callinfoFrame: c.currentCallinfoFrame[core.VFO1],
		esmState:      append([]core.ESMState(nil), c.esmState...),
		esmMessage:    append([]string(nil), c.esmMessage...),
	}
	c.setFocusedVFOSilent(core.VFO1)
	c.editing = true
	c.editQSO = qso
	c.claimedSerial[core.VFO1] = qso.MyNumber
	c.claimSnapshot[core.VFO1] = c.logbook.NextQSONumber()
	// TODO step 6: c.view.SetVFOEnabled(core.VFO2, false)
}

// leaveEditMode restores the pre-edit state captured by enterEditMode. No-op if not editing.
func (c *Controller) leaveEditMode() {
	if c.editSnapshot == nil {
		return
	}
	snap := c.editSnapshot
	c.input[core.VFO1] = snap.input
	c.claimedSerial[core.VFO1] = snap.claimedSerial
	c.claimSnapshot[core.VFO1] = snap.claimSnapshot
	c.activeField[core.VFO1] = snap.activeField
	c.errorField[core.VFO1] = snap.errorField
	c.currentCallinfoFrame[core.VFO1] = snap.callinfoFrame
	copy(c.esmState, snap.esmState)
	copy(c.esmMessage, snap.esmMessage)
	c.editing = false
	c.editQSO = core.QSO{}
	c.editSnapshot = nil
	c.setFocusedVFOSilent(snap.focusedVFO)
	// TODO step 6: c.view.SetVFOEnabled(core.VFO2, true)
}

func (c *Controller) SetActiveField(field core.EntryField) {
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
	c.SetActiveField(core.CallsignField)
	c.Enter(callsign)
	c.view.SetCallsign(c.input[c.focusedVFO].callsign)
	c.view.SetActiveField(c.activeField[c.focusedVFO])
}

func (c *Controller) Enter(text string) {
	switch c.activeField[c.focusedVFO] {
	case core.CallsignField:
		c.input[c.focusedVFO].callsign = text
		c.enterCallsign(text)
	case core.BandField:
		c.input[c.focusedVFO].band = text
		c.bandSelected(text)
	case core.ModeField:
		c.input[c.focusedVFO].mode = text
		c.modeSelected(text)
	}

	i := c.activeField[c.focusedVFO].ExchangeIndex() - 1
	switch {
	case c.activeField[c.focusedVFO].IsMyExchange():
		c.input[c.focusedVFO].myExchange[i] = text
	case c.activeField[c.focusedVFO].IsTheirExchange():
		c.input[c.focusedVFO].theirExchange[i] = text
		c.enterTheirExchange(c.activeField[c.focusedVFO])
	}

	c.updateESM()
}

func (c *Controller) frequencyEntered(frequency core.Frequency) {
	// log.Printf("Frequency selected: %s", frequency)
	c.vfos[c.focusedVFO].SetFrequency(frequency)
}

func (c *Controller) bandEntered(band core.Band) {
	c.input[c.focusedVFO].band = band.String()
	c.vfos[c.focusedVFO].SetBand(band)
}

func (c *Controller) SetXITActive(active bool) {
	c.vfos[core.VFO1].SetXITActive(active)
	c.view.SetActiveField(c.activeField[c.focusedVFO])
}

func (c *Controller) VFOFrequencyChanged(vfo core.VFOID, frequency core.Frequency) {
	if vfo == core.VFO1 && c.editing {
		return
	}
	if c.selectedFrequency[vfo] == frequency {
		return
	}
	jump := math.Abs(float64(c.selectedFrequency[vfo]-frequency)) > float64(jumpThreshold)
	c.selectedFrequency[vfo] = frequency

	c.view.SetFrequency(vfo, frequency)

	if jump && !c.ignoreFrequencyJump {
		c.Clear()
		c.view.SetActiveField(c.activeField[c.focusedVFO])
	}
	c.ignoreFrequencyJump = false
}

func (c *Controller) bandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		// log.Printf("Band selected: %v", band)
		c.selectedBand[c.focusedVFO] = band
		c.vfos[c.focusedVFO].SetBand(band)
		c.enterCallsign(c.input[c.focusedVFO].callsign)
	}
}

func (c *Controller) VFOBandChanged(vfo core.VFOID, band core.Band) {
	if vfo == core.VFO1 && c.editing {
		return
	}
	if band == core.NoBand || band == c.selectedBand[c.focusedVFO] {
		return
	}
	c.selectedBand[c.focusedVFO] = band
	c.input[vfo].band = band.String()
	c.view.SetBand(vfo, c.input[vfo].band)
}

func (c *Controller) modeSelected(s string) {
	if mode, err := parse.Mode(s); err == nil {
		// log.Printf("Mode selected: %v", mode)
		c.selectedMode[c.focusedVFO] = mode
		c.vfos[c.focusedVFO].SetMode(mode)
		if c.generateReport {
			c.generateReportForMode(mode)
		}
		c.enterCallsign(c.input[c.focusedVFO].callsign)
	}
}

func (c *Controller) generateReportForMode(mode core.Mode) {
	generatedReport := defaultReportForMode(mode)
	myIndex := c.myReportExchangeField.Field.ExchangeIndex()
	if myIndex > 0 {
		c.input[c.focusedVFO].myReport = generatedReport
		c.input[c.focusedVFO].myExchange[myIndex-1] = generatedReport
		c.view.SetMyExchange(myIndex, c.input[c.focusedVFO].myReport)
	}
	theirIndex := c.theirReportExchangeField.Field.ExchangeIndex()
	if theirIndex > 0 {
		c.input[c.focusedVFO].theirReport = generatedReport
		c.input[c.focusedVFO].theirExchange[theirIndex-1] = generatedReport
		c.view.SetTheirExchange(theirIndex, c.input[c.focusedVFO].myReport)
	}
}

func defaultReportForMode(mode core.Mode) string {
	switch mode {
	case core.ModeCW, core.ModeDigital, core.ModeRTTY:
		return "599"
	case core.ModeSSB, core.ModeFM:
		return "59"
	default:
		return ""
	}
}

func (c *Controller) VFOModeChanged(vfo core.VFOID, mode core.Mode) {
	if vfo == core.VFO1 && c.editing {
		return
	}
	if mode == core.NoMode || mode == c.selectedMode[c.focusedVFO] {
		return
	}
	c.selectedMode[c.focusedVFO] = mode
	c.input[vfo].mode = mode.String()
	c.view.SetMode(vfo, c.input[vfo].mode)
}

func (c *Controller) VFOXITChanged(vfo core.VFOID, active bool, offset core.Frequency) {
	c.view.SetXIT(vfo, active, offset)
}

func (c *Controller) XITActiveChanged(active bool) {
	// TODO: add VFO parameter to XITActiveChanged
	c.view.SetXITActive(core.VFO1, active)
}

func (c *Controller) VFOPTTChanged(vfo core.VFOID, active bool) {
	if vfo != core.VFO1 {
		c.view.SetTXState(vfo, active, false, 0)
		return
	}
	c.ptt = active
	c.updateTXState()
}

func (c *Controller) ParrotActive(active bool) {
	c.parrotActive = active
	c.updateTXState()
	if active {
		c.Clear()
	}
}

func (c *Controller) ParrotTimeLeft(timeLeft time.Duration) {
	c.parrotTimeLeft = timeLeft
	c.updateTXState()
}

func (c *Controller) updateTXState() {
	c.view.SetTXState(core.VFO1, c.ptt, c.parrotActive, c.parrotTimeLeft)
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
		c.keyer.SendQuestion(c.input[c.focusedVFO].callsign)
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
	c.notifyCallinfoInputChanged(c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], c.input[c.focusedVFO].theirExchange)

	callsign, err := core.ParseCallsign(s)
	if err != nil {
		return
	}

	// Sticky per-keystroke serial claim. Skip while editing — editQSO's serial owns the slot.
	if !c.editing {
		c.claimSerialFor(c.focusedVFO)
	}

	qso, found := c.isDuplicate(c.focusedVFO, callsign)
	if !found {
		c.view.ClearMessage()
		return
	}

	c.showErrorOnField(fmt.Errorf("%s was worked before in QSO #%s", qso.Callsign, qso.MyNumber.String()), core.CallsignField)
}

func (c *Controller) enterTheirExchange(field core.EntryField) {
	if c.callinfo == nil {
		return
	}
	c.notifyCallinfoInputChanged(c.input[c.focusedVFO].callsign, c.selectedBand[c.focusedVFO], c.selectedMode[c.focusedVFO], c.input[c.focusedVFO].theirExchange)
	c.clearErrorOnField(field)
}

func (c *Controller) QSOSelected(qso core.QSO) {
	if c.ignoreQSOSelection {
		return
	}

	log.Printf("QSO selected: %v", qso)
	c.enterEditMode(qso)

	c.showQSO(qso)
	c.view.SetActiveField(core.CallsignField)
	c.view.SetEditingMarker(true)
	c.notifyCallinfoInputChanged(qso.Callsign.String(), qso.Band, qso.Mode, qso.TheirExchange)
}

func (c *Controller) EnterPressed() {
	if c.parseCallsignCommand() {
		c.input[c.focusedVFO].callsign = ""
		c.enterCallsign(c.input[c.focusedVFO].callsign)
		c.view.SetCallsign(c.input[c.focusedVFO].callsign)
		return
	}

	if c.esmEnabled && !c.editing {
		c.NextESMStep()
	} else {
		c.Log()
	}
}

func (c *Controller) CurrentQSOState() (core.Callsign, core.QSODataState) {
	callEmpty := (c.input[c.focusedVFO].callsign == "")

	call, err := core.ParseCallsign(c.input[c.focusedVFO].callsign)
	callOK := (err == nil)

	theirExchange := make([]string, len(c.theirExchangeFields))
	_, err = c.parseTheirExchange(theirExchange, nil, nil)
	exchangeOK := (err == nil)

	switch {
	case callEmpty, !callOK:
		return core.NoCallsign, core.QSODataEmpty
	case callOK && !exchangeOK:
		return call, core.QSODataInvalid
	case callOK && exchangeOK:
		return call, core.QSODataValid
	default:
		log.Printf("invalid QSO state: %s, %+v", c.input[c.focusedVFO].callsign, c.input[c.focusedVFO].theirExchange)
		return core.NoCallsign, core.QSODataEmpty
	}
}

func (c *Controller) Log() {
	var err error
	qso := core.QSO{}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Callsign, err = core.ParseCallsign(c.input[c.focusedVFO].callsign)
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}

	qso.Frequency = c.selectedFrequency[c.focusedVFO]

	qso.Band, err = parse.Band(c.input[c.focusedVFO].band)
	if err != nil {
		c.showErrorOnField(err, core.BandField)
		return
	}

	qso.Mode, err = parse.Mode(c.input[c.focusedVFO].mode)
	if err != nil {
		c.showErrorOnField(err, core.ModeField)
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
	myNumber, err := strconv.Atoi(c.input[c.focusedVFO].myNumber)
	if err == nil {
		qso.MyNumber = core.QSONumber(myNumber)
	}
	qso.MyExchange = make([]string, len(c.myExchangeFields))
	for i, field := range c.myExchangeFields {
		value := c.input[c.focusedVFO].myExchange[i]
		qso.MyExchange[i] = value

		// TODO parse the value using the conval validators and show an error on the field

		switch field.Field {
		case c.myReportExchangeField.Field:
			qso.MyReport, err = parse.RST(value)
			if err != nil {
				c.showErrorOnField(err, field.Field)
				return
			}
		case c.myNumberExchangeField.Field:
			myNumber, err := strconv.Atoi(value)
			if err == nil {
				qso.MyExchange[i] = fmt.Sprintf("%03d", myNumber)
				qso.MyNumber = core.QSONumber(myNumber)
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
	}

	// NextQSONumber may have advanced; refresh the other VFO's serial preview.
	c.refreshMyNumberInputs()

	c.emitCallsignLogged(qso.Callsign.String(), qso.Frequency)

	if c.workmode == core.SearchPounce {
		spot := core.Spot{
			Call:      qso.Callsign,
			Frequency: qso.Frequency,
			Band:      qso.Band,
			Mode:      qso.Mode,
			Time:      qso.Time,
			Source:    core.WorkedSpot,
		}
		c.bandmap.Add(spot)
	}

	c.Clear()
}

func (c *Controller) parseCallsignCommand() bool {
	if c.activeField[c.focusedVFO] != core.CallsignField {
		return false
	}

	if f, ok := parseKilohertz(c.input[c.focusedVFO].callsign); ok {
		c.frequencyEntered(f)
		return true
	}

	if b, err := parse.Band(c.input[c.focusedVFO].callsign); err == nil {
		c.bandEntered(b)
		return true
	}

	if call, ok := parseBandmapCallsign(c.input[c.focusedVFO].callsign); ok {
		c.bandmap.SelectByCallsign(call)
		return true
	}

	return false
}

func parseKilohertz(s string) (core.Frequency, bool) {
	kHz, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return core.Frequency(kHz * 1000), true
}

func parseBandmapCallsign(s string) (core.Callsign, bool) {
	if !strings.HasPrefix(s, "@") {
		return core.Callsign{}, false
	}

	call, err := core.ParseCallsign(s[1:])
	if err != nil {
		log.Printf("invalid bandmap callsign: %v", err)
		return core.Callsign{}, false
	}
	return call, true
}

// parseTheirExchange parses their exchange fields and writes the values into the slice/pointers. Parsing errors are returned.
// The arguments may be nil, this can be used to just validate the input.
func (c *Controller) parseTheirExchange(theirExchange []string, theirReport *core.RST, theirNumber *core.QSONumber) (core.ExchangeField, error) {
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
			rst, err := parse.RST(value)
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
					*theirNumber = core.QSONumber(n)
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
	return core.ExchangeField{}, nil
}

func (c *Controller) showErrorOnField(err error, field core.EntryField) {
	c.SetActiveField(field)
	c.errorField[c.focusedVFO] = field
	c.view.SetActiveField(c.activeField[c.focusedVFO])
	c.view.ShowMessage(err)
}

func (c *Controller) clearErrorOnField(field core.EntryField) {
	if c.errorField[c.focusedVFO] != field {
		return
	}
	c.view.ClearMessage()
}

func (c *Controller) Clear() {
	// If we are in edit mode, exiting the modal is the whole job of Clear.
	// leaveEditMode restores the pre-edit state across all the per-VFO fields.
	if c.editing {
		c.leaveEditMode()
		c.showInput()
		c.view.SetMyCall(c.stationCallsign)
		c.view.SetFrequency(c.focusedVFO, c.selectedFrequency[c.focusedVFO])
		c.view.SetActiveField(c.activeField[c.focusedVFO])
		c.view.SetDuplicateMarker(false)
		c.view.SetEditingMarker(false)
		c.view.ClearMessage()
		c.selectLastQSO()
		c.notifyCallinfoInputChanged("", core.NoBand, core.NoMode, []string{})
		return
	}

	// Release before the wipe so claim slots reflect "nothing pending" while we rebuild.
	// The companion refresh writes the serial display below, after the input zero/fill pass.
	c.claimedSerial[c.focusedVFO] = 0
	c.claimSnapshot[c.focusedVFO] = 0

	c.vfos[c.focusedVFO].Refresh()

	c.activeField[c.focusedVFO] = core.CallsignField
	c.input[c.focusedVFO].callsign = ""
	if c.selectedBand[c.focusedVFO] != core.NoBand {
		c.input[c.focusedVFO].band = c.selectedBand[c.focusedVFO].String()
	}
	generatedReport := ""
	if c.selectedMode[c.focusedVFO] != core.NoMode {
		c.input[c.focusedVFO].mode = c.selectedMode[c.focusedVFO].String()
		generatedReport = defaultReportForMode(c.selectedMode[c.focusedVFO])
	}

	c.input[c.focusedVFO].myReport = ""
	c.input[c.focusedVFO].myNumber = ""
	c.input[c.focusedVFO].theirReport = ""
	c.input[c.focusedVFO].theirNumber = ""
	c.input[c.focusedVFO].theirExchange = make([]string, len(c.theirExchangeFields))
	c.input[c.focusedVFO].myExchange = make([]string, len(c.myExchangeFields))
	lastExchange := c.logbook.LastExchange()
	for i, value := range c.defaultExchangeValues {
		if value == "" && i < len(lastExchange) {
			value = lastExchange[i]
		}

		if i >= len(c.myExchangeFields) {
			continue
		}

		c.input[c.focusedVFO].myExchange[i] = value
		if i == c.myReportExchangeField.Field.ExchangeIndex()-1 {
			if c.generateReport {
				value = generatedReport
			}
			c.input[c.focusedVFO].myReport = value
			c.input[c.focusedVFO].myExchange[i] = value

			c.input[c.focusedVFO].theirExchange[i] = value
			c.input[c.focusedVFO].theirReport = value
		}
	}
	// Refresh serial displays for both VFOs (other VFO's preview may shift after the release).
	c.refreshMyNumberInputs()

	c.updateESM()

	c.showInput()
	c.view.SetMyCall(c.stationCallsign)
	c.view.SetFrequency(c.focusedVFO, c.selectedFrequency[c.focusedVFO])
	c.view.SetActiveField(c.activeField[c.focusedVFO])
	c.view.SetDuplicateMarker(false)
	c.view.SetEditingMarker(false)
	c.view.ClearMessage()
	c.selectLastQSO()
	c.notifyCallinfoInputChanged("", core.NoBand, core.NoMode, []string{})
}

func (c *Controller) Activate() {
	c.view.SetActiveField(c.activeField[c.focusedVFO])
}

func (c *Controller) EditLastQSO() {
	c.activeField[c.focusedVFO] = core.CallsignField
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

func (c *Controller) CurrentValues() core.KeyerValues {
	myNumber, _ := strconv.Atoi(c.input[c.focusedVFO].myNumber)

	myXchanges := make([]string, 0, len(c.input[c.focusedVFO].myExchange))
	for i, field := range c.myExchangeFields {
		switch field.Field {
		case c.myReportExchangeField.Field, c.myNumberExchangeField.Field:
			continue
		default:
			myXchanges = append(myXchanges, c.input[c.focusedVFO].myExchange[i])
		}
	}

	values := core.KeyerValues{}
	values.MyReport, _ = parse.RST(c.input[c.focusedVFO].myReport)
	values.MyNumber = core.QSONumber(myNumber)
	values.MyXchange = strings.Join(myXchanges, " ")
	values.MyExchange = strings.Join(c.input[c.focusedVFO].myExchange, " ")
	values.MyExchanges = c.input[c.focusedVFO].myExchange
	values.TheirCall = c.input[c.focusedVFO].callsign

	return values
}

func (c *Controller) StationChanged(station core.Station) {
	c.stationCallsign = station.Callsign.String()
	c.view.SetMyCall(c.stationCallsign)
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.updateExchangeFields(contest)
}

func (c *Controller) WorkmodeChanged(workmode core.Workmode) {
	c.workmode = workmode
	c.updateESM()
}

func (c *Controller) updateExchangeFields(contest core.Contest) {
	c.myExchangeFields = contest.MyExchangeFields
	c.myReportExchangeField = contest.MyReportExchangeField
	c.myNumberExchangeField = contest.MyNumberExchangeField
	c.theirExchangeFields = contest.TheirExchangeFields
	c.theirReportExchangeField = contest.TheirReportExchangeField
	c.theirNumberExchangeField = contest.TheirNumberExchangeField
	c.generateSerialExchange = contest.GenerateSerialExchange
	c.generateReport = contest.GenerateReport
	c.defaultExchangeValues = contest.ExchangeValues

	for vfo := range core.VFOCount {
		c.input[vfo].myExchange = make([]string, len(contest.MyExchangeFields))
		c.input[vfo].theirExchange = make([]string, len(contest.TheirExchangeFields))
	}

	c.Clear()
}

func (c *Controller) MarkInBandmap() {
	call, err := core.ParseCallsign(c.input[c.focusedVFO].callsign)
	if err != nil {
		log.Printf("Cannot mark invalid call: %v", err)
		return
	}
	spot := core.Spot{
		Call:      call,
		Frequency: c.selectedFrequency[c.focusedVFO],
		Band:      c.selectedBand[c.focusedVFO],
		Mode:      c.selectedMode[c.focusedVFO],
		Time:      c.clock.Now(),
		Source:    core.ManualSpot,
	}
	c.bandmap.Add(spot)
}

func (c *Controller) EntrySelected(entry core.BandmapEntry) {
	// TODO: check if the entry's band is currently selected in one of the two VFOs

	c.Clear()
	c.ignoreFrequencyJump = true
	c.frequencyEntered(entry.Frequency)
	c.SetActiveField(core.CallsignField)
	c.Enter(entry.Call.String())
	c.view.SetCallsign(c.input[c.focusedVFO].callsign)
}
