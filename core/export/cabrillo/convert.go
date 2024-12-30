package cabrillo

import (
	"strings"

	"github.com/ftl/cabrillo"
	"github.com/ftl/conval"
)

func convalToCabrilloAssisted(category conval.Category) cabrillo.CategoryAssisted {
	if category.Assisted {
		return cabrillo.Assisted
	} else {
		return cabrillo.NonAssisted
	}
}

func convalToCabrilloBand(category conval.Category, availableBands []conval.ContestBand, qsoBand cabrillo.CategoryBand) cabrillo.CategoryBand {
	if category.BandCount == conval.AllBands {
		return cabrillo.BandAll
	}
	if hasBand(availableBands, qsoBand) {
		return qsoBand
	}
	if hasBand(category.Bands, qsoBand) {
		return qsoBand
	}
	if (len(category.Bands) > 0) && (category.Bands[0] == conval.BandAll) {
		return cabrillo.BandAll
	}
	if len(availableBands) == 1 {
		return convertBand(availableBands[0])
	}
	return ""
}

func hasBand(availableBands []conval.ContestBand, band cabrillo.CategoryBand) bool {
	for _, b := range availableBands {
		if convertBand(b) == band {
			return true
		}
	}
	return false
}

func convertBand(band conval.ContestBand) cabrillo.CategoryBand {
	return cabrillo.CategoryBand(strings.ToUpper(string(band)))
}

func convalToCabrilloMode(category conval.Category, availableModes []conval.Mode, qsoMode cabrillo.CategoryMode) cabrillo.CategoryMode {
	if hasMode(availableModes, qsoMode) {
		return qsoMode
	}
	if hasMode(category.Modes, qsoMode) {
		return qsoMode
	}
	if len(category.Modes) > 1 {
		return cabrillo.ModeMIXED
	}
	var mode conval.Mode
	if len(category.Modes) == 1 {
		mode = category.Modes[0]
	} else if len(availableModes) == 1 {
		mode = availableModes[0]
	}
	return convertMode(mode)
}

func hasMode(availableModes []conval.Mode, mode cabrillo.CategoryMode) bool {
	for _, m := range availableModes {
		if convertMode(m) == mode {
			return true
		}
	}
	return false
}

func convertMode(mode conval.Mode) cabrillo.CategoryMode {
	switch mode {
	case conval.ModeCW:
		return cabrillo.ModeCW
	case conval.ModeSSB:
		return cabrillo.ModeSSB
	case conval.ModeRTTY:
		return cabrillo.ModeRTTY
	case conval.ModeFM:
		return cabrillo.ModeFM
	case conval.ModeDigital:
		return cabrillo.ModeDIGI
	case conval.ModeALL:
		return cabrillo.ModeMIXED
	default:
		return ""
	}
}

func convalToCabrilloOperator(category conval.Category) cabrillo.CategoryOperator {
	if strings.ToUpper(category.Name) == "CHECKLOG" {
		return cabrillo.Checklog
	}
	switch category.Operator {
	case conval.SingleOperator:
		return cabrillo.SingleOperator
	case conval.MultiOperator:
		return cabrillo.MultiOperator
	default:
		return ""
	}
}

func convalToCabrilloPower(category conval.Category) cabrillo.CategoryPower {
	switch category.Power {
	case conval.HighPower:
		return cabrillo.HighPower
	case conval.LowPower:
		return cabrillo.LowPower
	case conval.QRPPower:
		return cabrillo.QRP
	default:
		return ""
	}
}
