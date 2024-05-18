package bandmap

import (
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const similarCallsignThreshold = 2

type FalseEntryCheckResult int

const (
	DifferentEntries FalseEntryCheckResult = iota
	FirstIsFalse
	SecondIsFalse
)

func CheckFalseEntry(entry1, entry2 core.BandmapEntry) FalseEntryCheckResult {
	if checkLocallyVerified(entry1) || checkLocallyVerified(entry2) {
		return DifferentEntries
	}

	callsignEqual := entry1.Call == entry2.Call
	var callsignSimilar bool
	if callsignEqual {
		callsignSimilar = true
	} else {
		callsignSimilar = checkCallsignSimilar(entry1.Call, entry2.Call)
	}
	frequencySimilar := entry1.OnFrequency(entry2.Frequency)
	firstHasFalseSpotCount := checkFirstSpotCountIsFalse(entry1.SpotCount, entry2.SpotCount)
	secondHasFalseSpotCount := checkFirstSpotCountIsFalse(entry2.SpotCount, entry1.SpotCount)

	switch {
	case !callsignSimilar:
		return DifferentEntries
	case !frequencySimilar:
		return DifferentEntries
	case callsignSimilar && firstHasFalseSpotCount:
		return FirstIsFalse
	case callsignSimilar && secondHasFalseSpotCount:
		return SecondIsFalse
	default:
		return DifferentEntries
	}
}

func calculateCallsignDistance(call1, call2 callsign.Callsign) int {
	options := levenshtein.DefaultOptions
	return levenshtein.DistanceForStrings([]rune(call1.String()), []rune(call2.String()), options)
}

func checkLocallyVerified(entry core.BandmapEntry) bool {
	return entry.Source == core.WorkedSpot || entry.Source == core.ManualSpot
}

func checkCallsignSimilar(call1, call2 callsign.Callsign) bool {
	return calculateCallsignDistance(call1, call2) <= similarCallsignThreshold
}

func checkFirstSpotCountIsFalse(count1, count2 int) bool {
	return count1 == 1 && count2 > 2
}
