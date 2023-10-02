package bandmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

func Test(t *testing.T) {
	tt := []struct {
		desc     string
		entry1   core.BandmapEntry
		entry2   core.BandmapEntry
		expected FalseEntryCheckResult
	}{
		{
			desc: "different callsign, same frequency",
			entry1: core.BandmapEntry{
				Call: callsign.MustParse("DL0ABC"),
			},
			entry2: core.BandmapEntry{
				Call: callsign.MustParse("OK0ZZZ"),
			},
			expected: DifferentEntries,
		},
		{
			desc: "equal callsign, equal frequency",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				SpotCount: 100,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				SpotCount: 1,
			},
			expected: EqualEntries,
		},
		{
			desc: "equal callsign, similar frequency",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000000,
				SpotCount: 100,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000050,
				SpotCount: 1,
			},
			expected: EqualEntries,
		},
		{
			desc: "similar callsign, same frequency, second more spots than first",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000000,
				SpotCount: 1,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0AB"),
				Frequency: 7000000,
				SpotCount: 100,
			},
			expected: FirstIsFalse,
		},
		{
			desc: "similar callsign, same frequency, first more spots than second",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000000,
				SpotCount: 100,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0AB"),
				Frequency: 7000000,
				SpotCount: 1,
			},
			expected: SecondIsFalse,
		},
		{
			desc: "similar callsign, similar frequency, second more spots than first",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000000,
				SpotCount: 1,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0AB"),
				Frequency: 7000050,
				SpotCount: 100,
			},
			expected: FirstIsFalse,
		},
		{
			desc: "similar callsign, similar frequency, first more spots than second",
			entry1: core.BandmapEntry{
				Call:      callsign.MustParse("DL0ABC"),
				Frequency: 7000000,
				SpotCount: 100,
			},
			entry2: core.BandmapEntry{
				Call:      callsign.MustParse("DL0AB"),
				Frequency: 7000050,
				SpotCount: 1,
			},
			expected: SecondIsFalse,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actual := CheckFalseEntry(tc.entry1, tc.entry2)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEntries_cleanoutFalseEntries(t *testing.T) {
	now := time.Now()
	type testEntry struct {
		call      string
		frequency core.Frequency
		spots     int
	}
	addTestEntries := func(entries *Entries, testEntries []testEntry) {
		for _, testEntry := range testEntries {
			call := callsign.MustParse(testEntry.call)
			for i := 0; i < testEntry.spots; i++ {
				spot := core.Spot{
					Call:      call,
					Frequency: testEntry.frequency,
					Source:    core.ManualSpot,
					Time:      now,
				}
				entries.Add(spot)
			}
		}
		for _, entry := range entries.entries {
			entry.SpotCount = len(entry.spots)
		}
	}
	getCurrentTestEntries := func(entries *Entries) []testEntry {
		result := make([]testEntry, len(entries.entries))
		for i, entry := range entries.entries {
			result[i] = testEntry{
				call:      entry.Call.String(),
				frequency: entry.Frequency,
				spots:     len(entry.spots),
			}
		}
		return result
	}

	tt := []struct {
		desc     string
		entries  []testEntry
		expected []testEntry
	}{
		{
			desc: "remove false entry1 with single spot",
			entries: []testEntry{
				{call: "DL1ABC", frequency: 7010000, spots: 1},
				{call: "DL1AB", frequency: 7010050, spots: 4},
				{call: "DL1ZZZ", frequency: 7020000, spots: 1},
			},
			expected: []testEntry{
				{call: "DL1AB", frequency: 7010050, spots: 4},
				{call: "DL1ZZZ", frequency: 7020000, spots: 1},
			},
		},
		{
			desc: "remove false entry2 with single spot",
			entries: []testEntry{
				{call: "DL1ABC", frequency: 7010000, spots: 4},
				{call: "DL1AB", frequency: 7010050, spots: 1},
				{call: "DL1ZZZ", frequency: 7020000, spots: 1},
			},
			expected: []testEntry{
				{call: "DL1ABC", frequency: 7010000, spots: 4},
				{call: "DL1ZZZ", frequency: 7020000, spots: 1},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entries := NewEntries(countAllEntries)
			addTestEntries(entries, tc.entries)
			entries.cleanOutFalseEntries()
			actual := getCurrentTestEntries(entries)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
