package model

import (
	"strings"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/dxcc"
)

type Contest struct {
	Definition             *conval.Definition
	Name                   string
	ExchangeValues         []string
	GenerateSerialExchange bool
	GenerateReport         bool
	StartTime              time.Time
	EnableQTCs             bool
	DetailedMultis         bool

	MyExchangeFields         []ExchangeField
	MyReportExchangeField    ExchangeField
	MyNumberExchangeField    ExchangeField
	TheirExchangeFields      []ExchangeField
	TheirReportExchangeField ExchangeField
	TheirNumberExchangeField ExchangeField

	OperationModeSprint   bool
	CallHistoryFilename   string
	CallHistoryFieldNames []string

	QSOsGoal   int
	PointsGoal int
	MultisGoal int
}

func ContestSetup(station Station, entity dxcc.Prefix) conval.Setup {
	return conval.Setup{
		MyCall:      station.Callsign,
		MyContinent: conval.Continent(strings.ToLower(entity.Continent)),
		MyCountry:   conval.DXCCEntity(strings.ToLower(entity.PrimaryPrefix)),
		GridLocator: station.Locator,
	}
}

func (c Contest) MultiProperties(setup conval.Setup) []conval.Property {
	if c.Definition == nil {
		return nil
	}
	return c.Definition.MultiProperties(setup)
}

func (c Contest) Bands() []Band {
	if c.Definition == nil {
		return nil
	}
	bands := c.Definition.Bands
	if len(bands) == 1 && bands[0] == conval.BandAll {
		bands = conval.AllHFBands
	}

	result := make([]Band, len(bands))
	for i, band := range c.Definition.Bands {
		result[i] = Band(band)
	}
	return result
}

func (c Contest) Modes() []Mode {
	if c.Definition == nil {
		return nil
	}
	result := make([]Mode, 0, len(c.Definition.Modes))
	for _, mode := range c.Definition.Modes {
		switch mode {
		case conval.ModeALL:
			return Modes
		case conval.ModeCW:
			result = append(result, ModeCW)
		case conval.ModeSSB:
			result = append(result, ModeSSB)
		case conval.ModeFM:
			result = append(result, ModeFM)
		case conval.ModeRTTY:
			result = append(result, ModeRTTY)
		case conval.ModeDigital:
			result = append(result, ModeDigital)
		}
	}
	return result
}

func (c Contest) Started(now time.Time) bool {
	if c.StartTime.IsZero() {
		return true
	}
	if c.Definition == nil {
		return true
	}
	if c.Definition.Duration == 0 {
		return true
	}

	return now.After(c.StartTime)
}

func (c Contest) Finished(now time.Time) bool {
	if c.StartTime.IsZero() {
		return false
	}
	if c.Definition == nil {
		return false
	}
	if c.Definition.Duration == 0 {
		return false
	}

	return now.After(c.StartTime.Add(c.Definition.Duration))
}

func (c Contest) Running(now time.Time) bool {
	return c.Started(now) && !c.Finished(now)
}

func (c *Contest) UpdateExchangeFields() {
	c.MyExchangeFields = nil
	c.MyReportExchangeField = ExchangeField{}
	c.MyNumberExchangeField = ExchangeField{}
	c.TheirExchangeFields = nil
	c.TheirReportExchangeField = ExchangeField{}
	c.TheirNumberExchangeField = ExchangeField{}

	if c.Definition == nil {
		return
	}

	fieldDefinitions := c.Definition.ExchangeFields()

	c.MyExchangeFields = DefinitionsToExchangeFields(fieldDefinitions, MyExchangeField)
	for i, field := range c.MyExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.MyReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			if c.GenerateSerialExchange {
				field.ReadOnly = true
				field.Short = "#"
				field.Hint = "Serial Number"
				c.MyExchangeFields[i] = field
			}
			c.MyNumberExchangeField = field
		}
	}

	c.TheirExchangeFields = DefinitionsToExchangeFields(fieldDefinitions, TheirExchangeField)
	for _, field := range c.TheirExchangeFields {
		switch {
		case field.Properties.Contains(conval.RSTProperty):
			c.TheirReportExchangeField = field
		case field.Properties.Contains(conval.SerialNumberProperty):
			c.TheirNumberExchangeField = field
		}
	}
}
