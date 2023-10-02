package bandmap

import (
	"testing"

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
