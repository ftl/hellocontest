package settings

import (
	"fmt"
	"log"

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

	station core.Station
	keyer   core.Keyer
	contest core.Contest

	defaultStation core.Station
	defaultKeyer   core.Keyer
	defaultContest core.Contest
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
}

func (s *Settings) Save() {
	log.Println("save the modified settings")
}

func (s *Settings) Reset() {
	log.Println("reset the settings to default values")
	s.station = s.defaultStation
	s.contest = s.defaultContest

	s.showSettings()
	s.emitStationChanged(s.station)
	s.emitContestChanged(s.contest)
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

type nullView struct{}

func (v *nullView) Show()                              {}
func (v *nullView) ShowMessage(string)                 {}
func (v *nullView) HideMessage()                       {}
func (v *nullView) SetStationCallsign(string)          {}
func (v *nullView) SetStationOperator(string)          {}
func (v *nullView) SetStationLocator(string)           {}
func (v *nullView) SetContestName(string)              {}
func (v *nullView) SetContestEnterTheirNumber(bool)    {}
func (v *nullView) SetContestEnterTheirXchange(bool)   {}
func (v *nullView) SetContestRequireTheirXchange(bool) {}
