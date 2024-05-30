package callinfo

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
)

func New(entities DXCCFinder, callsigns CallsignFinder, callHistory CallHistoryFinder, dupeChecker DupeChecker, valuer Valuer, exchangeFilter ExchangeFilter, asyncRunner core.AsyncRunner) *Callinfo {
	result := &Callinfo{
		view:           new(nullView),
		asyncRunner:    asyncRunner,
		entities:       entities,
		callsigns:      callsigns,
		callHistory:    callHistory,
		dupeChecker:    dupeChecker,
		valuer:         valuer,
		exchangeFilter: exchangeFilter,
		totalScore:     core.BandScore{Points: 1, Multis: 1},
	}

	return result
}

type Callinfo struct {
	view        View
	asyncRunner core.AsyncRunner

	entities       DXCCFinder
	callsigns      CallsignFinder
	callHistory    CallHistoryFinder
	dupeChecker    DupeChecker
	valuer         Valuer
	exchangeFilter ExchangeFilter

	lastCallsign        string
	lastBand            core.Band
	lastMode            core.Mode
	lastExchange        []string
	predictedExchange   []string
	theirExchangeFields []core.ExchangeField
	totalScore          core.BandScore

	matchOnFrequency          core.AnnotatedCallsign
	matchOnFrequencyAvailable bool
	bestMatch                 core.AnnotatedCallsign
	bestMatchAvailable        bool
	bestMatches               []string
}

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// CallsignFinder returns a list of matching callsigns for the given partial string.
type CallsignFinder interface {
	FindStrings(string) ([]string, error)
	Find(string) ([]core.AnnotatedCallsign, error)
}

// CallHistoryFinder returns additional information for a given callsign if a call history file is used.
type CallHistoryFinder interface {
	FindEntry(string) (core.AnnotatedCallsign, bool)
	Find(string) ([]core.AnnotatedCallsign, error)
}

// DupeChecker can be used to find out if the given callsign was already worked, according to the contest rules.
type DupeChecker interface {
	FindWorkedQSOs(callsign.Callsign, core.Band, core.Mode) ([]core.QSO, bool)
}

// Valuer provides the points and multis of a QSO based on the given information.
type Valuer interface {
	Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string)
}

// ExchangeFilter clears the exchange values that cannot be predicted (RST, serial).
type ExchangeFilter interface {
	FilterExchange([]string) []string
}

// View defines the visual part of the call information window.
type View interface {
	SetBestMatchingCallsign(callsign core.AnnotatedCallsign)
	SetDXCC(string, string, int, int, bool)
	SetValue(points, multis, value int)
	SetPredictedExchange(index int, text string)
	SetPredictedExchangeFields(fields []core.ExchangeField)
	SetUserInfo(string)
	SetSupercheck(callsigns []core.AnnotatedCallsign)
}

func (c *Callinfo) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
	c.view.SetPredictedExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) Refresh() {
	c.ShowInfo(c.lastCallsign, c.lastBand, c.lastMode, c.lastExchange)
}

func (c *Callinfo) ContestChanged(contest core.Contest) {
	if contest.Definition == nil {
		log.Printf("there is no contest definition!")
		return
	}
	c.theirExchangeFields = contest.TheirExchangeFields
	c.view.SetPredictedExchangeFields(c.theirExchangeFields)
}

func (c *Callinfo) ScoreUpdated(score core.Score) {
	c.totalScore = score.Result()
}

func (c *Callinfo) EntryOnFrequency(entry core.BandmapEntry, available bool) {
	c.asyncRunner(func() {
		c.matchOnFrequencyAvailable = available

		if available && c.matchOnFrequency.Callsign.String() == entry.Call.String() {
			// go on
		} else if available {
			normalizedCall := strings.TrimSpace(strings.ToUpper(c.lastCallsign))
			exactMatch := normalizedCall == entry.Call.String()
			c.matchOnFrequency = core.AnnotatedCallsign{
				Callsign:          entry.Call,
				Assembly:          core.MatchingAssembly{{OP: core.Matching, Value: entry.Call.String()}},
				Duplicate:         entry.Info.Duplicate,
				Worked:            entry.Info.Worked,
				ExactMatch:        exactMatch,
				Points:            entry.Info.Points,
				Multis:            entry.Info.Multis,
				PredictedExchange: entry.Info.PredictedExchange,
				OnFrequency:       true,
			}
		} else {
			c.matchOnFrequency = core.AnnotatedCallsign{}
		}

		c.showBestMatch()
	})
}

func (c *Callinfo) BestMatches() []string {
	return c.bestMatches
}

func (c *Callinfo) BestMatch() string {
	bestMatch, available := c.findBestMatch()
	if !available {
		return ""
	}
	return bestMatch.Callsign.String()
}

func (c *Callinfo) PredictedExchange() []string {
	return c.predictedExchange
}

func (c *Callinfo) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, exchange []string) core.Callinfo {
	result := core.Callinfo{
		Call: call,
	}

	entity, found := c.findDXCCEntity(call.String())
	if found {
		result.DXCCName = entity.Name
		result.PrimaryPrefix = entity.PrimaryPrefix
		result.Continent = entity.Continent
		result.ITUZone = int(entity.ITUZone)
		result.CQZone = int(entity.CQZone)
	}

	entry, found := c.callHistory.FindEntry(call.String())
	var historicExchange []string
	if found {
		historicExchange = entry.PredictedExchange
		result.UserText = joinAvailableValues(entry.Name, entry.UserText)
	}

	qsos, duplicate := c.dupeChecker.FindWorkedQSOs(call, band, mode)
	result.Duplicate = duplicate
	result.Worked = len(qsos) > 0
	result.PredictedExchange = c.predictExchange(entity, qsos, exchange, historicExchange)
	result.FilteredExchange = c.exchangeFilter.FilterExchange(result.PredictedExchange)
	result.ExchangeText = strings.Join(result.FilteredExchange, " ")

	result.Points, result.Multis, result.MultiValues = c.valuer.Value(call, entity, band, mode, result.PredictedExchange)
	result.Value = (result.Points * c.totalScore.Multis) + (result.Multis * c.totalScore.Points) + (result.Points * result.Multis)

	return result
}

func (c *Callinfo) ShowInfo(call string, band core.Band, mode core.Mode, exchange []string) {
	c.lastCallsign = call
	c.lastBand = band
	c.lastMode = mode
	c.lastExchange = exchange

	entity, _ := c.findDXCCEntity(call)

	var callinfo core.Callinfo
	parsedCallsign, err := callsign.Parse(call)
	if err == nil {
		callinfo = c.GetInfo(parsedCallsign, band, mode, exchange)
		c.predictedExchange = callinfo.PredictedExchange
	} else {
		c.predictedExchange = exchange
	}

	supercheck := c.calculateSupercheck(call)
	c.bestMatches = make([]string, 0, len(supercheck))
	c.bestMatch = core.AnnotatedCallsign{}
	c.bestMatchAvailable = false
	for i, match := range supercheck {
		c.bestMatches = append(c.bestMatches, match.Callsign.String())
		if i == 0 {
			c.bestMatch = match
			c.bestMatchAvailable = true
		}
	}

	c.showDXCCEntity(entity)
	c.showBestMatch()
	c.view.SetUserInfo(callinfo.UserText)
	c.view.SetValue(callinfo.Points, callinfo.Multis, callinfo.Value)
	for i := range c.theirExchangeFields {
		text := ""
		if i < len(callinfo.FilteredExchange) {
			text = callinfo.FilteredExchange[i]
		}
		c.view.SetPredictedExchange(i+1, text)
	}
	c.view.SetSupercheck(supercheck)
}

func joinAvailableValues(values ...string) string {
	availableValues := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			availableValues = append(availableValues, value)
		}
	}
	return strings.Join(availableValues, ", ")
}

func (c *Callinfo) GetValue(call callsign.Callsign, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string) {
	entity, found := c.findDXCCEntity(call.String())
	if !found {
		return 0, 0, nil
	}
	callinfo := c.GetInfo(call, band, mode, exchange)

	return c.valuer.Value(call, entity, band, mode, callinfo.PredictedExchange)
}

func (c *Callinfo) findDXCCEntity(call string) (dxcc.Prefix, bool) {
	if c.entities == nil {
		return dxcc.Prefix{}, false
	}
	return c.entities.Find(call)
}

func (c *Callinfo) showDXCCEntity(entity dxcc.Prefix) {
	var dxccName string
	if entity.PrimaryPrefix != "" {
		dxccName = fmt.Sprintf("%s (%s)", entity.Name, entity.PrimaryPrefix)
	}
	c.view.SetDXCC(dxccName, entity.Continent, int(entity.ITUZone), int(entity.CQZone), !entity.NotARRLCompliant)
}

func (c *Callinfo) findBestMatch() (core.AnnotatedCallsign, bool) {
	match := c.matchOnFrequency

	if c.bestMatchAvailable {
		match = c.bestMatch
	}

	return match, (match.Callsign.String() != "")
}

func (c *Callinfo) showBestMatch() {
	bestMatch, _ := c.findBestMatch()
	c.view.SetBestMatchingCallsign(bestMatch)
}

func (c *Callinfo) calculateSupercheck(s string) []core.AnnotatedCallsign {
	normalizedInput := strings.TrimSpace(strings.ToUpper(s))
	scpMatches, err := c.callsigns.Find(s)
	if err != nil {
		log.Printf("Callsign search for failed for %s: %v", s, err)
		return nil
	}
	historicMatches, _ := c.callHistory.Find(s)

	annotatedCallsigns := make(map[callsign.Callsign]core.AnnotatedCallsign, len(scpMatches)+len(historicMatches))
	for _, match := range scpMatches {
		annotatedCallsigns[match.Callsign] = match
	}
	for _, match := range historicMatches {
		var annotatedCallsign core.AnnotatedCallsign
		storedCallsign, found := annotatedCallsigns[match.Callsign]
		if found {
			annotatedCallsign = storedCallsign
		} else {
			annotatedCallsign = match
		}
		annotatedCallsign.PredictedExchange = match.PredictedExchange
		annotatedCallsigns[annotatedCallsign.Callsign] = annotatedCallsign
	}

	filter := placeholderToFilter(normalizedInput)

	result := make([]core.AnnotatedCallsign, 0, len(annotatedCallsigns))
	for _, annotatedCallsign := range annotatedCallsigns {
		matchString := annotatedCallsign.Callsign.String()
		exactMatch := (matchString == normalizedInput)
		if filter != nil && !filter.MatchString(matchString) {
			continue
		}
		entity, _ := c.findDXCCEntity(matchString)

		qsos, duplicate := c.dupeChecker.FindWorkedQSOs(annotatedCallsign.Callsign, c.lastBand, c.lastMode)
		predictedExchange := c.predictExchange(entity, qsos, nil, annotatedCallsign.PredictedExchange)

		entity, entityFound := c.entities.Find(matchString)

		var points, multis int
		if entityFound {
			points, multis, _ = c.valuer.Value(annotatedCallsign.Callsign, entity, c.lastBand, c.lastMode, predictedExchange)
		}

		annotatedCallsign.Duplicate = duplicate
		annotatedCallsign.Worked = len(qsos) > 0
		annotatedCallsign.ExactMatch = exactMatch
		annotatedCallsign.Points = points
		annotatedCallsign.Multis = multis
		annotatedCallsign.PredictedExchange = predictedExchange

		result = append(result, annotatedCallsign)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LessThan(result[j])
	})

	return result
}

func placeholderToFilter(s string) *regexp.Regexp {
	if !strings.Contains(s, core.FilterPlaceholder) {
		return nil
	}

	parts := strings.Split(s, core.FilterPlaceholder)
	for i := range parts {
		parts[i] = regexp.QuoteMeta(parts[i])
	}
	return regexp.MustCompile(strings.Join(parts, "."))
}

func (c *Callinfo) predictExchange(entity dxcc.Prefix, qsos []core.QSO, currentExchange []string, historicExchange []string) []string {
	result := make([]string, len(c.theirExchangeFields))
	copy(result, currentExchange)

	if entity.PrimaryPrefix != "" {
		for i, field := range c.theirExchangeFields {
			if i < len(historicExchange) && historicExchange[i] != "" {
				continue
			}
			switch {
			case field.Properties.Contains(conval.CQZoneProperty):
				result[i] = strconv.Itoa(int(entity.CQZone))
			case field.Properties.Contains(conval.ITUZoneProperty):
				result[i] = strconv.Itoa(int(entity.ITUZone))
			case field.Properties.Contains(conval.DXCCEntityProperty), field.Properties.Contains(conval.DXCCPrefixProperty):
				result[i] = entity.PrimaryPrefix
			}
		}
	}

	for i := range result {
		foundInQSO := false
		for _, qso := range qsos {
			if i >= len(qso.TheirExchange) {
				break
			}

			if result[i] == "" {
				result[i] = qso.TheirExchange[i]
				foundInQSO = true
			} else if result[i] != qso.TheirExchange[i] {
				result[i] = ""
				foundInQSO = false
				break
			}
		}
		if !foundInQSO && i < len(historicExchange) && historicExchange[i] != "" {
			result[i] = historicExchange[i]
		}
	}

	return result
}

type nullView struct{}

func (v *nullView) Show()                                                   {}
func (v *nullView) Hide()                                                   {}
func (v *nullView) SetBestMatchingCallsign(callsign core.AnnotatedCallsign) {}
func (v *nullView) SetDXCC(string, string, int, int, bool)                  {}
func (v *nullView) SetValue(int, int, int)                                  {}
func (v *nullView) SetPredictedExchange(int, string)                        {}
func (v *nullView) SetPredictedExchangeFields(fields []core.ExchangeField)  {}
func (v *nullView) SetUserInfo(string)                                      {}
func (v *nullView) SetSupercheck(callsigns []core.AnnotatedCallsign)        {}
