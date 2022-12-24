package entry

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ftl/conval"
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
	Log(core.QSO)
}

// QSOList functionality used for QSO entry.
type QSOList interface {
	Find(callsign.Callsign, core.Band, core.Mode) []core.QSO
	FindDuplicateQSOs(callsign.Callsign, core.Band, core.Mode) []core.QSO
	SelectQSO(core.QSO)
	SelectLastQSO()
}

// Keyer functionality used for QSO entry.
type Keyer interface {
	SendQuestion(q string)
	Stop()
}

// Callinfo functionality used for QSO entry.
type Callinfo interface {
	ShowInfo(call string, band core.Band, mode core.Mode, xchange string)
	PredictedXchange() string
}

// VFO functionality used for QSO entry.
type VFO interface {
	Active() bool
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

// NewController returns a new entry controller.
func NewController(settings core.Settings, clock core.Clock, qsoList QSOList, asyncRunner core.AsyncRunner) *Controller {
	contest := settings.Contest()
	result := &Controller{
		clock:       clock,
		view:        new(nullView),
		logbook:     new(nullLogbook),
		callinfo:    new(nullCallinfo),
		vfo:         new(nullVFO),
		asyncRunner: asyncRunner,
		qsoList:     qsoList,

		stationCallsign:     settings.Station().Callsign.String(),
		enableTheirNumber:   contest.EnterTheirNumber,
		enableTheirXchange:  contest.EnterTheirXchange,
		requireTheirXchange: contest.RequireTheirXchange,
	}
	result.refreshTicker = ticker.New(result.refreshUTC)
	if contest.Definition != nil {
		result.updateExchangeFields(contest.Definition, contest.GenerateSerialExchange, contest.ExchangeValues)
	}
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

	asyncRunner   core.AsyncRunner
	refreshTicker *ticker.Ticker

	stationCallsign     string
	enableTheirNumber   bool
	enableTheirXchange  bool
	requireTheirXchange bool

	myExchangeFields         []core.ExchangeField
	theirExchangeFields      []core.ExchangeField
	myReportExchangeField    core.ExchangeField
	myNumberExchangeField    core.ExchangeField
	theirReportExchangeField core.ExchangeField
	theirNumberExchangeField core.ExchangeField
	generateSerialExchange   bool
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
	if c.selectedBand == core.NoBand || !c.vfo.Active() {
		lastBand := c.logbook.LastBand()
		if lastBand != core.NoBand {
			c.selectedBand = lastBand
			c.input.band = lastBand.String()
		} else {
			c.selectedBand = core.Band160m
			c.input.band = c.selectedBand.String()
		}
	}
	if c.selectedMode == core.NoMode || !c.vfo.Active() {
		lastMode := c.logbook.LastMode()
		if lastMode != core.NoMode {
			c.selectedMode = lastMode
			c.input.mode = lastMode.String()
		} else {
			c.selectedMode = core.ModeCW
			c.input.mode = c.selectedMode.String()
		}
	}

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

func (c *Controller) leaveCallsignField() {
	callsign, err := callsign.Parse(c.input.callsign)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO fill new exchange fields with predicted values
	// if c.enableTheirXchange && c.input.theirXchange == "" {
	// 	c.setTheirXchangePrediction(c.callinfo.PredictedXchange())
	// }

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
	c.view.SetBand(c.input.band)
	c.view.SetMode(c.input.mode)
}

func (c *Controller) setTheirXchangePrediction(predictedXchange string) {
	// TODO: fill the new exchange fields
	// c.input.theirXchange = predictedXchange
	// c.view.SetTheirXchange(c.input.theirXchange)
}

func (c *Controller) selectQSO(qso core.QSO) {
	c.ignoreQSOSelection = true
	c.qsoList.SelectQSO(qso)
	c.ignoreQSOSelection = false
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
		// TODO: c.enterTheirXchange -> update callinfo
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
	if c.selectedBand == band {
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
		c.enterCallsign(c.input.callsign)
	}
}

func (c *Controller) SetMode(mode core.Mode) {
	if c.selectedMode == mode {
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
		c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, "") // c.input.theirXchange) // TODO use new exchange fields
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

func (c *Controller) enterTheirXchange(s string) {
	// TODO: also handle input in new exchange fields
	if c.callinfo == nil {
		return
	}
	c.callinfo.ShowInfo(c.input.callsign, c.selectedBand, c.selectedMode, "") // c.input.theirXchange) // TODO use new exchange fields
	// c.clearErrorOnField(core.TheirXchangeField)
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
	c.callinfo.ShowInfo(qso.Callsign.String(), qso.Band, qso.Mode, "") // qso.TheirXchange) // TODO use new exchange fields
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
	for i, field := range c.theirExchangeFields {
		value := c.input.theirExchange[i]
		if value == "" {
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
				qso.TheirNumber = core.QSONumber(theirNumber)
				break
			}
			if len(field.Properties) == 1 {
				c.showErrorOnField(err, field.Field)
				return
			}
		default:
			// TODO check the predicted value
			// qso.TheirXchange = c.input.theirXchange
			// if qso.TheirXchange == "" && c.requireTheirXchange {
			// 	predictedXchange := c.callinfo.PredictedXchange()
			// 	if predictedXchange != "" {
			// 		c.setTheirXchangePrediction(predictedXchange)
			// 		c.showErrorOnField(fmt.Errorf("check their exhange"), core.TheirXchangeField)
			// 		return
			// 	}
			// 	c.showErrorOnField(errors.New("their exchange is missing"), core.TheirXchangeField)
			// 	return
			// }
		}
	}

	// handle my exchange
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
			if err != nil {
				c.showErrorOnField(err, field.Field)
				return
			}
			qso.MyNumber = core.QSONumber(myNumber)
		}
	}

	c.logbook.Log(qso)
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

	nextNumber := c.logbook.NextNumber()
	c.activeField = core.CallsignField
	c.input.callsign = ""
	if c.selectedBand != core.NoBand {
		c.input.band = c.selectedBand.String()
	}
	if c.selectedMode != core.NoMode {
		c.input.mode = c.selectedMode.String()
	}

	c.input.myReport = ""
	c.input.myNumber = ""
	c.input.theirReport = ""
	c.input.theirNumber = ""
	for i := range c.input.theirExchange {
		c.input.theirExchange[i] = ""
	}
	for i, value := range c.defaultExchangeValues {
		c.input.myExchange[i] = value
		if i == c.myReportExchangeField.Field.ExchangeIndex()-1 {
			c.input.myReport = value

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
		c.callinfo.ShowInfo("", core.NoBand, core.NoMode, "")
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
	values.TheirCall = c.input.callsign

	return values
}

func (c *Controller) StationChanged(station core.Station) {
	c.stationCallsign = station.Callsign.String()
	c.view.SetMyCall(c.stationCallsign)
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.enableTheirNumber = contest.EnterTheirNumber
	c.enableTheirXchange = contest.EnterTheirXchange
	c.requireTheirXchange = contest.RequireTheirXchange

	c.updateExchangeFields(contest.Definition, contest.GenerateSerialExchange, contest.ExchangeValues)
}

func (c *Controller) updateExchangeFields(definition *conval.Definition, generateSerialExchange bool, defaultExchangeValues []string) {
	c.myExchangeFields = nil
	c.myReportExchangeField = core.ExchangeField{}
	c.myNumberExchangeField = core.ExchangeField{}
	c.input.myExchange = nil
	c.theirExchangeFields = nil
	c.theirReportExchangeField = core.ExchangeField{}
	c.theirNumberExchangeField = core.ExchangeField{}
	c.input.theirExchange = nil
	c.generateSerialExchange = generateSerialExchange
	c.defaultExchangeValues = defaultExchangeValues

	if definition == nil {
		c.updateViewExchangeFields()
		return
	}

	fieldDefinitions := definition.ExchangeFields()

	c.myExchangeFields = core.DefinitionsToExchangeFields(fieldDefinitions, core.MyExchangeField)
	for i, field := range c.myExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.myReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			if generateSerialExchange {
				field.ReadOnly = true
				field.Short = "#"
				field.Hint = "Serial Number"
				c.myExchangeFields[i] = field
			}
			c.myNumberExchangeField = field
		}
	}
	c.input.myExchange = make([]string, len(c.myExchangeFields))

	c.theirExchangeFields = core.DefinitionsToExchangeFields(fieldDefinitions, core.TheirExchangeField)
	for _, field := range c.theirExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.theirReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			c.theirNumberExchangeField = field
		}
	}
	c.input.theirExchange = make([]string, len(c.theirExchangeFields))

	c.updateViewExchangeFields()
}

func (c *Controller) updateViewExchangeFields() {
	c.view.SetMyExchangeFields(c.myExchangeFields)
	c.view.SetTheirExchangeFields(c.theirExchangeFields)
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
func (n *nullView) SetDuplicateMarker(bool)                     {}
func (n *nullView) SetEditingMarker(bool)                       {}
func (n *nullView) ShowMessage(...interface{})                  {}
func (n *nullView) ClearMessage()                               {}

type nullVFO struct{}

func (n *nullVFO) Active() bool                { return false }
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

func (n *nullCallinfo) ShowInfo(string, core.Band, core.Mode, string) {}
func (n *nullCallinfo) PredictedXchange() string                      { return "" }
