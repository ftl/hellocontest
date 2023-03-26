package entry

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
	"github.com/ftl/hellocontest/core/ticker"
)

// View represents the visual part of the QSO data entry.
type View interface {
	SetUTC(string)
	SetMyCall(string)
	SetFrequency(core.Frequency)
	SetCallsign(string)
	SetBand(text string)
	SetMode(text string)
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
	LastBand() core.Band
	LastMode() core.Mode
	LastExchange() []string
	Log(core.QSO)
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	Find(callsign.Callsign, core.Band, core.Mode) []core.QSO
	FindDuplicateQSOs(callsign.Callsign, core.Band, core.Mode) []core.QSO
	SelectQSO(core.QSO)
	SelectLastQSO()
	LastBandAndMode() (core.Band, core.Mode)
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	ShowInfo(call string, band core.Band, mode core.Mode, exchange []string)
	PredictedExchange() []string
}

// VFO functionality used for QSO entry.
type VFO interface {
	Active() bool
	Refresh()
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

type Bandmap interface {
	Add(core.Spot)
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
	vfo      VFO
	bandmap  Bandmap

	asyncRunner   core.AsyncRunner
	refreshTicker *ticker.Ticker

	stationCallsign string

	myExchangeFields         []core.ExchangeField
	theirExchangeFields      []core.ExchangeField
	myReportExchangeField    core.ExchangeField
	myNumberExchangeField    core.ExchangeField
	theirReportExchangeField core.ExchangeField
	theirNumberExchangeField core.ExchangeField
	generateSerialExchange   bool
	generateReport           bool
	defaultExchangeValues    []string

	input              input
	activeField        core.EntryField
	errorField         core.EntryField
	selectedFrequency  core.Frequency
	selectedBand       core.Band
	selectedMode       core.Mode
	editing            bool
	editQSO            core.QSO
	ignoreQSOSelection bool
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = &nullView{}
		return
	}
	c.view = view
	c.Clear()
	c.refreshUTC()
	c.updateViewExchangeFields()
}

func (c *Controller) SetLogbook(logbook Logbook) {
	c.logbook = logbook

	if c.vfo.Active() {
		c.vfo.Refresh()
	}

	lastBand, lastMode := c.qsoList.LastBandAndMode()
	if c.selectedBand == core.NoBand {
		c.selectedBand = lastBand
		c.selectedFrequency = 0
	}
	if c.selectedMode == core.NoMode {
		c.selectedMode = lastMode
	}

	if c.selectedBand == core.NoBand {
		c.selectedBand = core.Band160m
	}
	if c.selectedMode == core.NoMode {
		c.selectedMode = core.ModeCW
	}

	c.input.band = c.selectedBand.String()
	c.input.mode = c.selectedMode.String()

	c.showInput()
}

func (c *Controller) SetKeyer(keyer Keyer) {
	c.keyer = keyer
}

func (c *Controller) SetCallinfo(callinfo Callinfo) {
	c.callinfo = callinfo
}

func (c *Controller) SetVFO(vfo VFO) {
	if vfo == nil {
		c.vfo = new(nullVFO)
	}
	c.vfo = vfo
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

	predictedExchange := c.callinfo.PredictedExchange()
	if len(c.input.theirExchange) == len(predictedExchange) {
		for i, field := range c.theirExchangeFields {
			switch field.Field {
			case c.theirReportExchangeField.Field:
				continue
			case c.theirNumberExchangeField.Field:
				if len(c.theirNumberExchangeField.Properties) == 1 {
					continue
				}
			}
			if c.input.theirExchange[i] == "" && predictedExchange[i] != "" {
				c.setTheirExchangePrediction(i, predictedExchange[i])
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

func (c *Controller) StartAutoRefresh() {
	c.refreshTicker.Start()
}

func (c *Controller) refreshUTC() {
	c.asyncRunner(func() {
		if c.view == nil {
			return
		}

		utc := c.clock.Now().UTC()
		c.view.SetUTC(utc.Format("15:04"))
	})
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

func (c *Controller) frequencySelected(frequency core.Frequency) {
	log.Printf("Frequency selected: %s", frequency)
	c.selectedFrequency = frequency
	c.vfo.SetFrequency(frequency)
	c.input.callsign = ""
	c.enterCallsign(c.input.callsign)
	c.view.SetCallsign(c.input.callsign)
	c.view.SetFrequency(frequency)
}

func (c *Controller) SetFrequency(frequency core.Frequency) {
	if c.editing {
		return
	}
	if c.selectedFrequency == frequency {
		return
	}
	c.selectedFrequency = frequency
	c.view.SetFrequency(c.selectedFrequency)
}

func (c *Controller) bandSelected(s string) {
	if band, err := parse.Band(s); err == nil {
		log.Printf("Band selected: %v", band)
		c.selectedBand = band
		c.vfo.SetBand(band)
		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) SetBand(band core.Band) {
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
	generatedReport := defaultReportForMode(c.selectedMode)
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

func (c *Controller) SetMode(mode core.Mode) {
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
	if c.callinfo != nil {
		c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirExchange)
	}

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
	c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, c.input.theirExchange)
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
	c.callinfo.ShowInfo(qso.Callsign.String(), qso.Band, qso.Mode, qso.TheirExchange)
}

func (c *Controller) Log() {
	if f, ok := parseKilohertz(c.input.callsign); ok && c.activeField == core.CallsignField {
		c.frequencySelected(f)
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
	predictedExchange := c.callinfo.PredictedExchange()

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
			if qso.TheirExchange[i] == "" && len(predictedExchange) == len(qso.TheirExchange) && predictedExchange[i] != "" {
				c.setTheirExchangePrediction(i, predictedExchange[i])
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

	c.logbook.Log(qso)

	if !c.vfo.Active() {
		c.selectedBand, c.selectedMode = c.qsoList.LastBandAndMode()
	}

	spot := core.Spot{
		Call:      qso.Callsign,
		Frequency: qso.Frequency,
		Band:      qso.Band,
		Mode:      qso.Mode,
		Time:      qso.Time,
		Source:    core.WorkedSpot,
	}
	c.bandmap.Add(spot)

	c.Clear()
}

func parseKilohertz(s string) (core.Frequency, bool) {
	kHz, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return core.Frequency(kHz * 1000), true
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

	if c.vfo.Active() {
		c.vfo.Refresh()
	}

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
	if c.callinfo != nil {
		c.callinfo.ShowInfo("", core.NoBand, core.NoMode, []string{})
	}
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

func (c *Controller) FilterExchange(values []string) []string {
	result := make([]string, len(values))
	for i := range values {
		if i >= len(c.theirExchangeFields) {
			break
		}
		field := c.theirExchangeFields[i]
		switch field.Field {
		case c.theirReportExchangeField.Field:
			continue
		case c.theirNumberExchangeField.Field:
			if len(field.Properties) == 1 {
				continue
			}
		}
		result[i] = values[i]
	}
	return result
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

type nullView struct{}

func (n *nullView) SetUTC(string)                               {}
func (n *nullView) SetMyCall(string)                            {}
func (n *nullView) SetFrequency(core.Frequency)                 {}
func (n *nullView) SetCallsign(string)                          {}
func (n *nullView) SetBand(text string)                         {}
func (n *nullView) SetMode(text string)                         {}
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

func (n *nullVFO) Active() bool                { return false }
func (n *nullVFO) Refresh()                    {}
func (n *nullVFO) SetFrequency(core.Frequency) {}
func (n *nullVFO) SetBand(core.Band)           {}
func (n *nullVFO) SetMode(core.Mode)           {}

type nullLogbook struct{}

func (n *nullLogbook) NextNumber() core.QSONumber { return 0 }
func (n *nullLogbook) LastBand() core.Band        { return core.NoBand }
func (n *nullLogbook) LastMode() core.Mode        { return core.NoMode }
func (n *nullLogbook) LastExchange() []string     { return nil }
func (n *nullLogbook) Log(core.QSO)               {}

type nullCallinfo struct{}

func (n *nullCallinfo) ShowInfo(string, core.Band, core.Mode, []string) {}
func (n *nullCallinfo) PredictedExchange() []string                     { return []string{} }

type nullBandmap struct{}

func (n *nullBandmap) Add(core.Spot) {}
