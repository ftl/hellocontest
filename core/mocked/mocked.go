package mocked

import (
	"time"

	"github.com/ftl/hamradio/dxcc"
	"github.com/stretchr/testify/mock"

	"github.com/ftl/hellocontest/core/model"
)

//nolint

type Log struct {
	mock.Mock
	active bool
}

func (m *Log) Activate() {
	m.active = true
}

func (m *Log) ClearRowAddedListeners() {
	if !m.active {
		return
	}
	m.Called()
}

func (m *Log) NextQSONumber() model.QSONumber {
	if !m.active {
		return model.QSONumber(0)
	}
	args := m.Called()
	return args.Get(0).(model.QSONumber)
}

func (m *Log) LastBand() model.Band {
	if !m.active {
		return model.NoBand
	}
	args := m.Called()
	return args.Get(0).(model.Band)
}

func (m *Log) LastMode() model.Mode {
	if !m.active {
		return model.NoMode
	}
	args := m.Called()
	return args.Get(0).(model.Mode)
}

func (m *Log) LastExchange() []string {
	if !m.active {
		return nil
	}
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *Log) AddQSO(qso model.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) UpdateQSO(qso model.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) FindDuplicateQSOs(call model.Callsign, band model.Band, mode model.Mode) []model.QSO {
	if !m.active {
		return []model.QSO{}
	}
	args := m.Called(call, band, mode)
	return args.Get(0).([]model.QSO)
}

type QSOList struct {
	mock.Mock
	active bool
}

func (m *QSOList) Activate() {
	m.active = true
}

func (m *QSOList) Find(callsign model.Callsign, band model.Band, mode model.Mode) []model.QSO {
	if !m.active {
		return []model.QSO{}
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]model.QSO)
}

func (m *QSOList) FindWorkedQSOs(callsign model.Callsign, band model.Band, mode model.Mode) ([]model.QSO, bool) {
	if !m.active {
		return []model.QSO{}, false
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]model.QSO), args.Get(1).(bool)
}

func (m *QSOList) SelectQSO(qso model.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *QSOList) SelectLastQSO() {
	if !m.active {
		return
	}
	m.Called()
}

func (m *QSOList) LastBandAndMode() (model.Band, model.Mode) {
	if !m.active {
		return model.NoBand, model.NoMode
	}
	args := m.Called()
	return args.Get(0).(model.Band), args.Get(1).(model.Mode)
}

type AppView struct {
	mock.Mock
}

func (m *AppView) ShowFilename(filename string) {
	m.Called(filename)
}

func (m *AppView) SelectOpenFile(title string, patterns ...string) (string, bool, error) {
	args := m.Called(title, patterns)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *AppView) SelectSaveFile(title string, patterns ...string) (string, bool, error) {
	args := m.Called(title, patterns)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *AppView) ShowInfoDialog(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *AppView) ShowErrorDialog(format string, args ...interface{}) {
	m.Called(format, args)
}

type LogbookView struct {
	mock.Mock
}

func (m *LogbookView) UpdateAllRows(qsos []model.QSO) {
	m.Called(qsos)
}

func (m *LogbookView) RowAdded(qso model.QSO) {
	m.Called(qso)
}

type Reader struct {
	mock.Mock
}

func (m *Reader) ReadAllQSOs() ([]model.QSO, error) {
	args := m.Called()
	return args.Get(0).([]model.QSO), args.Error(1)
}

type EntryView struct {
	mock.Mock
	active bool
}

func (m *EntryView) Activate() {
	m.active = true
}

func (m *EntryView) SetUTC(utc string) {
	if !m.active {
		return
	}
	m.Called(utc)
}

func (m *EntryView) SetFrequency(vfo model.VFOID, frequency model.Frequency) {
	if !m.active {
		return
	}
	m.Called(vfo, frequency)
}

func (m *EntryView) SetCallsign(vfo model.VFOID, callsign string) {
	if !m.active {
		return
	}
	m.Called(vfo, callsign)
}

func (m *EntryView) SetTheirExchange(vfo model.VFOID, index int, value string) {
	if !m.active {
		return
	}
	m.Called(vfo, index, value)
}

func (m *EntryView) SetBand(vfo model.VFOID, text string) {
	if !m.active {
		return
	}
	m.Called(vfo, text)
}

func (m *EntryView) SetMode(vfo model.VFOID, text string) {
	if !m.active {
		return
	}
	m.Called(vfo, text)
}

func (m *EntryView) SetTXState(vfo model.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
	if !m.active {
		return
	}
	m.Called(vfo, ptt, parrotActive, parrotTimeLeft)
}

func (m *EntryView) SetMyExchange(index int, value string) {
	if !m.active {
		return
	}
	m.Called(index, value)
}

func (m *EntryView) SetMyExchangeFields(fields []model.ExchangeField) {
	if !m.active {
		return
	}
	m.Called(fields)
}

func (m *EntryView) SetTheirExchangeFields(fields []model.ExchangeField) {
	if !m.active {
		return
	}
	m.Called(fields)
}

func (m *EntryView) SetSerialClaim(vfo model.VFOID, n model.QSONumber, committed bool) {
	if !m.active {
		return
	}
	m.Called(vfo, n, committed)
}

func (m *EntryView) SetActiveVFO(vfo model.VFOID) {
	if !m.active {
		return
	}
	m.Called(vfo)
}

func (m *EntryView) SetActiveField(vfo model.VFOID, field model.EntryField) {
	if !m.active {
		return
	}
	m.Called(vfo, field)
}

func (m *EntryView) SelectText(vfo model.VFOID, field model.EntryField, s string) {
	if !m.active {
		return
	}
	m.Called(vfo, field, s)
}

func (m *EntryView) SetDuplicateMarker(vfo model.VFOID, active bool) {
	if !m.active {
		return
	}
	m.Called(vfo, active)
}

func (m *EntryView) SetEditingMarker(vfo model.VFOID, active bool) {
	if !m.active {
		return
	}
	m.Called(vfo, active)
}

func (m *EntryView) ShowMessage(vfo model.VFOID, args ...interface{}) {
	if !m.active {
		return
	}
	m.Called(vfo, args)
}

func (m *EntryView) ClearMessage(vfo model.VFOID) {
	if !m.active {
		return
	}
	m.Called(vfo)
}

func (m *EntryView) SetVFOEnabled(vfo model.VFOID, enabled bool) {
	if !m.active {
		return
	}
	m.Called(vfo, enabled)
}

func (m *EntryView) SetVFOWorkmode(vfo model.VFOID, workmode model.Workmode) {
	if !m.active {
		return
	}
	m.Called(vfo, workmode)
}

func (m *EntryView) SetTXVFO(vfo model.VFOID) {
	if !m.active {
		return
	}
	m.Called(vfo)
}

type Clock struct {
	mock.Mock
}

func (m *Clock) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

type CWClient struct {
	mock.Mock
}

func (m *CWClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *CWClient) Disconnect() {
	m.Called()
}

func (m *CWClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *CWClient) Speed(speed int) {
	m.Called(speed)
}

func (m *CWClient) Send(text string) {
	m.Called(text)
}

func (m *CWClient) Abort() {
	m.Called()
}

type KeyerView struct {
	mock.Mock
}

func (m *KeyerView) ShowMessage(args ...interface{}) {
	m.Called(args)
}

func (m *KeyerView) Pattern(index int) string {
	args := m.Called(index)
	return args.String(0)
}

func (m *KeyerView) SetLabel(index int, pattern string) {
	m.Called(index, pattern)
}

func (m *KeyerView) SetPattern(index int, pattern string) {
	m.Called(index, pattern)
}

func (m *KeyerView) Speed() int {
	args := m.Called()
	return args.Int(0)
}

func (m *KeyerView) SetSpeed(speed int) {
	m.Called(speed)
}

func (m *KeyerView) SetPresetNames(names []string) {
	m.Called(names)
}

func (m *KeyerView) SetPreset(name string) {
	m.Called(name)
}

func (m *KeyerView) SetLastTransmission(s string) {
	m.Called(s)
}

type DXCCFinder struct {
	mock.Mock
}

func (m *DXCCFinder) Find(callsign string) (dxcc.Prefix, bool) {
	args := m.Called(callsign)
	return args.Get(0).(dxcc.Prefix), args.Get(1).(bool)
}
