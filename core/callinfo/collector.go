package callinfo

import (
	"strconv"
	"strings"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
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
	c.initializeCallinfo(info)
	c.addDXCC(info)
	c.addHistoryData(info)
	if !info.CallValid {
		return
	}

	if c.dupes == nil {
		return
	}
	workedQSOs, duplicate := c.dupes.FindWorkedQSOs(info.Call, band, mode)
	c.addWorkedState(info, workedQSOs, duplicate)
	c.addPredictedExchange(info, workedQSOs, exchange)
	c.addValue(info, band, mode)
}

func (c *Collector) initializeCallinfo(info *core.Callinfo) {
	info.DXCCEntity = dxcc.Prefix{}
	info.UserText = ""
	info.PredictedExchange = make([]string, 0, len(c.theirExchangeFields))
	info.FilteredExchange = make([]string, 0, len(c.theirExchangeFields))
	info.ExchangeText = ""
	info.Duplicate = false
	info.Worked = false
	info.Points = 0
	info.Multis = 0
	info.MultiValues = make(map[conval.Property]string)
	info.Value = 0
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
	if c.history == nil {
		return false
	}

	entry, found := c.history.FindEntry(info.Input)
	if !found {
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
	info.PredictedExchange = c.predictExchange(info.DXCCEntity, workedQSOs, currentExchange, info.PredictedExchange)
	if c.exchangeFilter != nil {
		info.FilteredExchange = c.exchangeFilter.FilterExchange(info.PredictedExchange)
	} else {
		info.FilteredExchange = info.PredictedExchange
	}
	info.ExchangeText = strings.Join(info.FilteredExchange, " ")
}

func (c *Collector) predictExchange(dxcc dxcc.Prefix, qsos []core.QSO, currentExchange []string, historicExchange []string) []string {
	result := make([]string, len(c.theirExchangeFields))
	copy(result, currentExchange)

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
		} else if dxcc.PrimaryPrefix != "" {
			if i >= len(c.theirExchangeFields) {
				continue
			}
			field := c.theirExchangeFields[i]
			switch {
			case field.Properties.Contains(conval.CQZoneProperty):
				result[i] = strconv.Itoa(int(dxcc.CQZone))
			case field.Properties.Contains(conval.ITUZoneProperty):
				result[i] = strconv.Itoa(int(dxcc.ITUZone))
			case field.Properties.Contains(conval.DXCCEntityProperty), field.Properties.Contains(conval.DXCCPrefixProperty):
				result[i] = dxcc.PrimaryPrefix
			}
		}
	}

	return result
}

func (c *Collector) addValue(info *core.Callinfo, band core.Band, mode core.Mode) bool {
	if c.valuer == nil {
		return false
	}

	info.Points, info.Multis, info.MultiValues = c.valuer.Value(info.Call, info.DXCCEntity, band, mode, info.PredictedExchange)
	info.Value = (info.Points * c.totalScore.Multis) + (info.Multis * c.totalScore.Points) + (info.Points * info.Multis)

	return true
}
