package callinfo

import (
	"sync"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"

	"github.com/ftl/hellocontest/core"
)

type Collector struct {
	dxcc      DXCCFinder
	callsigns CallsignFinder
	history   CallHistoryFinder
	dupes     DupeChecker
	valuer    Valuer

	dataLock *sync.RWMutex

	myLocator                locator.Locator
	theirExchangeFields      []core.ExchangeField
	theirReportExchangeField core.ExchangeField
	theirNumberExchangeField core.ExchangeField

	totalScore core.BandScore
}

func NewCollector(dxcc DXCCFinder, callsigns CallsignFinder, history CallHistoryFinder,
	dupes DupeChecker, valuer Valuer) *Collector {

	return &Collector{
		dxcc:       dxcc,
		callsigns:  callsigns,
		history:    history,
		dupes:      dupes,
		valuer:     valuer,
		dataLock:   &sync.RWMutex{},
		totalScore: core.BandScore{Points: 1, Multis: 1},
	}
}

func (c *Collector) SetMyLocator(loc locator.Locator) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.myLocator = loc
}

func (c *Collector) SetTheirExchangeFields(fields []core.ExchangeField, theirReportExchangeField core.ExchangeField, theirNumberExchangeField core.ExchangeField) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.theirExchangeFields = fields
	c.theirReportExchangeField = theirReportExchangeField
	c.theirNumberExchangeField = theirNumberExchangeField
}

func (c *Collector) ScoreUpdated(score core.Score) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

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

func (c *Collector) UpdateValue(info *core.Callinfo, band core.Band, mode core.Mode) bool {
	if c.dxcc == nil || c.dupes == nil || c.valuer == nil {
		return false
	}

	workedQSOs, _ := c.dupes.FindWorkedQSOs(info.Call, band, mode)
	c.dataLock.RLock()
	c.addPredictedExchange(info, workedQSOs, nil)
	c.dataLock.RUnlock()

	c.addValue(info, band, mode)

	return true
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
	c.dataLock.RLock()
	c.addPredictedExchange(info, workedQSOs, currentExchange)
	c.dataLock.RUnlock()

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
	if !c.myLocator.IsZero() {
		entityLocator := locator.LatLonToLocator(entity.LatLon, 6)
		info.Azimuth = locator.Azimuth(c.myLocator, entityLocator)
		info.Distance = locator.Distance(c.myLocator, entityLocator)
	}

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
	// ATTENTION: temporal coupling! addPredictedExchange relies on addHistoryData putting
	// the historic exchange into the Callinfo.PredictedExchange field.
	info.PredictedExchange = predictExchange(c.theirExchangeFields, info.DXCCEntity, workedQSOs, currentExchange, info.PredictedExchange)
	info.PredictedExchange = c.clearUnpredictableFields(info.PredictedExchange)
}

// clearUnpredictableValues clears the values of unpredictable exchange fields (RST, serial).
func (c *Collector) clearUnpredictableFields(values []string) []string {
	result := make([]string, len(values))
	for i := range values {
		if i >= len(c.theirExchangeFields) {
			break
		}
		field := c.theirExchangeFields[i]
		switch field.Field {
		case c.theirReportExchangeField.Field:
			continue
		case c.theirNumberExchangeField.Field:
			if len(field.Properties) == 1 {
				continue
			}
		}
		result[i] = values[i]
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
