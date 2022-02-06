package settings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"

	"github.com/ftl/hellocontest/core"
)

type StationListener interface {
	StationChanged(core.Station)
}

type StationListenerFunc func(core.Station)

func (f StationListenerFunc) StationChanged(station core.Station) {
	f(station)
}

type ContestListener interface {
	ContestChanged(core.Contest)
}

type ContestListenerFunc func(core.Contest)

func (f ContestListenerFunc) ContestChanged(contest core.Contest) {
	f(contest)
}

type SettingsListener interface {
	SettingsChanged(core.Settings)
}

type SettingsListenerFunc func(core.Settings)

func (f SettingsListenerFunc) SettingsChanged(settings core.Settings) {
	f(settings)
}

type Writer interface {
	WriteStation(core.Station) error
	WriteContest(core.Contest) error
}

type DefaultsOpener func()

type XchangeRegexpMatcher func(*regexp.Regexp, string) (string, bool)

type View interface {
	Show()
	ShowMessage(string)
	HideMessage()

	SetStationCallsign(string)
	SetStationOperator(string)
	SetStationLocator(string)
	SetContestName(string)
	SetContestEnterTheirNumber(bool)
	SetContestEnterTheirXchange(bool)
	SetContestRequireTheirXchange(bool)
	SetContestAllowMultiBand(bool)
	SetContestAllowMultiMode(bool)
	SetContestSameCountryPoints(string)
	SetContestSameContinentPoints(string)
	SetContestSpecificCountryPoints(string)
	SetContestSpecificCountryPrefixes(string)
	SetContestOtherPoints(string)
	SetContestMultis(dxcc, wpx, xchange bool)
	SetContestXchangeMultiPattern(string)
	SetContestXchangeMultiPatternResult(string)
	SetContestCountPerBand(bool)
	SetContestCallHistoryFile(string)
	SetContestCallHistoryField(string)
	SetContestCabrilloQSOTemplate(string)
}

func New(defaultsOpener DefaultsOpener, xchangeRegexpMatcher XchangeRegexpMatcher, station core.Station, contest core.Contest) *Settings {
	return &Settings{
		writer:               new(nullWriter),
		defaultsOpener:       defaultsOpener,
		xchangeRegexpMatcher: xchangeRegexpMatcher,
		station:              station,
		contest:              contest,
		defaultStation:       station,
		defaultContest:       contest,
		savedStation:         station,
		savedContest:         contest,
	}
}

type Settings struct {
	writer                Writer
	view                  View
	defaultsOpener        DefaultsOpener
	xchangeRegexpMatcher  XchangeRegexpMatcher
	xchangeMultiTestValue string

	listeners []interface{}

	station core.Station
	contest core.Contest

	defaultStation core.Station
	defaultContest core.Contest

	savedStation core.Station
	savedContest core.Contest
}

func (s *Settings) SetWriter(writer Writer) {
	if writer == nil {
		s.writer = new(nullWriter)
		return
	}
	s.writer = writer
}

func (s *Settings) SetView(view View) {
	if view == nil {
		s.view = new(nullView)
		return
	}
	s.view = view
	s.showSettings()
}

func (s *Settings) Notify(listener interface{}) {
	s.listeners = append(s.listeners, listener)
}

func (s *Settings) Station() core.Station {
	return s.station
}

func (s *Settings) SetStation(station core.Station) {
	s.station = station
	s.savedStation = station
	s.emitStationChanged()
}

func (s *Settings) StationDirty() bool {
	return s.savedStation != s.station
}

func (s *Settings) emitStationChanged() {
	for _, listener := range s.listeners {
		if stationListener, ok := listener.(StationListener); ok {
			stationListener.StationChanged(s.station)
		}
	}
}

func (s *Settings) Contest() core.Contest {
	return s.contest
}

func (s *Settings) SetContest(contest core.Contest) {
	s.contest = contest
	s.savedContest = contest
	s.emitContestChanged()
}

func (s *Settings) ContestDirty() bool {
	return fmt.Sprintf("%v", s.savedContest) != fmt.Sprintf("%v", s.contest)
}

func (s *Settings) emitContestChanged() {
	for _, listener := range s.listeners {
		if contestListener, ok := listener.(ContestListener); ok {
			contestListener.ContestChanged(s.contest)
		}
	}
}

func (s *Settings) emitSettingsChanged() {
	for _, listener := range s.listeners {
		if SettingsListener, ok := listener.(SettingsListener); ok {
			SettingsListener.SettingsChanged(s)
		}
	}
}

func (s *Settings) Show() {
	s.view.Show()
	s.showSettings()
	s.view.HideMessage()
}

func (s *Settings) showSettings() {
	// station
	s.view.SetStationCallsign(s.station.Callsign.String())
	s.view.SetStationOperator(s.station.Operator.String())
	s.view.SetStationLocator(s.station.Locator.String())

	// contest
	s.view.SetContestName(s.contest.Name)
	s.view.SetContestEnterTheirNumber(s.contest.EnterTheirNumber)
	s.view.SetContestEnterTheirXchange(s.contest.EnterTheirXchange)
	s.view.SetContestRequireTheirXchange(s.contest.RequireTheirXchange)
	s.view.SetContestAllowMultiBand(s.contest.AllowMultiBand)
	s.view.SetContestAllowMultiMode(s.contest.AllowMultiMode)
	s.view.SetContestSameCountryPoints(strconv.Itoa(s.contest.SameCountryPoints))
	s.view.SetContestSameContinentPoints(strconv.Itoa(s.contest.SameContinentPoints))
	s.view.SetContestSpecificCountryPoints(strconv.Itoa(s.contest.SpecificCountryPoints))
	s.view.SetContestSpecificCountryPrefixes(strings.Join(s.contest.SpecificCountryPrefixes, ","))
	s.view.SetContestOtherPoints(strconv.Itoa(s.contest.OtherPoints))
	s.view.SetContestMultis(s.contest.Multis.DXCC, s.contest.Multis.WPX, s.contest.Multis.Xchange)
	s.view.SetContestXchangeMultiPattern(s.contest.XchangeMultiPattern)
	s.view.SetContestCountPerBand(s.contest.CountPerBand)
	s.view.SetContestCallHistoryFile(s.contest.CallHistoryFilename)
	s.view.SetContestCallHistoryField(s.contest.CallHistoryField)
	s.view.SetContestCabrilloQSOTemplate(s.contest.CabrilloQSOTemplate)
	s.updateXchangeMultiPatternResult()
}

func (s *Settings) Save() {
	modified := false
	if s.StationDirty() {
		modified = true
		s.emitStationChanged()
		s.savedStation = s.station
		s.writer.WriteStation(s.savedStation)
	}
	if s.ContestDirty() {
		modified = true
		s.emitContestChanged()
		s.savedContest = s.contest
		s.writer.WriteContest(s.savedContest)
	}
	if modified {
		s.emitSettingsChanged()
	}
}

func (s *Settings) Reset() {
	s.station = s.defaultStation
	s.contest = s.defaultContest

	s.showSettings()
	s.emitStationChanged()
	s.emitContestChanged()
}

func (s *Settings) OpenDefaults() {
	if s.defaultsOpener == nil {
		return
	}
	s.defaultsOpener()
}

func (s *Settings) EnterStationCallsign(value string) {
	cs, err := callsign.Parse(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Callsign = cs
}

func (s *Settings) EnterStationOperator(value string) {
	var cs callsign.Callsign
	var err error
	if value != "" {
		cs, err = callsign.Parse(value)
	}
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Operator = cs
}

func (s *Settings) EnterStationLocator(value string) {
	var loc locator.Locator
	var err error
	if value != "" {
		loc, err = locator.Parse(value)
	}
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Locator = loc
}

func (s *Settings) EnterContestName(value string) {
	s.contest.Name = value
}

func (s *Settings) EnterContestEnterTheirNumber(value bool) {
	s.contest.EnterTheirNumber = value
}

func (s *Settings) EnterContestEnterTheirXchange(value bool) {
	s.contest.EnterTheirXchange = value
}

func (s *Settings) EnterContestRequireTheirXchange(value bool) {
	s.contest.RequireTheirXchange = value
}

func (s *Settings) EnterContestAllowMultiBand(value bool) {
	s.contest.AllowMultiBand = value
}

func (s *Settings) EnterContestAllowMultiMode(value bool) {
	s.contest.AllowMultiMode = value
}

func (s *Settings) EnterContestSameCountryPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SameCountryPoints = points
}

func (s *Settings) EnterContestSameContinentPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SameContinentPoints = points
}

func (s *Settings) EnterContestSpecificCountryPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SpecificCountryPoints = points
}

func (s *Settings) EnterContestSpecificCountryPrefixes(value string) {
	rawPrefixes := strings.Split(value, ",")
	prefixes := make([]string, 0, len(rawPrefixes))
	for _, rawPrefix := range rawPrefixes {
		prefix := strings.TrimSpace(strings.ToUpper(rawPrefix))
		if prefix != "" {
			prefixes = append(prefixes, prefix)
		}
	}
	s.contest.SpecificCountryPrefixes = prefixes
}

func (s *Settings) EnterContestOtherPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.OtherPoints = points
}

func (s *Settings) EnterContestMultis(dxcc, wpx, xchange bool) {
	s.contest.Multis.DXCC = dxcc
	s.contest.Multis.WPX = wpx
	s.contest.Multis.Xchange = xchange
}

func (s *Settings) EnterContestXchangeMultiPattern(value string) {
	_, err := regexp.Compile(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.XchangeMultiPattern = value
	s.updateXchangeMultiPatternResult()
}

func (s *Settings) EnterContestTestXchangeValue(value string) {
	s.xchangeMultiTestValue = value
	s.updateXchangeMultiPatternResult()
}

func (s *Settings) updateXchangeMultiPatternResult() {
	var exp *regexp.Regexp
	var err error
	if s.contest.XchangeMultiPattern == "" {
		exp = nil
	} else {
		exp, err = regexp.Compile(s.contest.XchangeMultiPattern)
	}
	if err != nil {
		s.view.SetContestXchangeMultiPatternResult("(invalid)")
		return
	}

	multi, matched := s.xchangeRegexpMatcher(exp, s.xchangeMultiTestValue)
	if !matched {
		s.view.SetContestXchangeMultiPatternResult("(no match)")
		return
	}
	s.view.SetContestXchangeMultiPatternResult(multi)
}

func (s *Settings) EnterContestCountPerBand(value bool) {
	s.contest.CountPerBand = value
}

func (s *Settings) EnterContestCallHistoryFile(value string) {
	s.contest.CallHistoryFilename = value
}

func (s *Settings) EnterContestCallHistoryField(value string) {
	s.contest.CallHistoryField = value
}

func (s *Settings) EnterContestCabrilloQSOTemplate(value string) {
	_, err := template.New("").Parse(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.CabrilloQSOTemplate = value
}

type nullWriter struct{}

func (w *nullWriter) WriteStation(core.Station) error { return nil }
func (w *nullWriter) WriteContest(core.Contest) error { return nil }

type nullView struct{}

func (v *nullView) Show()                                      {}
func (v *nullView) ShowMessage(string)                         {}
func (v *nullView) HideMessage()                               {}
func (v *nullView) SetStationCallsign(string)                  {}
func (v *nullView) SetStationOperator(string)                  {}
func (v *nullView) SetStationLocator(string)                   {}
func (v *nullView) SetContestName(string)                      {}
func (v *nullView) SetContestEnterTheirNumber(bool)            {}
func (v *nullView) SetContestEnterTheirXchange(bool)           {}
func (v *nullView) SetContestRequireTheirXchange(bool)         {}
func (v *nullView) SetContestAllowMultiBand(bool)              {}
func (v *nullView) SetContestAllowMultiMode(bool)              {}
func (v *nullView) SetContestSameCountryPoints(string)         {}
func (v *nullView) SetContestSameContinentPoints(string)       {}
func (v *nullView) SetContestSpecificCountryPoints(string)     {}
func (v *nullView) SetContestSpecificCountryPrefixes(string)   {}
func (v *nullView) SetContestOtherPoints(string)               {}
func (v *nullView) SetContestMultis(dxcc, wpx, xchange bool)   {}
func (v *nullView) SetContestXchangeMultiPattern(string)       {}
func (v *nullView) SetContestXchangeMultiPatternResult(string) {}
func (v *nullView) SetContestCountPerBand(bool)                {}
func (v *nullView) SetContestCallHistoryFile(string)           {}
func (v *nullView) SetContestCallHistoryField(string)          {}
func (v *nullView) SetContestCabrilloQSOTemplate(string)       {}
