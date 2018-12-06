package mocked

import (
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/mock"
)

//nolint

type Log struct {
	mock.Mock
}

func (m *Log) SetView(view core.LogView) {
	m.Called(view)
}

func (m *Log) OnRowAdded(listener core.RowAddedListener) {
	m.Called(listener)
}

func (m *Log) ClearRowAddedListeners() {
	m.Called()
}

func (m *Log) GetNextNumber() core.QSONumber {
	args := m.Called()
	return args.Get(0).(core.QSONumber)
}

func (m *Log) Log(qso core.QSO) {
	m.Called(qso)
}

func (m *Log) Find(callsign callsign.Callsign) (core.QSO, bool) {
	args := m.Called(callsign)
	return args.Get(0).(core.QSO), args.Bool(1)
}

func (m *Log) GetQsosByMyNumber() []core.QSO {
	args := m.Called()
	return args.Get(0).([]core.QSO)
}

func (m *Log) WriteAll(writer core.Writer) error {
	args := m.Called(writer)
	return args.Error(0)
}

type LogView struct {
	mock.Mock
}

func (m *LogView) SetLog(log core.Log) {
	m.Called(log)
}

func (m *LogView) UpdateAllRows(qsos []core.QSO) {
	m.Called(qsos)
}

func (m *LogView) RowAdded(qso core.QSO) {
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
}

func (m *EntryView) SetEntryController(controller core.EntryController) {
	m.Called(controller)
}

func (m *EntryView) GetCallsign() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetCallsign(callsign string) {
	m.Called(callsign)
}

func (m *EntryView) GetTheirReport() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetTheirReport(report string) {
	m.Called(report)
}

func (m *EntryView) GetTheirNumber() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetTheirNumber(number string) {
	m.Called(number)
}

func (m *EntryView) GetBand() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetBand(text string) {
	m.Called(text)
}

func (m *EntryView) GetMode() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetMode(text string) {
	m.Called(text)
}

func (m *EntryView) GetMyReport() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetMyReport(report string) {
	m.Called(report)
}

func (m *EntryView) GetMyNumber() string {
	args := m.Called()
	return args.String(0)
}

func (m *EntryView) SetMyNumber(number string) {
	m.Called(number)
}

func (m *EntryView) SetActiveField(field core.EntryField) {
	m.Called(field)
}

func (m *EntryView) SetDuplicateMarker(active bool) {
	m.Called(active)
}

func (m *EntryView) ShowError(err error) {
	m.Called(err)
}

func (m *EntryView) ClearError() {
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

func (m *CWClient) Send(text string) {
	m.Called(text)
}
