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
			frequency: frequency - 300,
			valid:     true,
		},
		{
			desc:      "same call, higher similar frequency",
			call:      "dl1abc",
			frequency: frequency + 300,
			valid:     true,
		},
		{
			desc:      "same call, frequency to low",
			call:      "dl1abc",
			frequency: frequency - 301,
			valid:     false,
		},
		{
			desc:      "same call, frequency to high",
			call:      "dl1abc",
			frequency: frequency + 301,
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
				BandmapEntry: core.BandmapEntry{
					Call:      callsign.MustParse("dl1abc"),
					Frequency: frequency,
				},
			}

			_, added := entry.Add(core.Spot{Call: callsign.MustParse(tc.call), Frequency: tc.frequency})
			assert.Equal(t, tc.valid, added)
		})
	}
}

func TestEntry_Add_MaintainsLastHeard(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: core.BandmapEntry{Call: call, Frequency: frequency}}

	entry.Add(core.Spot{Call: call, Frequency: frequency, Time: now.Add(-1 * time.Hour)})
	assert.Equal(t, now.Add(-1*time.Hour), entry.LastHeard)

	entry.Add(core.Spot{Call: call, Frequency: frequency, Time: now.Add(-30 * time.Minute)})
	assert.Equal(t, now.Add(-30*time.Minute), entry.LastHeard)

	entry.Add(core.Spot{Call: call, Frequency: frequency, Time: now.Add(-10 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)

	entry.Add(core.Spot{Call: call, Frequency: frequency, Time: now.Add(-40 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)
}

func TestEntry_Add_MaintainsFrequency(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	entry := Entry{BandmapEntry: core.BandmapEntry{Call: call, Frequency: frequency}}
	entry.Add(core.Spot{Call: call, Frequency: frequency})

	entry.Add(core.Spot{Call: call, Frequency: frequency + 18})
	assert.Equal(t, frequency+10, entry.Frequency)

	entry.Add(core.Spot{Call: call, Frequency: frequency - 13})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(core.Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(core.Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency-10, entry.Frequency)
}

func TestEntry_Add_MaintainsHighestRankedSource(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: core.BandmapEntry{Call: call, Frequency: frequency}}

	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.SkimmerSpot, Time: now})
	assert.Equal(t, core.SkimmerSpot, entry.Source)

	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.RBNSpot, Time: now})
	assert.Equal(t, core.SkimmerSpot, entry.Source)

	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.ManualSpot, Time: now})
	assert.Equal(t, core.ManualSpot, entry.Source)
}

func TestEntry_RemoveSpotsBefore(t *testing.T) {
	call := callsign.MustParse("dl1abc")
	frequency := core.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: core.BandmapEntry{Call: call, Frequency: frequency}}
	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.ManualSpot, Time: now.Add(-10 * time.Hour)})
	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.SkimmerSpot, Time: now.Add(-5 * time.Hour)})
	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.RBNSpot, Time: now.Add(-1 * time.Hour)})
	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.ClusterSpot, Time: now.Add(-30 * time.Minute)})
	entry.Add(core.Spot{Call: call, Frequency: frequency, Source: core.ClusterSpot, Time: now.Add(-1 * time.Hour)})

	valid := entry.RemoveSpotsBefore(now.Add(-10 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 5, entry.Len())
	assert.Equal(t, core.ManualSpot, entry.Source, "manual")

	valid = entry.RemoveSpotsBefore(now.Add(-5 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 4, entry.Len())
	assert.Equal(t, core.SkimmerSpot, entry.Source, "skimmer")

	valid = entry.RemoveSpotsBefore(now.Add(-40 * time.Minute))
	require.True(t, valid)
	assert.Equal(t, 1, entry.Len())
	assert.Equal(t, now.Add(-30*time.Minute), entry.spots[0].Time)
	assert.Equal(t, core.ClusterSpot, entry.Source, "cluster")

	valid = entry.RemoveSpotsBefore(now.Add(-10 * time.Minute))
	require.False(t, valid)
	assert.Equal(t, 0, entry.Len())
}

func TestEntries_AddNewEntry(t *testing.T) {
	now := time.Now()
	entries := NewEntries(countAllEntries)
	assert.Equal(t, 0, entries.Len())

	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3760000, Time: now}, now, defaultWeights)

	assert.Equal(t, 1, entries.Len())

	newEntry := entries.entries[0]
	assert.Equal(t, "DL1ABC", newEntry.Call.String())
	assert.Equal(t, core.Frequency(3760000), newEntry.Frequency)
	assert.Equal(t, now, newEntry.LastHeard)
	assert.Equal(t, 0, newEntry.Index)
	assert.Equal(t, 1, newEntry.Len())
}

func TestEntries_findIndexForInsert(t *testing.T) {
	tt := []struct {
		desc     string
		fixture  []int
		value    int
		expected int
	}{
		{
			desc:     "empty",
			value:    1,
			expected: 0,
		},
		{
			desc:     "before first",
			fixture:  []int{2, 3, 4},
			value:    1,
			expected: 0,
		},
		{
			desc:     "at the first",
			fixture:  []int{2, 3, 4},
			value:    2,
			expected: 0,
		},
		{
			desc:     "after the first",
			fixture:  []int{2, 4, 5},
			value:    3,
			expected: 1,
		},
		{
			desc:     "at the center",
			fixture:  []int{2, 3, 5, 6},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the existing center",
			fixture:  []int{2, 3, 4, 5, 6},
			value:    4,
			expected: 2,
		},
		{
			desc:     "before the last",
			fixture:  []int{2, 3, 5},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the last",
			fixture:  []int{2, 3, 4},
			value:    4,
			expected: 2,
		},
		{
			desc:     "after last",
			fixture:  []int{2, 3, 4},
			value:    5,
			expected: 3,
		},
	}
	newEntry := func(value int) *Entry {
		return &Entry{BandmapEntry: core.BandmapEntry{Frequency: core.Frequency(value)}}
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entries := &Entries{
				entries: make([]*Entry, 0, len(tc.fixture)+1),
				order:   core.BandmapByFrequency,
			}
			for _, value := range tc.fixture {
				entries.entries = append(entries.entries, newEntry(value))
			}

			actual := entries.findIndexForInsert(newEntry(tc.value))

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEntries_insert(t *testing.T) {
	tt := []struct {
		desc     string
		fixture  []int
		value    int
		expected int
	}{
		{
			desc:     "empty",
			value:    1,
			expected: 0,
		},
		{
			desc:     "before first",
			fixture:  []int{2, 3, 4},
			value:    1,
			expected: 0,
		},
		{
			desc:     "at the first",
			fixture:  []int{2, 3, 4},
			value:    2,
			expected: 0,
		},
		{
			desc:     "after the first",
			fixture:  []int{2, 4, 5},
			value:    3,
			expected: 1,
		},
		{
			desc:     "at the center",
			fixture:  []int{2, 3, 5, 6},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the existing center",
			fixture:  []int{2, 3, 4, 5, 6},
			value:    4,
			expected: 2,
		},
		{
			desc:     "before the last",
			fixture:  []int{2, 3, 5},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the last",
			fixture:  []int{2, 3, 4},
			value:    4,
			expected: 2,
		},
		{
			desc:     "after last",
			fixture:  []int{2, 3, 4},
			value:    5,
			expected: 3,
		},
	}
	newEntry := func(value int) *Entry {
		return &Entry{BandmapEntry: core.BandmapEntry{Frequency: core.Frequency(value)}}
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entries := &Entries{
				entries: make([]*Entry, 0, len(tc.fixture)+1),
				order:   core.BandmapByFrequency,
			}
			for i, value := range tc.fixture {
				entry := newEntry(value)
				entry.Index = i
				entries.entries = append(entries.entries, entry)
			}

			entry := newEntry(tc.value)
			entry.Label = "inserted"
			entries.insert(entry)

			assert.Equal(t, "inserted", entries.entries[tc.expected].Label, "label")

			for i, e := range entries.entries {
				assert.Equal(t, i, e.Index, "index %d", i)
			}
		})
	}
}

func TestEntries_CleanOutOldEntries(t *testing.T) {
	now := time.Now()
	entries := NewEntries(countAllEntries)

	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)}, now, defaultWeights)
	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-30 * time.Minute)}, now, defaultWeights)
	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-10 * time.Minute)}, now, defaultWeights)
	entries.Add(core.Spot{Call: callsign.MustParse("dl2abc"), Frequency: 3535500, Time: now.Add(-10 * time.Hour)}, now, defaultWeights)

	assert.Equal(t, 2, entries.Len())
	assert.Equal(t, "DL1ABC", entries.entries[0].Call.String())
	assert.Equal(t, 3, entries.entries[0].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[0].LastHeard)

	entries.CleanOut(30*time.Minute, now, defaultWeights)

	assert.Equal(t, 1, entries.Len())
	assert.Equal(t, "DL1ABC", entries.entries[0].Call.String())
	assert.Equal(t, 2, entries.entries[0].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[0].LastHeard)
}

func TestEntries_Notify(t *testing.T) {
	now := time.Now()
	entries := NewEntries(countAllEntries)
	listener := new(testEntryListener)
	entries.Notify(listener)

	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)}, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.added[0].Call.String())

	entries.Add(core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-40 * time.Minute)}, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.updated[0].Call.String())

	entries.CleanOut(30*time.Minute, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.removed[0].Call.String())
}

func TestEntry_Matches(t *testing.T) {
	now := time.Now()
	spot := core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3535000, Time: now.Add(-5 * time.Minute)}
	entry := NewEntry(spot)
	assert.Equal(t, core.UnknownSpotQuality, entry.Quality)

	similarSpot := core.Spot{Call: callsign.MustParse("dl2abc"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match := entry.Matches(similarSpot)
	assert.False(t, match)
	assert.Equal(t, core.UnknownSpotQuality, quality)
	_, added := entry.Add(similarSpot)
	assert.False(t, added)

	quality, match = entry.Matches(spot)
	assert.True(t, match)
	assert.Equal(t, core.UnknownSpotQuality, quality)

	_, added = entry.Add(spot)
	assert.True(t, added)
	assert.Equal(t, core.UnknownSpotQuality, entry.Quality)

	quality, match = entry.Matches(spot)
	assert.True(t, match)
	assert.Equal(t, core.ValidSpotQuality, quality)

	_, added = entry.Add(spot)
	assert.True(t, added)
	assert.Equal(t, core.ValidSpotQuality, entry.Quality)

	qsySpot := core.Spot{Call: callsign.MustParse("dl1abc"), Frequency: 3545000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(qsySpot)
	assert.False(t, match)
	assert.Equal(t, core.QSYSpotQuality, quality)
	_, added = entry.Add(qsySpot)
	assert.False(t, added)

	bustedSpot := core.Spot{Call: callsign.MustParse("dl2abc"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(bustedSpot)
	assert.False(t, match)
	assert.Equal(t, core.BustedSpotQuality, quality)
	_, added = entry.Add(bustedSpot)
	assert.False(t, added)

	completeDifferentSpot := core.Spot{Call: callsign.MustParse("dl3xyz"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(completeDifferentSpot)
	assert.False(t, match)
	assert.Equal(t, core.UnknownSpotQuality, quality)
	_, added = entry.Add(completeDifferentSpot)
	assert.False(t, added)
}

func TestFilterSlice(t *testing.T) {
	input := []int{1, 10, 5, 2, 9, 7, 6, 3, 4}

	output := filterSlice(input, func(i int) bool {
		return i < 6
	})

	assert.Equal(t, []int{1, 5, 2, 3, 4}, output)
}

type testEntryListener struct {
	added   []core.BandmapEntry
	updated []core.BandmapEntry
	removed []core.BandmapEntry
}

func (t *testEntryListener) EntryAdded(e core.BandmapEntry) {
	t.added = append(t.added, e)
}

func (t *testEntryListener) EntryUpdated(e core.BandmapEntry) {
	t.updated = append(t.updated, e)
}

func (t *testEntryListener) EntryRemoved(e core.BandmapEntry) {
	t.removed = append(t.removed, e)
}

func countAllEntries(core.BandmapEntry) bool {
	return true
}
