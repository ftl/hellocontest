package settings

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/ftl/conval"
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

type BrowserOpener func(string)

type XchangeRegexpMatcher func(*regexp.Regexp, string) (string, bool)

type View interface {
	Show()
	ShowMessage(string)
	HideMessage()
	Ready() bool

	SetStationCallsign(string)
	SetStationOperator(string)
	SetStationLocator(string)
	SetContestIdentifiers(ids []string, texts []string)
	SetContestPagesAvailable(bool, bool)
	SelectContestIdentifier(string)
	SetContestExchangeFields([]core.ExchangeField)
	SetContestExchangeValue(index int, value string)
	SetContestGenerateSerialExchange(active bool, sensitive bool)

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
	SetContestCallHistoryFieldName(i int, value string)
	SetContestCabrilloQSOTemplate(string)
}

func New(defaultsOpener DefaultsOpener, browserOpener BrowserOpener, station core.Station, contest core.Contest) *Settings {
	return &Settings{
		writer:         new(nullWriter),
		view:           new(nullView),
		defaultsOpener: defaultsOpener,
		browserOpener:  browserOpener,
		station:        station,
		contest:        contest,
		defaultStation: station,
		defaultContest: contest,
		savedStation:   station,
		savedContest:   deepCopyContest(contest),
	}
}

type Settings struct {
	writer                   Writer
	view                     View
	defaultsOpener           DefaultsOpener
	browserOpener            BrowserOpener
	xchangeMultiTestValue    string
	serialExchangeFieldIndex int

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
	s.savedContest = deepCopyContest(contest)
	s.contest.UpdateExchangeFields()
	s.emitContestChanged()
}

func deepCopyContest(contest core.Contest) core.Contest {
	result := contest

	result.ExchangeValues = make([]string, len(contest.ExchangeValues))
	copy(result.ExchangeValues, contest.ExchangeValues)

	result.CallHistoryFieldNames = make([]string, len(contest.CallHistoryFieldNames))
	copy(result.CallHistoryFieldNames, contest.CallHistoryFieldNames)

	return result
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
	if !s.view.Ready() {
		return
	}

	// station
	s.view.SetStationCallsign(s.station.Callsign.String())
	s.view.SetStationOperator(s.station.Operator.String())
	s.view.SetStationLocator(s.station.Locator.String())

	// contest definition
	definitionNames, err := conval.IncludedDefinitionNames()
	if err != nil {
		log.Printf("Cannot get the included contest definitions: %v", err)
	} else {
		ids := make([]string, 1, len(definitionNames)+1)
		ids[0] = ""
		ids = append(ids, definitionNames...)
		texts := make([]string, len(ids))
		for i, id := range ids {
			definition, err := conval.IncludedDefinition(id)
			if err != nil {
				continue
			}
			ids[i] = strings.ToUpper(id)
			texts[i] = fmt.Sprintf("%s - %s", strings.ToUpper(id), definition.Name)
		}
		s.view.SetContestIdentifiers(ids, texts)
	}
	if s.contest.Definition != nil {
		s.view.SelectContestIdentifier(strings.ToUpper(string(s.contest.Definition.Identifier)))
	} else {
		s.view.SelectContestIdentifier("")
	}
	s.updateContestPages()
	s.updateExchangeFields()

	// contest (old - will be removed)
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
		s.contest.UpdateExchangeFields()
		s.emitContestChanged()
		s.savedContest = deepCopyContest(s.contest)
		s.writer.WriteContest(s.savedContest)
	}
	if modified {
		s.emitSettingsChanged()
	}
}

func (s *Settings) Reset() {
	s.station = s.defaultStation
	s.contest = deepCopyContest(s.defaultContest)
	s.contest.UpdateExchangeFields()

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

func (s *Settings) SelectContestIdentifier(value string) {
	var definition *conval.Definition
	var err error

	if value == "" {
		definition = nil
	} else {
		definition, err = conval.IncludedDefinition(value)
		if err != nil {
			log.Printf("Cannot find the selected contest definition %s: %v", value, err)
			definition = nil
		}
	}

	s.contest.Definition = definition
	s.updateContestPages()
	s.updateExchangeFields()
}

func (s *Settings) updateContestPages() {
	if s.contest.Definition == nil {
		s.view.SetContestPagesAvailable(false, false)
	} else {
		s.view.SetContestPagesAvailable(s.contest.Definition.OfficialRules != "", s.contest.Definition.UploadURL != "")
	}
}

func (s *Settings) updateExchangeFields() {
	var exchangeFields []core.ExchangeField
	if s.contest.Definition == nil {
		exchangeFields = nil
	} else {
		exchangeFields = core.DefinitionsToExchangeFields(s.contest.Definition.ExchangeFields(), core.MyExchangeField)
	}
	s.view.SetContestExchangeFields(exchangeFields)

	newLen := len(exchangeFields)
	s.contest.ExchangeValues = ensureLen(s.contest.ExchangeValues, newLen)
	s.contest.CallHistoryFieldNames = ensureLen(s.contest.CallHistoryFieldNames, newLen)

	exclusiveSerialField := false
	s.serialExchangeFieldIndex = -1
	for i, value := range s.contest.ExchangeValues {
		field := exchangeFields[i]
		if field.CanContainSerial {
			s.serialExchangeFieldIndex = i
		}
		if field.CanContainSerial && len(field.Properties) == 1 {
			s.contest.ExchangeValues[i] = ""
			value = ""
			s.contest.GenerateSerialExchange = true
			exclusiveSerialField = true
		}
		s.view.SetContestExchangeValue(i+1, value)
	}

	s.view.SetContestGenerateSerialExchange(s.contest.GenerateSerialExchange, !exclusiveSerialField)

	for i, value := range s.contest.CallHistoryFieldNames {
		s.view.SetContestCallHistoryFieldName(i, value)
	}
}

func ensureLen(a []string, l int) []string {
	if len(a) < l {
		return append(a, make([]string, l-len(a))...)
	}
	if len(a) > l {
		return a[:l]
	}
	return a
}

func (s *Settings) OpenContestRulesPage() {
	if s.contest.Definition == nil {
		return
	}
	url := s.contest.Definition.OfficialRules
	if url == "" {
		return
	}
	s.browserOpener(url)
}

func (s *Settings) OpenContestUploadPage() {
	if s.contest.Definition == nil {
		return
	}
	url := s.contest.Definition.UploadURL
	if url == "" {
		return
	}
	s.browserOpener(url)
}

func (s *Settings) EnterContestExchangeValue(field core.EntryField, value string) {
	i := field.ExchangeIndex() - 1
	if i < 0 {
		log.Printf("%s is not an exchange field!", field)
		return
	}
	if i >= len(s.contest.ExchangeValues) {
		log.Printf("%s is outside the exchange field array (%d)", field, len(s.contest.ExchangeValues))
		return
	}

	s.contest.ExchangeValues[i] = value
}

func (s *Settings) EnterContestGenerateSerialExchange(value bool) {
	s.contest.GenerateSerialExchange = value
	if s.serialExchangeFieldIndex >= 0 {
		s.view.SetContestExchangeValue(s.serialExchangeFieldIndex+1, "")
	}
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
	// TODO remove
}

func (s *Settings) EnterContestCountPerBand(value bool) {
	s.contest.CountPerBand = value
}

func (s *Settings) EnterContestCallHistoryFile(value string) {
	s.contest.CallHistoryFilename = value
}

func (s *Settings) EnterContestCallHistoryFieldName(field core.EntryField, value string) {
	i := field.ExchangeIndex() - 1
	if i < 0 || i >= len(s.contest.TheirExchangeFields) {
		log.Printf("call history field name is out of range: %d", i)
		return
	}

	s.contest.CallHistoryFieldNames[i] = value
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

func (v *nullView) Show()                                              {}
func (v *nullView) ShowMessage(string)                                 {}
func (v *nullView) HideMessage()                                       {}
func (v *nullView) Ready() bool                                        { return false }
func (v *nullView) SetStationCallsign(string)                          {}
func (v *nullView) SetStationOperator(string)                          {}
func (v *nullView) SetStationLocator(string)                           {}
func (v *nullView) SetContestIdentifiers(ids []string, texts []string) {}
func (v *nullView) SetContestPagesAvailable(bool, bool)                {}
func (v *nullView) SelectContestIdentifier(string)                     {}
func (v *nullView) SetContestExchangeFields([]core.ExchangeField)      {}
func (v *nullView) SetContestExchangeValue(index int, value string)    {}
func (v *nullView) SetContestGenerateSerialExchange(bool, bool)        {}
func (v *nullView) SetContestName(string)                              {}
func (v *nullView) SetContestEnterTheirNumber(bool)                    {}
func (v *nullView) SetContestEnterTheirXchange(bool)                   {}
func (v *nullView) SetContestRequireTheirXchange(bool)                 {}
func (v *nullView) SetContestAllowMultiBand(bool)                      {}
func (v *nullView) SetContestAllowMultiMode(bool)                      {}
func (v *nullView) SetContestSameCountryPoints(string)                 {}
func (v *nullView) SetContestSameContinentPoints(string)               {}
func (v *nullView) SetContestSpecificCountryPoints(string)             {}
func (v *nullView) SetContestSpecificCountryPrefixes(string)           {}
func (v *nullView) SetContestOtherPoints(string)                       {}
func (v *nullView) SetContestMultis(dxcc, wpx, xchange bool)           {}
func (v *nullView) SetContestXchangeMultiPattern(string)               {}
func (v *nullView) SetContestXchangeMultiPatternResult(string)         {}
func (v *nullView) SetContestCountPerBand(bool)                        {}
func (v *nullView) SetContestCallHistoryFile(string)                   {}
func (v *nullView) SetContestCallHistoryFieldName(int, string)         {}
func (v *nullView) SetContestCabrilloQSOTemplate(string)               {}
