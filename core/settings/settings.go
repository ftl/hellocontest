package settings

import (
	"fmt"
	"log"

	"github.com/ftl/hamradio/callsign"
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

type View interface {
	Show()
	ShowMessage(string)
	HideMessage()

	SetStationCallsign(string)
	SetContestEnterTheirNumber(bool)
}

func New(station core.Station, contest core.Contest) *Settings {
	return &Settings{
		station:        station,
		contest:        contest,
		defaultStation: station,
		defaultContest: contest,
	}
}

type Settings struct {
	view View

	listeners []interface{}

	station core.Station
	contest core.Contest

	defaultStation core.Station
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

func (s *Settings) emitStationChanged(station core.Station) {
	for _, listener := range s.listeners {
		if stationListener, ok := listener.(StationListener); ok {
			stationListener.StationChanged(station)
		}
	}
}

func (s *Settings) emitContestChanged(contest core.Contest) {
	for _, listener := range s.listeners {
		if contestListener, ok := listener.(ContestListener); ok {
			contestListener.ContestChanged(contest)
		}
	}
}

func (s *Settings) Station() core.Station {
	return s.station
}

func (s *Settings) Contest() core.Contest {
	return s.contest
}

func (s *Settings) Show() {
	s.view.Show()
	s.showSettings()
	s.view.HideMessage()
}

func (s *Settings) showSettings() {
	// station
	s.view.SetStationCallsign(s.station.Callsign.String())

	// contest
	s.view.SetContestEnterTheirNumber(s.contest.EnterTheirNumber)
}

func (s *Settings) Save() {
	log.Println("save the modified settings")
}

func (s *Settings) Reset() {
	log.Println("reset the settings to default values")
	s.station = s.defaultStation
	s.contest = s.defaultContest
	s.emitStationChanged(s.station)
	s.emitContestChanged(s.contest)
}

func (s *Settings) SetStation(station core.Station) {
	s.station = station
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

func (s *Settings) EnterContestEnterTheirNumber(value bool) {
	s.contest.EnterTheirNumber = value
	s.emitContestChanged(s.contest)
}

type nullView struct{}

func (v *nullView) Show()                           {}
func (v *nullView) ShowMessage(string)              {}
func (v *nullView) HideMessage()                    {}
func (v *nullView) SetStationCallsign(string)       {}
func (v *nullView) SetContestEnterTheirNumber(bool) {}
