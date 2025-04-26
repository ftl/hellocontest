package entry

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/callsign"

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
	SetFrequency(core.Frequency)
	SetCallsign(string)
	SetBand(text string)
	SetMode(text string)
	SetXITActive(active bool)
	SetXIT(active bool, offset core.Frequency)
	SetMyExchange(int, string)
	SetTheirExchange(int, string)

	SetMyExchangeFields([]core.ExchangeField)
	SetTheirExchangeFields([]core.ExchangeField)
	SetActiveField(core.EntryField)
	SelectText(core.EntryField, string)
	SetDuplicateMarker(bool)
	SetEditingMarker(bool)
	ShowMessage(...interface{})
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
	NextNumber() core.QSONumber
	LastExchange() []string
	Log(core.QSO)
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	FindDuplicateQSOs(callsign.Callsign, core.Band, core.Mode) []core.QSO
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	InputChanged(call string, band core.Band, mode core.Mode, exchange []string)
}

type Bandmap interface {
	Add(core.Spot)
	SelectByCallsign(callsign.Callsign)
}

// NewController returns a new entry controller.
func NewController(settings core.Settings, clock core.Clock, qsoList QSOList, bandmap Bandmap, asyncRunner core.AsyncRunner) *Controller {
	result := &Controller{
		clock:       clock,
		view:        new(nullView),
		logbook:     new(nullLogbook),
		callinfo:    new(nullCallinfo),
		vfo:         new(nullVFO),
		asyncRunner: asyncRunner,
		qsoList:     qsoList,
		bandmap:     bandmap,

		stationCallsign: settings.Station().Callsign.String(),
	}
	result.refreshTicker = ticker.New(result.refreshUTC)
	result.updateExchangeFields(settings.Contest())
	return result
}

type Controller struct {
	clock    core.Clock
	view     View
	logbook  Logbook
	qsoList  QSOList
	keyer    Keyer
	callinfo Callinfo
	vfo      core.VFO
	bandmap  Bandmap

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
	currentCallinfoFrame     core.CallinfoFrame

	input               input
	activeField         core.EntryField
	errorField          core.EntryField
	selectedFrequency   core.Frequency
	selectedBand        core.Band
	selectedMode        core.Mode
	editing             bool
	editQSO             core.QSO
	ignoreQSOSelection  bool
	ignoreFrequencyJump bool
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
	c.updateViewExchangeFields()
}

func (c *Controller) SetLogbook(logbook Logbook) {
	c.logbook = logbook
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
	c.currentCallinfoFrame = frame
	// TODO what do we need to update here?
}

func (c *Controller) SetVFO(vfo core.VFO) {
	if vfo == nil {
		c.vfo = new(nullVFO)
	} else {
		c.vfo = vfo
	}
	vfo.Notify(c)
}

func (c *Controller) GotoNextField() core.EntryField {
	switch c.activeField {
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

	nextField := transitions[c.activeField]
	if nextField == "" {
		nextField = core.CallsignField
	}

	c.activeField = nextField
	c.view.SetActiveField(c.activeField)
	return c.activeField
}

func (c *Controller) GotoNextPlaceholder() {
	c.activeField = core.CallsignField
	c.view.SetActiveField(c.activeField)
	c.view.SelectText(c.activeField, core.FilterPlaceholder)
}

func (c *Controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.input.callsign)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(c.input.theirExchange) == len(c.currentCallinfoFrame.PredictedExchange) {
		for i, field := range c.theirExchangeFields {
			if !c.isPredictable(field.Field) {
				continue
			}
			if c.input.theirExchange[i] == "" {
				c.setTheirExchangePrediction(i, c.currentCallinfoFrame.PredictedExchange[i])
			}
		}
	}

	_, found := c.isDuplicate(callsign)
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
	c.notifyCallinfoInputChanged(c.input.callsign, c.selectedBand, c.selectedMode, []string{})

	if len(c.input.theirExchange) == len(c.currentCallinfoFrame.PredictedExchange) {
		for i, field := range c.theirExchangeFields {
			if !c.isPredictable(field.Field) {
				continue
			}
			c.setTheirExchangePrediction(i, c.currentCallinfoFrame.PredictedExchange[i])
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
	c.input.callsign = qso.Callsign.String()
	c.input.theirReport = qso.TheirReport.String()
	c.input.theirNumber = qso.TheirNumber.String()
	c.input.theirExchange = ensureLen(qso.TheirExchange, len(c.theirExchangeFields))
	c.input.myReport = qso.MyReport.String()
	c.input.myNumber = qso.MyNumber.String()
	c.input.myExchange = ensureLen(qso.MyExchange, len(c.myExchangeFields))
	c.input.band = qso.Band.String()
	c.input.mode = qso.Mode.String()

	c.selectedFrequency = qso.Frequency
	c.selectedBand = qso.Band
	c.selectedMode = qso.Mode

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
	c.view.SetCallsign(c.input.callsign)
	for i, value := range c.input.theirExchange {
		c.view.SetTheirExchange(i+1, value)
	}
	for i, value := range c.input.myExchange {
		c.view.SetMyExchange(i+1, value)
	}
	c.view.SetFrequency(c.selectedFrequency)
	c.view.SetBand(c.input.band)
	c.view.SetMode(c.input.mode)
}

// setTheirExchangePrediction replaces the value of the given field with the given predicted value,
// if the given value is not empty.
func (c *Controller) setTheirExchangePrediction(i int, value string) {
	if value == "" {
		return
	}
	c.input.theirExchange[i] = value
	c.view.SetTheirExchange(i+1, value)
}

func (c *Controller) isDuplicate(callsign callsign.Callsign) (core.QSO, bool) {
	qsos := c.qsoList.FindDuplicateQSOs(callsign, c.selectedBand, c.selectedMode)
	if len(qsos) == 0 {
		return core.QSO{}, false
	}
	return qsos[len(qsos)-1], true
}

func (c *Controller) SetActiveField(field core.EntryField) {
	c.activeField = field
}

func (c *Controller) SelectMatch(index int) {
	c.selectCallsign(c.currentCallinfoFrame.GetMatch(index))
}

func (c *Controller) SelectBestMatchOnFrequency() {
	c.selectCallsign(c.currentCallinfoFrame.BestMatchOnFrequency().Callsign.String())
}

func (c *Controller) selectCallsign(callsign string) {
	if callsign == "" {
		return
	}
	c.activeField = core.CallsignField
	c.Enter(callsign)
	c.view.SetCallsign(c.input.callsign)
	c.GotoNextField()
}

func (c *Controller) Enter(text string) {
	switch c.activeField {
	case core.CallsignField:
		c.input.callsign = text
		c.enterCallsign(text)
	case core.BandField:
		c.input.band = text
		c.bandSelected(text)
	case core.ModeField:
		c.input.mode = text
		c.modeSelected(text)
	}

	i := c.activeField.ExchangeIndex() - 1
	switch {
	case c.activeField.IsMyExchange():
		c.input.myExchange[i] = text
	case c.activeField.IsTheirExchange():
		c.input.theirExchange[i] = text
		c.enterTheirExchange(c.activeField)
	}
}

func (c *Controller) frequencyEntered(frequency core.Frequency) {
	// log.Printf("Frequency selected: %s", frequency)
	c.vfo.SetFrequency(frequency)
}

func (c *Controller) bandEntered(band core.Band) {
	c.input.band = band.String()
	c.vfo.SetBand(band)
}

func (c *Controller) SetXITActive(active bool) {
	c.vfo.SetXITActive(active)
}

func (c *Controller) VFOFrequencyChanged(frequency core.Frequency) {
	if c.editing {
		return
	}
	if c.selectedFrequency == frequency {
		return
	}
	jump := math.Abs(float64(c.selectedFrequency-frequency)) > float64(jumpThreshold)
	c.selectedFrequency = frequency

	c.view.SetFrequency(frequency)

	if jump && !c.ignoreFrequencyJump {
		c.Clear()
		c.activeField = core.CallsignField
		c.view.SetActiveField(c.activeField)
	}
	c.ignoreFrequencyJump = false
}

func (c *Controller) bandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		// log.Printf("Band selected: %v", band)
		c.selectedBand = band
		c.vfo.SetBand(band)
		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) VFOBandChanged(band core.Band) {
	if c.editing {
		return
	}
	if band == core.NoBand || band == c.selectedBand {
		return
	}
	c.selectedBand = band
	c.input.band = c.selectedBand.String()
	c.view.SetBand(c.input.band)
}

func (c *Controller) modeSelected(s string) {
	if mode, err := parse.Mode(s); err == nil {
		log.Printf("Mode selected: %v", mode)
		c.selectedMode = mode

		c.vfo.SetMode(mode)
		if c.generateReport {
			c.generateReportForMode(mode)
		}
		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) generateReportForMode(mode core.Mode) {
	generatedReport := defaultReportForMode(mode)
	myIndex := c.myReportExchangeField.Field.ExchangeIndex()
	if myIndex > 0 {
		c.input.myReport = generatedReport
		c.input.myExchange[myIndex-1] = generatedReport
		c.view.SetMyExchange(myIndex, c.input.myReport)
	}
	theirIndex := c.theirReportExchangeField.Field.ExchangeIndex()
	if theirIndex > 0 {
		c.input.theirReport = generatedReport
		c.input.theirExchange[theirIndex-1] = generatedReport
		c.view.SetTheirExchange(theirIndex, c.input.myReport)
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

func (c *Controller) VFOModeChanged(mode core.Mode) {
	if c.editing {
		return
	}
	if mode == core.NoMode || mode == c.selectedMode {
		return
	}
	c.selectedMode = mode
	c.input.mode = c.selectedMode.String()
	c.view.SetMode(c.input.mode)
}

func (c *Controller) VFOXITChanged(active bool, offset core.Frequency) {
	c.view.SetXIT(active, offset)
}

func (c *Controller) SendQuestion() {
	if c.keyer == nil {
		return
	}

	switch {
	case c.activeField.IsTheirExchange():
		c.keyer.SendQuestion("nr")
	default:
		c.keyer.SendQuestion(c.input.callsign)
	}
}

func (c *Controller) enterCallsign(s string) {
	c.emitCallsignEntered(c.input.callsign)
	c.notifyCallinfoInputChanged(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirExchange)

	callsign, err := callsign.Parse(s)
	if err != nil {
		return
	}

	qso, found := c.isDuplicate(callsign)
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
	c.notifyCallinfoInputChanged(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirExchange)
	c.clearErrorOnField(field)
}

func (c *Controller) QSOSelected(qso core.QSO) {
	if c.ignoreQSOSelection {
		return
	}

	log.Printf("QSO selected: %v", qso)
	c.editing = true
	c.editQSO = qso

	c.showQSO(qso)
	c.view.SetActiveField(core.CallsignField)
	c.view.SetEditingMarker(true)
	c.notifyCallinfoInputChanged(qso.Callsign.String(), qso.Band, qso.Mode, qso.TheirExchange)
}

func (c *Controller) Log() {
	if c.parseCallsignCommand() {
		c.input.callsign = ""
		c.enterCallsign(c.input.callsign)
		c.view.SetCallsign(c.input.callsign)
		return
	}

	var err error
	qso := core.QSO{}
	if c.editing {
		qso.Time = c.editQSO.Time
	} else {
		qso.Time = c.clock.Now()
	}

	qso.Callsign, err = callsign.Parse(c.input.callsign)
	if err != nil {
		c.showErrorOnField(err, core.CallsignField)
		return
	}

	qso.Frequency = c.selectedFrequency

	qso.Band, err = parse.Band(c.input.band)
	if err != nil {
		c.showErrorOnField(err, core.BandField)
		return
	}

	qso.Mode, err = parse.Mode(c.input.mode)
	if err != nil {
		c.showErrorOnField(err, core.ModeField)
		return
	}

	// handle their exchange
	qso.TheirExchange = make([]string, len(c.theirExchangeFields))

	for i, field := range c.theirExchangeFields {
		value := c.input.theirExchange[i]
		if value == "" && !field.EmptyAllowed {
			c.showErrorOnField(fmt.Errorf("%s is missing", field.Short), field.Field) // TODO use field.Name
			return
		}

		// TODO parse the value using the conval validators and show an error on the field

		qso.TheirExchange[i] = value

		switch field.Field {
		case c.theirReportExchangeField.Field:
			qso.TheirReport, err = parse.RST(value)
			if err != nil {
				c.showErrorOnField(err, field.Field)
				return
			}
		case c.theirNumberExchangeField.Field:
			theirNumber, err := strconv.Atoi(value)
			if err == nil {
				qso.TheirExchange[i] = fmt.Sprintf("%03d", theirNumber)
				qso.TheirNumber = core.QSONumber(theirNumber)
			} else if len(field.Properties) == 1 {
				c.showErrorOnField(err, field.Field)
				return
			}
		default:
			if qso.TheirExchange[i] == "" && len(c.currentCallinfoFrame.PredictedExchange) == len(qso.TheirExchange) {
				c.setTheirExchangePrediction(i, c.currentCallinfoFrame.PredictedExchange[i])
				c.showErrorOnField(fmt.Errorf("check their exchange"), field.Field)
				return
			}
		}
	}

	// handle my exchange
	myNumber, err := strconv.Atoi(c.input.myNumber)
	if err == nil {
		qso.MyNumber = core.QSONumber(myNumber)
	}
	qso.MyExchange = make([]string, len(c.myExchangeFields))
	for i, field := range c.myExchangeFields {
		value := c.input.myExchange[i]
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
	if !c.editing {
		qso.Workmode = c.workmode
	}

	c.logbook.Log(qso)
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
	if c.activeField != core.CallsignField {
		return false
	}

	if f, ok := parseKilohertz(c.input.callsign); ok {
		c.frequencyEntered(f)
		return true
	}

	if b, err := parse.Band(c.input.callsign); err == nil {
		c.bandEntered(b)
		return true
	}

	if call, ok := parseBandmapCallsign(c.input.callsign); ok {
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

func parseBandmapCallsign(s string) (callsign.Callsign, bool) {
	if !strings.HasPrefix(s, "@") {
		return callsign.Callsign{}, false
	}

	call, err := callsign.Parse(s[1:])
	if err != nil {
		log.Printf("invalid bandmap callsign: %v", err)
		return callsign.Callsign{}, false
	}
	return call, true
}

func (c *Controller) showErrorOnField(err error, field core.EntryField) {
	c.activeField = field
	c.errorField = field
	c.view.SetActiveField(c.activeField)
	c.view.ShowMessage(err)
}

func (c *Controller) clearErrorOnField(field core.EntryField) {
	if c.errorField != field {
		return
	}
	c.view.ClearMessage()
}

func (c *Controller) Clear() {
	c.editing = false
	c.editQSO = core.QSO{}

	c.vfo.Refresh()

	nextNumber := c.logbook.NextNumber()
	c.activeField = core.CallsignField
	c.input.callsign = ""
	if c.selectedBand != core.NoBand {
		c.input.band = c.selectedBand.String()
	}
	generatedReport := ""
	if c.selectedMode != core.NoMode {
		c.input.mode = c.selectedMode.String()
		generatedReport = defaultReportForMode(c.selectedMode)
	}

	c.input.myReport = ""
	c.input.myNumber = ""
	c.input.theirReport = ""
	c.input.theirNumber = ""
	c.input.theirExchange = make([]string, len(c.theirExchangeFields))
	c.input.myExchange = make([]string, len(c.myExchangeFields))
	lastExchange := c.logbook.LastExchange()
	for i, value := range c.defaultExchangeValues {
		if value == "" && i < len(lastExchange) {
			value = lastExchange[i]
		}

		if i >= len(c.myExchangeFields) {
			continue
		}

		c.input.myExchange[i] = value
		if i == c.myReportExchangeField.Field.ExchangeIndex()-1 {
			if c.generateReport {
				value = generatedReport
			}
			c.input.myReport = value
			c.input.myExchange[i] = value

			c.input.theirExchange[i] = value
			c.input.theirReport = value
		}
	}
	c.setMyNumberInput(nextNumber.String())

	c.showInput()
	c.view.SetMyCall(c.stationCallsign)
	c.view.SetFrequency(c.selectedFrequency)
	c.view.SetActiveField(c.activeField)
	c.view.SetDuplicateMarker(false)
	c.view.SetEditingMarker(false)
	c.view.ClearMessage()
	c.selectLastQSO()
	c.notifyCallinfoInputChanged("", core.NoBand, core.NoMode, []string{})
}

func (c *Controller) setMyNumberInput(value string) {
	c.input.myNumber = value
	i := c.myNumberExchangeField.Field.ExchangeIndex() - 1
	if i < 0 || !c.generateSerialExchange {
		return
	}
	c.input.myExchange[i] = value
}

func (c *Controller) Activate() {
	c.view.SetActiveField(c.activeField)
}

func (c *Controller) EditLastQSO() {
	c.activeField = core.CallsignField
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
	myNumber, _ := strconv.Atoi(c.input.myNumber)

	myXchanges := make([]string, 0, len(c.input.myExchange))
	for i, field := range c.myExchangeFields {
		switch field.Field {
		case c.myReportExchangeField.Field, c.myNumberExchangeField.Field:
			continue
		default:
			myXchanges = append(myXchanges, c.input.myExchange[i])
		}
	}

	values := core.KeyerValues{}
	values.MyReport, _ = parse.RST(c.input.myReport)
	values.MyNumber = core.QSONumber(myNumber)
	values.MyXchange = strings.Join(myXchanges, " ")
	values.MyExchange = strings.Join(c.input.myExchange, " ")
	values.MyExchanges = c.input.myExchange
	values.TheirCall = c.input.callsign

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
	log.Printf("ENTRY: workmode changed %d", workmode)
	c.workmode = workmode
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

	c.input.myExchange = make([]string, len(contest.MyExchangeFields))
	c.input.theirExchange = make([]string, len(contest.TheirExchangeFields))

	c.updateViewExchangeFields()
}

func (c *Controller) updateViewExchangeFields() {
	c.view.SetMyExchangeFields(c.myExchangeFields)
	c.view.SetTheirExchangeFields(c.theirExchangeFields)
}

func (c *Controller) MarkInBandmap() {
	call, err := callsign.Parse(c.input.callsign)
	if err != nil {
		log.Printf("Cannot mark invalid call: %v", err)
		return
	}
	spot := core.Spot{
		Call:      call,
		Frequency: c.selectedFrequency,
		Band:      c.selectedBand,
		Mode:      c.selectedMode,
		Time:      c.clock.Now(),
		Source:    core.ManualSpot,
	}
	c.bandmap.Add(spot)
}

func (c *Controller) EntrySelected(entry core.BandmapEntry) {
	c.Clear()
	c.ignoreFrequencyJump = true
	c.frequencyEntered(entry.Frequency)
	c.activeField = core.CallsignField
	c.Enter(entry.Call.String())
	c.view.SetCallsign(c.input.callsign)
	c.GotoNextField()
}

type nullView struct{}

func (n *nullView) SetUTC(string)                               {}
func (n *nullView) SetMyCall(string)                            {}
func (n *nullView) SetFrequency(core.Frequency)                 {}
func (n *nullView) SetCallsign(string)                          {}
func (n *nullView) SetBand(text string)                         {}
func (n *nullView) SetMode(text string)                         {}
func (n *nullView) SetXITActive(active bool)                    {}
func (n *nullView) SetXIT(active bool, offset core.Frequency)   {}
func (n *nullView) SetMyExchange(int, string)                   {}
func (n *nullView) SetTheirExchange(int, string)                {}
func (n *nullView) SetMyExchangeFields([]core.ExchangeField)    {}
func (n *nullView) SetTheirExchangeFields([]core.ExchangeField) {}
func (n *nullView) SetActiveField(core.EntryField)              {}
func (n *nullView) SelectText(core.EntryField, string)          {}
func (n *nullView) SetDuplicateMarker(bool)                     {}
func (n *nullView) SetEditingMarker(bool)                       {}
func (n *nullView) ShowMessage(...interface{})                  {}
func (n *nullView) ClearMessage()                               {}

type nullVFO struct{}

func (n *nullVFO) Notify(any)                  {}
func (n *nullVFO) Active() bool                { return false }
func (n *nullVFO) Refresh()                    {}
func (n *nullVFO) SetFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)           {}
func (n *nullVFO) SetMode(core.Mode)           {}
func (n *nullVFO) SetXIT(bool, core.Frequency) {}
func (n *nullVFO) XITActive() bool             { return false }
func (n *nullVFO) SetXITActive(bool)           {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() core.QSONumber { return 0 }
func (n *nullLogbook) LastBand() core.Band        { return core.NoBand }
func (n *nullLogbook) LastMode() core.Mode        { return core.NoMode }
func (n *nullLogbook) LastExchange() []string     { return nil }
func (n *nullLogbook) Log(core.QSO)               {}

type nullCallinfo struct{}

func (n *nullCallinfo) InputChanged(string, core.Band, core.Mode, []string) {}

type nullBandmap struct{}

func (n *nullBandmap) Add(core.Spot)                      {}
func (n *nullBandmap) SelectByCallsign(callsign.Callsign) {}
