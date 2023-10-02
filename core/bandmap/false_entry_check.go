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
	EqualEntries
	FirstIsFalse
	SecondIsFalse
)

func CheckFalseEntry(entry1, entry2 core.BandmapEntry) FalseEntryCheckResult {
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
		return DifferentEntries // TODO if one entry is much older than the other, the older one is false
	case callsignEqual:
		return EqualEntries
	case callsignSimilar && firstHasFalseSpotCount:
		return FirstIsFalse
	case callsignSimilar && secondHasFalseSpotCount:
		return SecondIsFalse
	default:
		return DifferentEntries
	}
}

func checkCallsignSimilar(call1, call2 callsign.Callsign) bool {
	options := levenshtein.DefaultOptions
	distance := levenshtein.DistanceForStrings([]rune(call1.String()), []rune(call2.String()), options)

	return distance <= similarCallsignThreshold
}

func checkFirstSpotCountIsFalse(count1, count2 int) bool {
	return count1 == 1 && count2 > 1
}
