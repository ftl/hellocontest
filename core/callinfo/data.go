package callinfo

import (
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
}

func (c *Collector) GetInfo(callInput string, band core.Band, mode core.Mode, exchange []string) core.Callinfo {
	result := core.Callinfo{
		Input: callInput,
	}

	c.addCallsign(&result)
	c.addDXCC(&result)

	return result
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

func (c *Collector) addDXCC(info *core.Callinfo) bool {
	if c.dxcc == nil {
		return false
	}

	entity, found := c.dxcc.Find(info.Input)
	if !found {
		return false
	}

	info.DXCCName = entity.Name
	info.PrimaryPrefix = entity.PrimaryPrefix
	info.Continent = entity.Continent
	info.ITUZone = int(entity.ITUZone)
	info.CQZone = int(entity.CQZone)

	return true
}
