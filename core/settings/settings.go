package settings

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hellocontest/core"
)

type KeyerListener interface {
	KeyerChanged(core.Keyer)
}

type KeyerListenerFunc func(core.Keyer)

func (f KeyerListenerFunc) KeyerChanged(keyer core.Keyer) {
	f(keyer)
}

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

type CabrilloListener interface {
	CabrilloChanged(core.Cabrillo)
}

type CabrilloListenerFunc func(core.Cabrillo)

func (f CabrilloListenerFunc) CabrilloChanged(cabrillo core.Cabrillo) {
	f(cabrillo)
}

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
	SetContestCountPerBand(bool)
}

func New(station core.Station, keyer core.Keyer, contest core.Contest) *Settings {
	return &Settings{
		station:        station,
		keyer:          keyer,
		contest:        contest,
		defaultStation: station,
		defaultKeyer:   keyer,
		defaultContest: contest,
	}
}

type Settings struct {
	view View

	listeners []interface{}

	station  core.Station
	keyer    core.Keyer
	contest  core.Contest
	cabrillo core.Cabrillo

	defaultStation  core.Station
	defaultKeyer    core.Keyer
	defaultContest  core.Contest
	defaultCabrillo core.Cabrillo
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
	s.emitStationChanged(station)
}

func (s *Settings) emitStationChanged(station core.Station) {
	for _, listener := range s.listeners {
		if stationListener, ok := listener.(StationListener); ok {
			stationListener.StationChanged(station)
		}
	}
}

func (s *Settings) Keyer() core.Keyer {
	return s.keyer
}

func (s *Settings) SetKeyer(keyer core.Keyer) {
	s.keyer = keyer
	s.emitKeyerChanged(keyer)
}

func (s *Settings) emitKeyerChanged(keyer core.Keyer) {
	for _, listener := range s.listeners {
		if keyerListener, ok := listener.(KeyerListener); ok {
			keyerListener.KeyerChanged(keyer)
		}
	}
}

func (s *Settings) Contest() core.Contest {
	return s.contest
}

func (s *Settings) SetContest(contest core.Contest) {
	s.contest = contest
	s.emitContestChanged(contest)
}

func (s *Settings) emitContestChanged(contest core.Contest) {
	for _, listener := range s.listeners {
		if contestListener, ok := listener.(ContestListener); ok {
			contestListener.ContestChanged(contest)
		}
	}
}

func (s *Settings) Cabrillo() core.Cabrillo {
	return s.cabrillo
}

func (s *Settings) SetCabrillo(cabrillo core.Cabrillo) {
	s.cabrillo = cabrillo
	s.emitCabrilloChanged(cabrillo)
}

func (s *Settings) emitCabrilloChanged(cabrillo core.Cabrillo) {
	for _, listener := range s.listeners {
		if CabrilloListener, ok := listener.(CabrilloListener); ok {
			CabrilloListener.CabrilloChanged(cabrillo)
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
}

func (s *Settings) Save() {
	log.Println("save the modified settings")
}

func (s *Settings) Reset() {
	s.station = s.defaultStation
	s.keyer = s.defaultKeyer
	s.contest = s.defaultContest
	s.cabrillo = s.defaultCabrillo

	s.showSettings()
	s.emitStationChanged(s.station)
	s.emitKeyerChanged(s.keyer)
	s.emitContestChanged(s.contest)
	s.emitCabrilloChanged(s.cabrillo)
}

func (s *Settings) EnterStationCallsign(value string) {
	cs, err := callsign.Parse(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Callsign = cs
	s.emitStationChanged(s.station)
}

func (s *Settings) EnterStationOperator(value string) {
	cs, err := callsign.Parse(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Operator = cs
	s.emitStationChanged(s.station)
}

func (s *Settings) EnterStationLocator(value string) {
	loc, err := locator.Parse(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.station.Locator = loc
	s.emitStationChanged(s.station)
}

func (s *Settings) EnterContestName(value string) {
	s.contest.Name = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestEnterTheirNumber(value bool) {
	s.contest.EnterTheirNumber = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestEnterTheirXchange(value bool) {
	s.contest.EnterTheirXchange = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestRequireTheirXchange(value bool) {
	s.contest.RequireTheirXchange = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestAllowMultiBand(value bool) {
	s.contest.AllowMultiBand = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestAllowMultiMode(value bool) {
	s.contest.AllowMultiMode = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestSameCountryPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SameCountryPoints = points
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestSameContinentPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SameContinentPoints = points
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestSpecificCountryPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.SpecificCountryPoints = points
	s.emitContestChanged(s.contest)
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
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestOtherPoints(value string) {
	points, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	s.view.HideMessage()
	s.contest.OtherPoints = points
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestMultis(dxcc, wpx, xchange bool) {
	s.contest.Multis.DXCC = dxcc
	s.contest.Multis.WPX = wpx
	s.contest.Multis.Xchange = xchange
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestXchangeMultiPattern(value string) {
	_, err := regexp.Compile(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
	}
	s.view.HideMessage()
	s.contest.XchangeMultiPattern = value
	s.emitContestChanged(s.contest)
}

func (s *Settings) EnterContestCountPerBand(value bool) {
	s.contest.CountPerBand = value
	s.emitContestChanged(s.contest)
}

type nullView struct{}

func (v *nullView) Show()                                    {}
func (v *nullView) ShowMessage(string)                       {}
func (v *nullView) HideMessage()                             {}
func (v *nullView) SetStationCallsign(string)                {}
func (v *nullView) SetStationOperator(string)                {}
func (v *nullView) SetStationLocator(string)                 {}
func (v *nullView) SetContestName(string)                    {}
func (v *nullView) SetContestEnterTheirNumber(bool)          {}
func (v *nullView) SetContestEnterTheirXchange(bool)         {}
func (v *nullView) SetContestRequireTheirXchange(bool)       {}
func (v *nullView) SetContestAllowMultiBand(bool)            {}
func (v *nullView) SetContestAllowMultiMode(bool)            {}
func (v *nullView) SetContestSameCountryPoints(string)       {}
func (v *nullView) SetContestSameContinentPoints(string)     {}
func (v *nullView) SetContestSpecificCountryPoints(string)   {}
func (v *nullView) SetContestSpecificCountryPrefixes(string) {}
func (v *nullView) SetContestOtherPoints(string)             {}
func (v *nullView) SetContestMultis(dxcc, wpx, xchange bool) {}
func (v *nullView) SetContestXchangeMultiPattern(string)     {}
func (v *nullView) SetContestCountPerBand(bool)              {}
