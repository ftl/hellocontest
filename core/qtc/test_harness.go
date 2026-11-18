package qtc

import (
	"fmt"

	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/model"
)

// FAKES

type fakeInfoDialogs struct {
	answer bool
}

func (*fakeInfoDialogs) ShowInfo(string, ...any)            {}
func (d *fakeInfoDialogs) ShowQuestion(string, ...any) bool { return d.answer }
func (*fakeInfoDialogs) ShowError(string, ...any)           {}

type fakeQTCList struct{}

func (l *fakeQTCList) SelectLastQTC()      {}
func (l *fakeQTCList) SetQTCsEnabled(bool) {}

type fakeEntryController struct {
	currentCallsign model.Callsign
	currentState    model.QSODataState
	frequency       model.Frequency
	band            model.Band
	mode            model.Mode
	vfoName         string
	so2v            bool
}

func (c *fakeEntryController) CurrentQSOState() (model.Callsign, model.QSODataState) {
	return c.currentCallsign, c.currentState
}
func (c *fakeEntryController) CurrentVFOState() (model.Frequency, model.Band, model.Mode) {
	return c.frequency, c.band, c.mode
}
func (c *fakeEntryController) FocusedVFO() (string, bool) {
	return c.vfoName, c.so2v
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
	lastCallsign     model.Callsign
	loggedQTCs       []model.QTC
	availableQTCs    []model.QTC
}

func (l *fakeLogbook) NextSeriesNumber() int        { return l.nextSeriesNumber }
func (l *fakeLogbook) LastCallsign() model.Callsign { return l.lastCallsign }
func (l *fakeLogbook) AddQTCSeries(series model.QTCSeries) {
	for _, qtc := range series.QTCs {
		l.loggedQTCs = append(l.loggedQTCs, qtc)
	}
}
func (l *fakeLogbook) PrepareFor(model.Callsign, int) []model.QTC { return l.availableQTCs }

type fakeView struct {
	visible        bool
	mode           model.QTCMode
	series         model.QTCSeries
	phase          model.QTCWorkflowPhase
	field          model.QTCField
	activeQTCIndex int
	errorMessage   string
	errorField     model.QTCField
	vfoName        string
}

func (v *fakeView) QuestionQTCCount(max int) (int, bool) {
	return max, true
}
func (v *fakeView) ShowFieldError(field model.QTCField, s string) {
	v.errorField = field
	v.errorMessage = s
}
func (v *fakeView) ClearFieldError() {
	v.errorField = model.QTCNoneField
	v.errorMessage = ""
}
func (v *fakeView) Show(mode model.QTCMode, series model.QTCSeries, vfoName string) {
	v.visible = true
	v.mode = mode
	v.series = series
	v.vfoName = vfoName
}
func (v *fakeView) UpdateQTC(index int, qtc model.QTC) {
	v.series.SetData(index, qtc)
}
func (v *fakeView) Close() {
	v.visible = false
}
func (v *fakeView) ClearDataInputs() {
	// no-op
}
func (v *fakeView) SetActivePhase(phase model.QTCWorkflowPhase) {
	v.phase = phase
}
func (v *fakeView) SetActiveField(field model.QTCField) {
	v.field = field
}
func (v *fakeView) SetActiveQTC(index int) {
	v.activeQTCIndex = index
}

// BUILDER

type controllerBuilder struct {
	clock           model.Clock
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

func (b *controllerBuilder) WithClock(clock model.Clock) *controllerBuilder {
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

func (b *controllerBuilder) WithQTCs(qtcs []model.QTC) *controllerBuilder {
	b.logbook = &fakeLogbook{
		availableQTCs: qtcs,
	}
	return b
}

func (b *controllerBuilder) WithEntryController(entryController EntryController) *controllerBuilder {
	b.entryController = entryController
	return b
}

func (b *controllerBuilder) WithCurrentQSOState(call model.Callsign, state model.QSODataState) *controllerBuilder {
	b.entryController = &fakeEntryController{
		currentCallsign: call,
		currentState:    state,
	}
	return b
}

func (b *controllerBuilder) WithVFOState(frequency model.Frequency, band model.Band, mode model.Mode) *controllerBuilder {
	fake, ok := b.entryController.(*fakeEntryController)
	if !ok {
		fake = &fakeEntryController{}
		b.entryController = fake
	}
	fake.frequency = frequency
	fake.band = band
	fake.mode = mode
	return b
}

func (b *controllerBuilder) WithFocusedVFO(name string, so2v bool) *controllerBuilder {
	fake, ok := b.entryController.(*fakeEntryController)
	if !ok {
		fake = &fakeEntryController{}
		b.entryController = fake
	}
	fake.vfoName = name
	fake.so2v = so2v
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

func (b *controllerBuilder) WithLastCallsign(lastCallsign model.Callsign) *controllerBuilder {
	b.logbook = &fakeLogbook{
		lastCallsign: lastCallsign,
	}
	return b
}

func (b *controllerBuilder) Build() *Controller {
	logbook := b.logbook
	if logbook == nil {
		logbook = new(fakeLogbook)
	}
	result := NewController(b.clock, b.infoDialogs, logbook, b.qtcList, b.entryController, b.keyer)
	return result
}

type qtcBuilder struct {
	kind          model.QTCKind
	theirCallsign model.Callsign
	qtcs          []model.QTC
}

func qtcsFor(kind model.QTCKind, theirCallsign model.Callsign) *qtcBuilder {
	return &qtcBuilder{
		kind:          kind,
		theirCallsign: theirCallsign,
		qtcs:          []model.QTC{},
	}
}

func (b *qtcBuilder) Add(time, call string, qtcNumber model.QSONumber) *qtcBuilder {
	var lastQTC model.QTC
	if len(b.qtcs) > 0 {
		lastQTC = b.qtcs[len(b.qtcs)-1]
	}
	qtcTime, err := model.ParseQTCTime(time, lastQTC.QTCTime)
	if err != nil {
		panic(err)
	}
	qtcCallsign := model.MustParseCallsign(call)
	qtc := model.QTC{
		Kind:          b.kind,
		TheirCallsign: b.theirCallsign,
		QTCTime:       qtcTime,
		QTCCallsign:   qtcCallsign,
		QTCNumber:     qtcNumber,
	}
	b.qtcs = append(b.qtcs, qtc)
	return b
}

func (b *qtcBuilder) Build() []model.QTC {
	return b.qtcs
}
