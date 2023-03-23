package bandmap

import (
	"testing"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestEntry_Add_OnlySameCallAndSimilarFrequency(t *testing.T) {
	const frequency core.Frequency = 7035000
	tt := []struct {
		desc      string
		call      string
		frequency core.Frequency
		valid     bool
	}{
		{
			desc:      "same call and frequency",
			call:      "dl1abc",
			frequency: frequency,
			valid:     true,
		},
		{
			desc:      "same call, lower similar frequency",
			call:      "dl1abc",
			frequency: frequency - core.Frequency(spotFrequencyDeltaThreshold) + 1,
			valid:     true,
		},
		{
			desc:      "same call, higher similar frequency",
			call:      "dl1abc",
			frequency: frequency + core.Frequency(spotFrequencyDeltaThreshold) - 1,
			valid:     true,
		},
		{
			desc:      "same call, frequency to low",
			call:      "dl1abc",
			frequency: frequency - core.Frequency(spotFrequencyDeltaThreshold) - 1,
			valid:     false,
		},
		{
			desc:      "same call, frequency to high",
			call:      "dl1abc",
			frequency: frequency + core.Frequency(spotFrequencyDeltaThreshold) + 1,
			valid:     false,
		},
		{
			desc:      "different call, same frequency",
			call:      "dl2abc",
			frequency: frequency,
			valid:     false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entry := Entry{
				Call:      callsign.MustParse("dl1abc"),
				Frequency: frequency,
			}

			added := entry.Add(Spot{Call: callsign.MustParse(tc.call), Frequency: tc.frequency})
			assert.Equal(t, tc.valid, added)
		})
	}
}

func TestEntry_Add_MaintainsLastHeard(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{Call: call, Frequency: frequency}

	entry.Add(Spot{Call: call, Frequency: frequency, Time: now.Add(-1 * time.Hour)})
	assert.Equal(t, now.Add(-1*time.Hour), entry.LastHeard)

	entry.Add(Spot{Call: call, Frequency: frequency, Time: now.Add(-30 * time.Minute)})
	assert.Equal(t, now.Add(-30*time.Minute), entry.LastHeard)

	entry.Add(Spot{Call: call, Frequency: frequency, Time: now.Add(-10 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)

	entry.Add(Spot{Call: call, Frequency: frequency, Time: now.Add(-40 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)
}

func TestEntry_Add_MaintainsFrequency(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	entry := Entry{Call: call, Frequency: frequency}
	entry.Add(Spot{Call: call, Frequency: frequency})

	entry.Add(Spot{Call: call, Frequency: frequency + 18})
	assert.Equal(t, frequency+10, entry.Frequency)

	entry.Add(Spot{Call: call, Frequency: frequency - 13})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency-10, entry.Frequency)
}

func TestEntry_Add_MaintainsHighestRangedSource(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{Call: call, Frequency: frequency}

	entry.Add(Spot{Call: call, Frequency: frequency, Source: SkimmerSpot, Time: now})
	assert.Equal(t, SkimmerSpot, entry.Source)

	entry.Add(Spot{Call: call, Frequency: frequency, Source: RBNSpot, Time: now})
	assert.Equal(t, SkimmerSpot, entry.Source)

	entry.Add(Spot{Call: call, Frequency: frequency, Source: ManualSpot, Time: now})
	assert.Equal(t, ManualSpot, entry.Source)
}

func TestEntry_RemoveSpotsBefore(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{Call: call, Frequency: frequency}
	entry.Add(Spot{Call: call, Frequency: frequency, Source: ManualSpot, Time: now.Add(-10 * time.Hour)})
	entry.Add(Spot{Call: call, Frequency: frequency, Source: SkimmerSpot, Time: now.Add(-5 * time.Hour)})
	entry.Add(Spot{Call: call, Frequency: frequency, Source: RBNSpot, Time: now.Add(-1 * time.Hour)})
	entry.Add(Spot{Call: call, Frequency: frequency, Source: ClusterSpot, Time: now.Add(-30 * time.Minute)})
	entry.Add(Spot{Call: call, Frequency: frequency, Source: ClusterSpot, Time: now.Add(-1 * time.Hour)})

	valid := entry.RemoveSpotsBefore(now.Add(-10 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 5, entry.Len())
	assert.Equal(t, ManualSpot, entry.Source, "manual")

	valid = entry.RemoveSpotsBefore(now.Add(-5 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 4, entry.Len())
	assert.Equal(t, SkimmerSpot, entry.Source, "skimmer")

	valid = entry.RemoveSpotsBefore(now.Add(-40 * time.Minute))
	require.True(t, valid)
	assert.Equal(t, 1, entry.Len())
	assert.Equal(t, now.Add(-30*time.Minute), entry.spots[0].Time)
	assert.Equal(t, ClusterSpot, entry.Source, "cluster")

	valid = entry.RemoveSpotsBefore(now.Add(-10 * time.Minute))
	require.False(t, valid)
	assert.Equal(t, 0, entry.Len())
}

func TestEntry_ProximityFactor(t *testing.T) {
	const frequency core.Frequency = 7035000
	tt := []struct {
		desc      string
		frequency core.Frequency
		expected  float64
	}{
		{
			desc:      "same frequency",
			frequency: frequency,
			expected:  1.0,
		},
		{
			desc:      "lower frequency in proximity",
			frequency: frequency - core.Frequency(spotFrequencyProximityThreshold/2),
			expected:  0.5,
		},
		{
			desc:      "higher frequency in proximity",
			frequency: frequency + core.Frequency(spotFrequencyProximityThreshold/2),
			expected:  0.5,
		},
		{
			desc:      "frequency to low",
			frequency: frequency - core.Frequency(spotFrequencyProximityThreshold) - 1,
			expected:  0.0,
		},
		{
			desc:      "frequency to high",
			frequency: frequency + core.Frequency(spotFrequencyProximityThreshold) + 1,
			expected:  0.0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entry := Entry{
				Call:      callsign.MustParse("dl1abc"),
				Frequency: frequency,
			}

			actual := entry.ProximityFactor(tc.frequency)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEntries_AddNewEntry(t *testing.T) {
	now := time.Now()
	entries := NewEntries()
	assert.Equal(t, 0, entries.Len())

	entries.Add(Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3760000, Time: now})

	assert.Equal(t, 1, entries.Len())

	newEntry := entries.entries[0]
	assert.Equal(t, "DL1ABC", newEntry.Call.String())
	assert.Equal(t, core.Frequency(3760000), newEntry.Frequency)
	assert.Equal(t, now, newEntry.LastHeard)
	assert.Equal(t, 1, newEntry.Len())
}

func TestEntries_CleanOutOldEntries(t *testing.T) {
	now := time.Now()
	entries := NewEntries()

	entries.Add(Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)})
	entries.Add(Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-30 * time.Minute)})
	entries.Add(Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-10 * time.Minute)})
	entries.Add(Spot{Call: callsign.MustParse("dl2abc"), Frequency: 3535000, Time: now.Add(-10 * time.Hour)})

	assert.Equal(t, 2, entries.Len())
	assert.Equal(t, "DL1ABC", entries.entries[0].Call.String())
	assert.Equal(t, 3, entries.entries[0].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[0].LastHeard)

	entries.CleanOut(30*time.Minute, now)

	assert.Equal(t, 1, entries.Len())
	assert.Equal(t, "DL1ABC", entries.entries[0].Call.String())
	assert.Equal(t, 2, entries.entries[0].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[0].LastHeard)
}

func TestEntries_Notify(t *testing.T) {
	now := time.Now()
	entries := NewEntries()
	listener := new(testEntryListener)
	entries.Notify(listener)

	entries.Add(Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)})
	assert.Equal(t, "DL1ABC", listener.added[0].Call.String())

	entries.CleanOut(30*time.Minute, now)
	assert.Equal(t, "DL1ABC", listener.removed[0].Call.String())
}

func TestFilterSlice(t *testing.T) {
	input := []int{1, 10, 5, 2, 9, 7, 6, 3, 4}

	output := filterSlice(input, func(i int) bool {
		return i < 6
	})

	assert.Equal(t, []int{1, 5, 2, 3, 4}, output)
}

type testEntryListener struct {
	added   []Entry
	removed []Entry
}

func (t *testEntryListener) EntryAdded(e Entry) {
	t.added = append(t.added, e)
}

func (t *testEntryListener) EntryRemoved(e Entry) {
	t.removed = append(t.removed, e)
}
