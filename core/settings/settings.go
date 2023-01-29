package settings

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/callhistory"
)

const contestStartTimeFormat = "02-01-2006 15:04"

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
	SetContestGenerateReport(active bool, sensitive bool)

	SetContestName(string)
	SetContestStartTime(string)
	SetOperationModeSprint(bool)
	SetContestCallHistoryFile(string)
	SetContestCallHistoryFieldName(i int, value string)
	SetContestAvailableCallHistoryFieldNames([]string)

	SetQSOsGoal(string)
	SetPointsGoal(string)
	SetMultisGoal(string)
}

func New(defaultsOpener DefaultsOpener, browserOpener BrowserOpener, station core.Station, contest core.Contest) *Settings {
	result := &Settings{
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

	result.availableCallHistoryFieldNames = make([]string, 0, len(scp.DefaultFieldSet))
	for _, fieldName := range scp.DefaultFieldSet {
		result.availableCallHistoryFieldNames = append(result.availableCallHistoryFieldNames, string(fieldName))
	}

	return result
}

type Settings struct {
	writer                         Writer
	view                           View
	defaultsOpener                 DefaultsOpener
	browserOpener                  BrowserOpener
	reportFieldIndex               int
	serialExchangeFieldIndex       int
	availableCallHistoryFieldNames []string

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

	s.view.SetContestName(s.contest.Name)
	s.view.SetContestStartTime(s.formattedContestStartTime())
	s.view.SetOperationModeSprint(s.contest.OperationModeSprint)
	s.view.SetContestCallHistoryFile(s.contest.CallHistoryFilename)
	s.view.SetQSOsGoal(strconv.Itoa(s.contest.QSOsGoal))
	s.view.SetPointsGoal(strconv.Itoa(s.contest.PointsGoal))
	s.view.SetMultisGoal(strconv.Itoa(s.contest.MultisGoal))
}

func (s *Settings) formattedContestStartTime() string {
	if s.contest.StartTime.IsZero() {
		return ""
	}
	return s.contest.StartTime.Format(contestStartTimeFormat)
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
	s.contest.StartTime = time.Time{}
	s.contest.UpdateExchangeFields()

	log.Printf("GOALS: q %d p %d m %d", s.contest.QSOsGoal, s.contest.PointsGoal, s.contest.MultisGoal)

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

	if definition != nil {
		year := time.Now().Format("2006")
		s.contest.Name = fmt.Sprintf("%s %s", definition.Identifier, year)
		s.view.SetContestName(s.contest.Name)
	}
}

func (s *Settings) updateContestPages() {
	if s.contest.Definition == nil {
		s.view.SetContestPagesAvailable(false, false)
	} else {
		s.view.SetContestPagesAvailable(s.contest.Definition.OfficialRules != "", s.contest.Definition.UploadURL != "")
	}
}

func (s *Settings) updateExchangeFields() {
	s.view.SetContestAvailableCallHistoryFieldNames(s.availableCallHistoryFieldNames)

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
	s.reportFieldIndex = -1
	for i, value := range s.contest.ExchangeValues {
		field := exchangeFields[i]
		if field.CanContainSerial {
			s.serialExchangeFieldIndex = i
		}
		if field.CanContainReport {
			s.reportFieldIndex = i
		}
		if len(field.Properties) == 1 && field.CanContainSerial {
			value = ""
			s.contest.ExchangeValues[i] = value
			s.contest.GenerateSerialExchange = true
			exclusiveSerialField = true
		}
		if len(field.Properties) == 1 && field.CanContainReport {
			value = ""
			s.contest.ExchangeValues[i] = value
			s.contest.GenerateReport = true
		}
		s.contest.ExchangeValues[i] = value
		s.view.SetContestExchangeValue(i+1, value)
	}

	s.view.SetContestGenerateSerialExchange(s.contest.GenerateSerialExchange, !exclusiveSerialField)
	s.view.SetContestGenerateReport(s.contest.GenerateReport, true)

	for i, value := range s.contest.CallHistoryFieldNames {
		if value != "" {
			s.view.SetContestCallHistoryFieldName(i, value)
			continue
		}

		field := exchangeFields[i]
		var fieldName string
		switch {
		case field.Properties.Contains(conval.RSTProperty) || field.Properties.Contains(conval.SerialNumberProperty):
			fieldName = ""
		case len(field.Properties) == 1 && field.Properties.Contains(conval.NameProperty):
			fieldName = callhistory.NameField
		default:
			fieldName = callhistory.Exch1Field
		}
		s.view.SetContestCallHistoryFieldName(i, fieldName)
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

func (s *Settings) ClearCallHistory() {
	s.contest.CallHistoryFilename = ""
	s.contest.CallHistoryFieldNames = make([]string, len(s.contest.ExchangeValues))

	s.view.SetContestCallHistoryFile(s.contest.CallHistoryFilename)
	for i := range s.contest.CallHistoryFieldNames {
		s.view.SetContestCallHistoryFieldName(i, s.contest.CallHistoryFieldNames[i])
	}
}

func (s *Settings) EnterContestExchangeValue(field core.EntryField, value string) {
	i := field.ExchangeIndex() - 1
	if i < 0 {
		s.view.ShowMessage(fmt.Sprintf("%s is not an exchange field!", field))
		return
	}
	if i >= len(s.contest.ExchangeValues) {
		s.view.ShowMessage(fmt.Sprintf("%s is outside the exchange field array (%d)", field, len(s.contest.ExchangeValues)))
		return
	}
	s.view.HideMessage()
	s.contest.ExchangeValues[i] = value
}

func (s *Settings) EnterContestGenerateSerialExchange(value bool) {
	s.contest.GenerateSerialExchange = value
	if s.serialExchangeFieldIndex >= 0 {
		s.view.SetContestExchangeValue(s.serialExchangeFieldIndex+1, "")
	}
}

func (s *Settings) EnterContestGenerateReport(value bool) {
	s.contest.GenerateReport = value
	if s.reportFieldIndex < 0 {
		return
	}

	report := ""
	if !value && s.contest.Definition != nil {
		contestModes := s.contest.Definition.Modes
		if len(contestModes) == 1 {
			report = defaultReportForMode(contestModes[0])
		}
	}
	s.view.SetContestExchangeValue(s.reportFieldIndex+1, report)
}

func defaultReportForMode(mode conval.Mode) string {
	switch mode {
	case conval.ModeCW, conval.ModeRTTY, conval.ModeDigital:
		return "599"
	case conval.ModeSSB, conval.ModeFM:
		return "59"
	default:
		return ""
	}
}

func (s *Settings) EnterContestName(value string) {
	s.contest.Name = value
}

func (s *Settings) EnterContestStartTime(value string) {
	if value == "" {
		s.contest.StartTime = time.Time{}
		return
	}

	startTime, err := time.Parse(contestStartTimeFormat, value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%s is not a valid start time, use the format dd-mm-yyyy hh:mm.", value))
		return
	}
	s.view.HideMessage()

	s.contest.StartTime = startTime
}

func (s *Settings) SetContestStartTimeToday() {
	year, month, day := time.Now().Date()
	s.contest.StartTime = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	s.view.HideMessage()
	s.view.SetContestStartTime(s.formattedContestStartTime())
}

func (s *Settings) SetContestStartTimeNow() {
	s.contest.StartTime = time.Now().UTC().Truncate(time.Hour)
	s.view.HideMessage()
	s.view.SetContestStartTime(s.formattedContestStartTime())
}

func (s *Settings) SetOperationModeSprint(value bool) {
	s.contest.OperationModeSprint = value
}

func (s *Settings) EnterContestCallHistoryFile(value string) {
	s.contest.CallHistoryFilename = value
}

func (s *Settings) EnterContestCallHistoryFieldName(field core.EntryField, value string) {
	i := field.ExchangeIndex() - 1
	if i < 0 || i >= len(s.contest.CallHistoryFieldNames) {
		log.Printf("call history field name is out of range: %d", i)
		return
	}

	s.contest.CallHistoryFieldNames[i] = value
}

func (s *Settings) EnterQSOsGoal(value string) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	if intValue < 0 {
		intValue = 0
	}

	s.view.HideMessage()
	s.contest.QSOsGoal = intValue
}

func (s *Settings) EnterPointsGoal(value string) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	if intValue < 0 {
		intValue = 0
	}

	s.view.HideMessage()
	s.contest.PointsGoal = intValue
}

func (s *Settings) EnterMultisGoal(value string) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		s.view.ShowMessage(fmt.Sprintf("%v", err))
		return
	}
	if intValue < 0 {
		intValue = 0
	}

	s.view.HideMessage()
	s.contest.MultisGoal = intValue
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
func (v *nullView) SetContestGenerateReport(bool, bool)                {}
func (v *nullView) SetContestName(string)                              {}
func (v *nullView) SetContestStartTime(string)                         {}
func (v *nullView) SetOperationModeSprint(bool)                        {}
func (v *nullView) SetContestCallHistoryFile(string)                   {}
func (v *nullView) SetContestCallHistoryFieldName(int, string)         {}
func (v *nullView) SetContestAvailableCallHistoryFieldNames([]string)  {}
func (v *nullView) SetQSOsGoal(string)                                 {}
func (v *nullView) SetPointsGoal(string)                               {}
func (v *nullView) SetMultisGoal(string)                               {}
