package qtc

import (
	"fmt"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
)

// FAKES

type fakeInfoDialogs struct {
	answer bool
}

func (*fakeInfoDialogs) ShowInfo(string, ...any)            {}
func (d *fakeInfoDialogs) ShowQuestion(string, ...any) bool { return d.answer }
func (*fakeInfoDialogs) ShowError(string, ...any)           {}

type fakeQTCList struct {
	qtcs []core.QTC
}

func (l *fakeQTCList) PrepareFor(core.Callsign, int) []core.QTC { return l.qtcs }

type fakeEntryController struct {
	currentCallsign core.Callsign
	currentState    core.QSODataState
}

func (c *fakeEntryController) CurrentQSOState() (core.Callsign, core.QSODataState) {
	return c.currentCallsign, c.currentState
}
func (c *fakeEntryController) Log() {}

type fakeKeyer struct {
	lastTransmission string
}

func (k *fakeKeyer) SendText(format string, args ...any) {
	k.lastTransmission = fmt.Sprintf(format, args...)
}
func (*fakeKeyer) Repeat()             {}
func (*fakeKeyer) Stop()               {}
func (*fakeKeyer) Cut(s string) string { return s }

type fakeLogbook struct {
	nextSeriesNumber int
	lastCallsign     core.Callsign
}

func (l *fakeLogbook) NextSeriesNumber() int       { return l.nextSeriesNumber }
func (l *fakeLogbook) LastCallsign() core.Callsign { return l.lastCallsign }
func (*fakeLogbook) LogQTC(core.QTC)               {}

type fakeView struct {
	visible        bool
	mode           core.QTCMode
	series         core.QTCSeries
	phase          core.QTCWorkflowPhase
	activeQTCIndex int
}

func (v *fakeView) QuestionQTCCount(max int) (int, bool) {
	return max, true
}
func (v *fakeView) Show(mode core.QTCMode, series core.QTCSeries) {
	v.visible = true
	v.mode = mode
	v.series = series
}
func (v *fakeView) UpdateQTC(index int, qtc core.QTC) {
	// TODO implement
}
func (v *fakeView) Close() {
	v.visible = false
}
func (v *fakeView) SetActivePhase(phase core.QTCWorkflowPhase) {
	v.phase = phase
}
func (v *fakeView) SetActiveQTC(index int) {
	v.activeQTCIndex = index
}

// BUILDER

type controllerBuilder struct {
	clock           core.Clock
	infoDialogs     InfoDialogs
	qtcList         QTCList
	entryController EntryController
	keyer           Keyer
	logbook         Logbook
}

func newController() *controllerBuilder {
	return &controllerBuilder{
		clock:           clock.New(),
		infoDialogs:     &fakeInfoDialogs{},
		qtcList:         &fakeQTCList{},
		entryController: &fakeEntryController{},
		keyer:           &fakeKeyer{},
	}
}

func (b *controllerBuilder) WithClock(clock core.Clock) *controllerBuilder {
	b.clock = clock
	return b
}

func (b *controllerBuilder) WithInfoDialogs(infoDialogs InfoDialogs) *controllerBuilder {
	b.infoDialogs = infoDialogs
	return b
}

func (b *controllerBuilder) WithAnswer(answer bool) *controllerBuilder {
	b.infoDialogs = &fakeInfoDialogs{
		answer: answer,
	}
	return b
}

func (b *controllerBuilder) WithQTCList(qtcList QTCList) *controllerBuilder {
	b.qtcList = qtcList
	return b
}

func (b *controllerBuilder) WithQTCs(qtcs []core.QTC) *controllerBuilder {
	b.qtcList = &fakeQTCList{
		qtcs: qtcs,
	}
	return b
}

func (b *controllerBuilder) WithEntryController(entryController EntryController) *controllerBuilder {
	b.entryController = entryController
	return b
}

func (b *controllerBuilder) WithCurrentQSOState(call core.Callsign, state core.QSODataState) *controllerBuilder {
	b.entryController = &fakeEntryController{
		currentCallsign: call,
		currentState:    state,
	}
	return b
}

func (b *controllerBuilder) WithKeyer(keyer Keyer) *controllerBuilder {
	b.keyer = keyer
	return b
}

func (b *controllerBuilder) WithLogbook(logbook Logbook) *controllerBuilder {
	b.logbook = logbook
	return b
}

func (b *controllerBuilder) WithNextSeriesNumber(nextSeriesNumber int) *controllerBuilder {
	b.logbook = &fakeLogbook{
		nextSeriesNumber: nextSeriesNumber,
	}
	return b
}

func (b *controllerBuilder) WithLastCallsign(lastCallsign core.Callsign) *controllerBuilder {
	b.logbook = &fakeLogbook{
		lastCallsign: lastCallsign,
	}
	return b
}

func (b *controllerBuilder) Build() *Controller {
	result := NewController(b.clock, b.infoDialogs, b.qtcList, b.entryController, b.keyer)
	if b.logbook != nil {
		result.SetLogbook(b.logbook)
	}
	return result
}

type qtcBuilder struct {
	kind          core.QTCKind
	theirCallsign core.Callsign
	qtcs          []core.QTC
}

func qtcsFor(kind core.QTCKind, theirCallsign core.Callsign) *qtcBuilder {
	return &qtcBuilder{
		kind:          kind,
		theirCallsign: theirCallsign,
		qtcs:          []core.QTC{},
	}
}

func (b *qtcBuilder) Add(time, call string, qtcNumber core.QSONumber) *qtcBuilder {
	var lastQTC core.QTC
	if len(b.qtcs) > 0 {
		lastQTC = b.qtcs[len(b.qtcs)-1]
	}
	qtcTime, err := core.ParseQTCTime(time, lastQTC.QTCTime)
	if err != nil {
		panic(err)
	}
	qtcCallsign := callsign.MustParse(call)
	qtc := core.QTC{
		Kind:          b.kind,
		TheirCallsign: b.theirCallsign,
		QTCTime:       qtcTime,
		QTCCallsign:   qtcCallsign,
		QTCNumber:     qtcNumber,
	}
	b.qtcs = append(b.qtcs, qtc)
	return b
}

func (b *qtcBuilder) Build() []core.QTC {
	return b.qtcs
}
