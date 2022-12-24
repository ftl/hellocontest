package mocked

import (
	"time"

	"github.com/ftl/hamradio/callsign"
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

func (m *Log) NextNumber() core.QSONumber {
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

func (m *Log) Log(qso core.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) QsosOrderedByMyNumber() []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called()
	return args.Get(0).([]core.QSO)
}

func (m *Log) UniqueQsosOrderedByMyNumber() []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called()
	return args.Get(0).([]core.QSO)
}

type QSOList struct {
	mock.Mock
	active bool
}

func (m *QSOList) Activate() {
	m.active = true
}

func (m *QSOList) Find(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]core.QSO)
}

func (m *QSOList) FindDuplicateQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]core.QSO)
}

func (m *QSOList) FindWorkedQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) ([]core.QSO, bool) {
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

func (m *EntryView) SetFrequency(frequency core.Frequency) {
	if !m.active {
		return
	}
	m.Called(frequency)
}

func (m *EntryView) SetCallsign(callsign string) {
	if !m.active {
		return
	}
	m.Called(callsign)
}

func (m *EntryView) SetTheirExchange(index int, value string) {
	if !m.active {
		return
	}
	m.Called(index, value)
}

func (m *EntryView) SetBand(text string) {
	if !m.active {
		return
	}
	m.Called(text)
}

func (m *EntryView) SetMode(text string) {
	if !m.active {
		return
	}
	m.Called(text)
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

func (m *EntryView) SetActiveField(field core.EntryField) {
	if !m.active {
		return
	}
	m.Called(field)
}

func (m *EntryView) SetDuplicateMarker(active bool) {
	if !m.active {
		return
	}
	m.Called(active)
}

func (m *EntryView) SetEditingMarker(active bool) {
	if !m.active {
		return
	}
	m.Called(active)
}

func (m *EntryView) ShowMessage(args ...interface{}) {
	if !m.active {
		return
	}
	m.Called(args)
}

func (m *EntryView) ClearMessage() {
	if !m.active {
		return
	}
	m.Called()
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

type DXCCFinder struct {
	mock.Mock
}

func (m *DXCCFinder) Find(callsign string) (dxcc.Prefix, bool) {
	args := m.Called(callsign)
	return args.Get(0).(dxcc.Prefix), args.Get(1).(bool)
}
