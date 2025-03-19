package callinfo

import (
	"strconv"
	"strings"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type Collector struct {
	dxcc           DXCCFinder
	callsigns      CallsignFinder
	history        CallHistoryFinder
	dupes          DupeChecker
	valuer         Valuer
	exchangeFilter ExchangeFilter

	theirExchangeFields []core.ExchangeField
	totalScore          core.BandScore
}

func NewColletor(dxcc DXCCFinder, callsigns CallsignFinder, history CallHistoryFinder,
	dupes DupeChecker, valuer Valuer, exchangeFilter ExchangeFilter) *Collector {

	return &Collector{
		dxcc:           dxcc,
		callsigns:      callsigns,
		history:        history,
		dupes:          dupes,
		valuer:         valuer,
		exchangeFilter: exchangeFilter,
		totalScore:     core.BandScore{Points: 1, Multis: 1},
	}
}

func (c *Collector) SetTheirExchangeFields(fields []core.ExchangeField) {
	c.theirExchangeFields = fields
}

func (c *Collector) ScoreUpdated(score core.Score) {
	c.totalScore = score.Result()
}

func (c *Collector) GetInfoForInput(input string, band core.Band, mode core.Mode, exchange []string) core.Callinfo {
	result := core.Callinfo{
		Input: normalizeInput(input),
	}
	c.addCallsign(&result)

	c.addInfos(&result, band, mode, exchange)

	return result
}

func (c *Collector) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, exchange []string) core.Callinfo {
	result := core.Callinfo{
		Input: normalizeInput(call.String()),
		Call:  call,
	}
	result.CallValid = (result.Input != "")

	c.addInfos(&result, band, mode, exchange)

	return result
}

func normalizeInput(input string) string {
	return strings.TrimSpace(strings.ToUpper(input))
}

func (c *Collector) addCallsign(info *core.Callinfo) bool {
	call, err := callsign.Parse(info.Input)
	info.CallValid = (err == nil)
	if info.CallValid {
		info.Call = call
	} else {
		info.Call = callsign.Callsign{}
	}
	return info.CallValid
}

func (c *Collector) addInfos(info *core.Callinfo, band core.Band, mode core.Mode, exchange []string) {
	c.addDXCC(info)
	c.addHistoryData(info)
	if !info.CallValid {
		return
	}

	workedQSOs, duplicate := c.dupes.FindWorkedQSOs(info.Call, band, mode)
	c.addWorkedState(info, workedQSOs, duplicate)
	c.addPredictedExchange(info, workedQSOs, exchange)
	c.addValue(info, band, mode)
}

func (c *Collector) addDXCC(info *core.Callinfo) bool {
	if c.dxcc == nil {
		return false
	}

	entity, found := c.dxcc.Find(info.Input)
	if !found {
		return false
	}

	info.DXCCEntity = entity

	return true
}

func (c *Collector) addHistoryData(info *core.Callinfo) bool {
	entry, found := c.history.FindEntry(info.Input)
	if !found {
		info.PredictedExchange = []string{}
		return false
	}

	info.UserText = entry.UserText
	info.PredictedExchange = entry.PredictedExchange

	return true
}

func (c *Collector) addWorkedState(info *core.Callinfo, workedQSOs []core.QSO, duplicate bool) {
	info.Duplicate = duplicate
	info.Worked = len(workedQSOs) > 0
}

func (c *Collector) addPredictedExchange(info *core.Callinfo, workedQSOs []core.QSO, currentExchange []string) {
	info.PredictedExchange = c.predictExchange(info, workedQSOs, currentExchange)
	info.FilteredExchange = c.exchangeFilter.FilterExchange(info.PredictedExchange)
	info.ExchangeText = strings.Join(info.FilteredExchange, " ")
}

func (c *Collector) predictExchange(info *core.Callinfo, qsos []core.QSO, currentExchange []string) []string {
	result := make([]string, len(c.theirExchangeFields))
	copy(result, currentExchange)

	historicExchange := info.PredictedExchange
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

		if foundInQSO {
			continue
		}

		if i < len(historicExchange) && historicExchange[i] != "" {
			result[i] = historicExchange[i]
		} else if info.DXCCEntity.PrimaryPrefix != "" {
			if i >= len(c.theirExchangeFields) {
				continue
			}
			field := c.theirExchangeFields[i]
			switch {
			case field.Properties.Contains(conval.CQZoneProperty):
				result[i] = strconv.Itoa(int(info.DXCCEntity.CQZone))
			case field.Properties.Contains(conval.ITUZoneProperty):
				result[i] = strconv.Itoa(int(info.DXCCEntity.ITUZone))
			case field.Properties.Contains(conval.DXCCEntityProperty), field.Properties.Contains(conval.DXCCPrefixProperty):
				result[i] = info.DXCCEntity.PrimaryPrefix
			}
		}
	}

	return result
}

func (c *Collector) addValue(info *core.Callinfo, band core.Band, mode core.Mode) {
	info.Points, info.Multis, info.MultiValues = c.valuer.Value(info.Call, info.DXCCEntity, band, mode, info.PredictedExchange)
	info.Value = (info.Points * c.totalScore.Multis) + (info.Multis * c.totalScore.Points) + (info.Points * info.Multis)
}
