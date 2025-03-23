package callinfo

import (
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

func NewCollector(dxcc DXCCFinder, callsigns CallsignFinder, history CallHistoryFinder,
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

func (c *Collector) GetInfoForInput(input string, band core.Band, mode core.Mode, currentExchange []string) core.Callinfo {
	result := core.Callinfo{
		Input: normalizeInput(input),
	}
	c.addCallsign(&result)

	c.addInfos(&result, band, mode, currentExchange)

	return result
}

func (c *Collector) GetInfo(call callsign.Callsign, band core.Band, mode core.Mode, currentExchange []string) core.Callinfo {
	result := core.Callinfo{
		Input: normalizeInput(call.String()),
		Call:  call,
	}
	result.CallValid = (result.Input != "")

	c.addInfos(&result, band, mode, currentExchange)

	return result
}

func (c *Collector) GetValue(call callsign.Callsign, band core.Band, mode core.Mode) (points, multis int, multiValues map[conval.Property]string) {
	// TODO: replace GetValue with UpdateValue, which will not need the exchange prediction
	// remember to invalidate the cached Callinfo when a new QSO with the callsign is added
	if c.dxcc == nil || c.dupes == nil || c.valuer == nil {
		return 0, 0, nil
	}

	info := &core.Callinfo{
		Input: normalizeInput(call.String()),
		Call:  call,
	}
	info.CallValid = (info.Input != "")
	c.initializeCallinfo(info)

	dxccValid := c.addDXCC(info)
	if !dxccValid {
		return 0, 0, nil
	}

	c.addHistoryData(info)
	workedQSOs, _ := c.dupes.FindWorkedQSOs(info.Call, band, mode)
	predictExchange(c.theirExchangeFields, info.DXCCEntity, workedQSOs, []string{}, info.PredictedExchange)
	c.addValue(info, band, mode)

	return info.Points, info.Multis, info.MultiValues
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

func (c *Collector) addInfos(info *core.Callinfo, band core.Band, mode core.Mode, currentExchange []string) {
	c.initializeCallinfo(info)
	dxccValid := c.addDXCC(info)
	if !info.CallValid || !dxccValid {
		return
	}

	c.addHistoryData(info)
	if c.dupes == nil {
		return
	}
	workedQSOs, duplicate := c.dupes.FindWorkedQSOs(info.Call, band, mode)
	c.addWorkedState(info, workedQSOs, duplicate)
	// ATTENTION: temporal coupling! addPredictedExchange relies on addHistoryData putting
	// the historic exchange into the Callinfo.PredictedExchange field.
	c.addPredictedExchange(info, workedQSOs, currentExchange)
	c.addValue(info, band, mode)
}

func (c *Collector) initializeCallinfo(info *core.Callinfo) {
	info.PredictedExchange = make([]string, 0, len(c.theirExchangeFields))
	info.FilteredExchange = make([]string, 0, len(c.theirExchangeFields))
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
	info.PredictedExchange = predictExchange(c.theirExchangeFields, info.DXCCEntity, workedQSOs, currentExchange, info.PredictedExchange)
	if c.exchangeFilter != nil {
		info.FilteredExchange = c.exchangeFilter.FilterExchange(info.PredictedExchange)
	} else {
		info.FilteredExchange = info.PredictedExchange
	}
	info.ExchangeText = strings.Join(info.FilteredExchange, " ")
}

func (c *Collector) addValue(info *core.Callinfo, band core.Band, mode core.Mode) bool {
	if c.valuer == nil {
		return false
	}

	info.Points, info.Multis, info.MultiValues = c.valuer.Value(info.Call, info.DXCCEntity, band, mode, info.PredictedExchange)
	info.Value = (info.Points * c.totalScore.Multis) + (info.Multis * c.totalScore.Points) + (info.Points * info.Multis)

	return true
}
