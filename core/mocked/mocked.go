package mocked

import (
	"time"

	"github.com/ftl/hamradio/callsign"
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

func (m *Log) ClearRowSelectedListeners() {
	if !m.active {
		return
	}
	m.Called()
}

func (m *Log) Select(i int) {
	if !m.active {
		return
	}
	m.Called(i)
}

func (m *Log) SelectQSO(qso core.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) SelectLastQSO() {
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

func (m *Log) Log(qso core.QSO) {
	if !m.active {
		return
	}
	m.Called(qso)
}

func (m *Log) Find(callsign callsign.Callsign) (core.QSO, bool) {
	if !m.active {
		return core.QSO{}, false
	}
	args := m.Called(callsign)
	return args.Get(0).(core.QSO), args.Bool(1)
}

func (m *Log) FindAll(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	if !m.active {
		return []core.QSO{}
	}
	args := m.Called(callsign, band, mode)
	return args.Get(0).([]core.QSO)
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

func (m *Reader) ReadAll() ([]core.QSO, error) {
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

func (m *EntryView) SetCallsign(callsign string) {
	if !m.active {
		return
	}
	m.Called(callsign)
}

func (m *EntryView) SetTheirReport(report string) {
	if !m.active {
		return
	}
	m.Called(report)
}

func (m *EntryView) SetTheirNumber(number string) {
	if !m.active {
		return
	}
	m.Called(number)
}

func (m *EntryView) SetTheirXchange(xchange string) {
	if !m.active {
		return
	}
	m.Called(xchange)
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

func (m *EntryView) SetMyReport(report string) {
	if !m.active {
		return
	}
	m.Called(report)
}

func (m *EntryView) SetMyNumber(number string) {
	if !m.active {
		return
	}
	m.Called(number)
}

func (m *EntryView) SetMyXchange(xchange string) {
	if !m.active {
		return
	}
	m.Called(xchange)
}

func (m *EntryView) EnableExchangeFields(theirNumber, theirXchange bool) {
	if !m.active {
		return
	}
	m.Called(theirNumber, theirXchange)
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
