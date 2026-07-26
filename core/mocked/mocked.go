package mocked

import (
	"time"

	"github.com/ftl/hamradio/dxcc"
	"github.com/stretchr/testify/mock"

	"github.com/ftl/hellocontest/core"
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

func (m *Log) NextQSONumber() core.QSONumber {
	if !m.active {
		return core.QSONumber(0)
	}
	args := m.Called()
	return args.Get(0).(core.QSONumber)
}

func (m *Log) LastBand() core.Band {
	if !m.active {
		return core.NoBand
	}
	args := m.Called()
	return args.Get(0).(core.Band)
}

func (m *Log) LastMode() core.Mode {
	if !m.active {
		return core.NoMode
	}
	args := m.Called()
	return args.Get(0).(core.Mode)
}

func (m *Log) LastExchange() []string {
	if !m.active {
		return nil
	}
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *Log) AddQSO(qso core.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) UpdateQSO(qso core.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) FindDuplicateQSOs(call core.Callsign, band core.Band, mode core.Mode) []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called(call, band, mode)
	return args.Get(0).([]core.QSO)
}

type QSOList struct {
	mock.Mock
	active bool
}

func (m *QSOList) Activate() {
	m.active = true
}

func (m *QSOList) Find(callsign core.Callsign, band core.Band, mode core.Mode) []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]core.QSO)
}

func (m *QSOList) FindWorkedQSOs(callsign core.Callsign, band core.Band, mode core.Mode) ([]core.QSO, bool) {
	if !m.active {
		return []core.QSO{}, false
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]core.QSO), args.Get(1).(bool)
}

func (m *QSOList) SelectQSO(qso core.QSO) {
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

func (m *QSOList) LastBandAndMode() (core.Band, core.Mode) {
	if !m.active {
		return core.NoBand, core.NoMode
	}
	args := m.Called()
	return args.Get(0).(core.Band), args.Get(1).(core.Mode)
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

func (m *LogbookView) UpdateAllRows(qsos []core.QSO) {
	m.Called(qsos)
}

func (m *LogbookView) RowAdded(qso core.QSO) {
	m.Called(qso)
}

type Reader struct {
	mock.Mock
}

func (m *Reader) ReadAllQSOs() ([]core.QSO, error) {
	args := m.Called()
	return args.Get(0).([]core.QSO), args.Error(1)
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

func (m *EntryView) SetMyCall(mycall string) {
	if !m.active {
		return
	}
	m.Called(mycall)
}

func (m *EntryView) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	if !m.active {
		return
	}
	m.Called(vfo, frequency)
}

func (m *EntryView) SetCallsign(vfo core.VFOID, callsign string) {
	if !m.active {
		return
	}
	m.Called(vfo, callsign)
}

func (m *EntryView) SetTheirExchange(vfo core.VFOID, index int, value string) {
	if !m.active {
		return
	}
	m.Called(vfo, index, value)
}

func (m *EntryView) SetBand(vfo core.VFOID, text string) {
	if !m.active {
		return
	}
	m.Called(vfo, text)
}

func (m *EntryView) SetMode(vfo core.VFOID, text string) {
	if !m.active {
		return
	}
	m.Called(vfo, text)
}

func (m *EntryView) SetTXState(vfo core.VFOID, ptt bool, parrotActive bool, parrotTimeLeft time.Duration) {
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

func (m *EntryView) SetMyExchangeFields(fields []core.ExchangeField) {
	if !m.active {
		return
	}
	m.Called(fields)
}

func (m *EntryView) SetTheirExchangeFields(fields []core.ExchangeField) {
	if !m.active {
		return
	}
	m.Called(fields)
}

func (m *EntryView) SetSerialClaim(vfo core.VFOID, n core.QSONumber, committed bool) {
	if !m.active {
		return
	}
	m.Called(vfo, n, committed)
}

func (m *EntryView) SetActiveVFO(vfo core.VFOID) {
	if !m.active {
		return
	}
	m.Called(vfo)
}

func (m *EntryView) SetActiveField(vfo core.VFOID, field core.EntryField) {
	if !m.active {
		return
	}
	m.Called(vfo, field)
}

func (m *EntryView) SelectText(vfo core.VFOID, field core.EntryField, s string) {
	if !m.active {
		return
	}
	m.Called(vfo, field, s)
}

func (m *EntryView) SetDuplicateMarker(vfo core.VFOID, active bool) {
	if !m.active {
		return
	}
	m.Called(vfo, active)
}

func (m *EntryView) SetEditingMarker(vfo core.VFOID, active bool) {
	if !m.active {
		return
	}
	m.Called(vfo, active)
}

func (m *EntryView) ShowMessage(vfo core.VFOID, args ...interface{}) {
	if !m.active {
		return
	}
	m.Called(vfo, args)
}

func (m *EntryView) ClearMessage(vfo core.VFOID) {
	if !m.active {
		return
	}
	m.Called(vfo)
}

func (m *EntryView) SetVFOEnabled(vfo core.VFOID, enabled bool) {
	if !m.active {
		return
	}
	m.Called(vfo, enabled)
}

func (m *EntryView) SetVFOWorkmode(vfo core.VFOID, workmode core.Workmode) {
	if !m.active {
		return
	}
	m.Called(vfo, workmode)
}

func (m *EntryView) SetTXVFO(vfo core.VFOID) {
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

type DXCCFinder struct {
	mock.Mock
}

func (m *DXCCFinder) Find(callsign string) (dxcc.Prefix, bool) {
	args := m.Called(callsign)
	return args.Get(0).(dxcc.Prefix), args.Get(1).(bool)
}
