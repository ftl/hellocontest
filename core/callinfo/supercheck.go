package callinfo

import (
	"regexp"
	"sort"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type Supercheck struct {
	dxcc      DXCCFinder
	callsigns CallsignFinder
	history   CallHistoryFinder
	dupes     DupeChecker
	valuer    Valuer

	theirExchangeFields []core.ExchangeField
}

func NewSupercheck(dxcc DXCCFinder, callsigns CallsignFinder, history CallHistoryFinder,
	dupes DupeChecker, valuer Valuer) *Supercheck {

	return &Supercheck{
		dxcc:      dxcc,
		callsigns: callsigns,
		history:   history,
		dupes:     dupes,
		valuer:    valuer,
	}
}

func (s *Supercheck) SetTheirExchangeFields(fields []core.ExchangeField) {
	s.theirExchangeFields = fields
}

func (s *Supercheck) Calculate(input string, band core.Band, mode core.Mode) []core.AnnotatedCallsign {
	normalizedInput := normalizeInput(input)

	annotatedCallsigns := s.findMatchingCallsigns(normalizedInput)
	if len(annotatedCallsigns) == 0 {
		return nil
	}

	filter := placeholderToFilter(normalizedInput)

	result := make([]core.AnnotatedCallsign, 0, len(annotatedCallsigns))
	for _, annotatedCallsign := range annotatedCallsigns {
		if filter != nil && !filter.MatchString(annotatedCallsign.Callsign.String()) {
			continue
		}

		s.addInfos(&annotatedCallsign, normalizedInput, band, mode)
		result = append(result, annotatedCallsign)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LessThan(result[j])
	})

	return result
}

func (s *Supercheck) findMatchingCallsigns(input string) map[callsign.Callsign]core.AnnotatedCallsign {
	// TODO: include the worked QSOs in the search?
	if s.callsigns == nil || s.history == nil {
		return map[callsign.Callsign]core.AnnotatedCallsign{}
	}

	normalizedInput := normalizeInput(input)
	scpMatches, err := s.callsigns.Find(normalizedInput)
	if err != nil {
		scpMatches = []core.AnnotatedCallsign{}
	}
	historicMatches, err := s.history.Find(normalizedInput)
	if err != nil {
		historicMatches = []core.AnnotatedCallsign{}
	}

	result := make(map[callsign.Callsign]core.AnnotatedCallsign, len(scpMatches)+len(historicMatches))
	for _, match := range scpMatches {
		result[match.Callsign] = match
	}
	for _, match := range historicMatches {
		var annotatedCallsign core.AnnotatedCallsign
		storedCallsign, found := result[match.Callsign]
		if found {
			annotatedCallsign = storedCallsign
		} else {
			annotatedCallsign = match
		}
		annotatedCallsign.PredictedExchange = match.PredictedExchange
		result[annotatedCallsign.Callsign] = annotatedCallsign
	}

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

func (s *Supercheck) addInfos(annotatedCallsign *core.AnnotatedCallsign, normalizedInput string, band core.Band, mode core.Mode) {
	// TODO: this should be merged with the Collector to remove duplicate code
	matchString := annotatedCallsign.Callsign.String()
	exactMatch := (matchString == normalizedInput)

	dxccEntity, entityFound := s.dxcc.Find(matchString)
	if !entityFound {
		dxccEntity = dxcc.Prefix{}
	}

	workedQSOs, duplicate := s.dupes.FindWorkedQSOs(annotatedCallsign.Callsign, band, mode)
	predictedExchange := predictExchange(s.theirExchangeFields, dxccEntity, workedQSOs, nil, annotatedCallsign.PredictedExchange)

	var points, multis int
	if entityFound {
		points, multis, _ = s.valuer.Value(annotatedCallsign.Callsign, dxccEntity, band, mode, predictedExchange)
	}

	annotatedCallsign.ExactMatch = exactMatch
	annotatedCallsign.Duplicate = duplicate
	annotatedCallsign.Worked = len(workedQSOs) > 0
	annotatedCallsign.PredictedExchange = predictedExchange
	annotatedCallsign.Points = points
	annotatedCallsign.Multis = multis
}
