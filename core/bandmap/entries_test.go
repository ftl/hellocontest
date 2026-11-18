package bandmap

import (
	"testing"
	"time"

	"github.com/ftl/conval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/model"
)

func TestEntry_Add_OnlySameCallAndSimilarFrequency(t *testing.T) {
	const frequency model.Frequency = 7035000
	tt := []struct {
		desc      string
		call      string
		frequency model.Frequency
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
				BandmapEntry: model.BandmapEntry{
					Call:      model.MustParseCallsign("dl1abc"),
					Frequency: frequency,
				},
			}

			_, added := entry.Add(model.Spot{Call: model.MustParseCallsign(tc.call), Frequency: tc.frequency})
			assert.Equal(t, tc.valid, added)
		})
	}
}

func TestEntry_Add_MaintainsLastHeard(t *testing.T) {
	call := model.MustParseCallsign("dl1abc")
	frequency := model.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: model.BandmapEntry{Call: call, Frequency: frequency}}

	entry.Add(model.Spot{Call: call, Frequency: frequency, Time: now.Add(-1 * time.Hour)})
	assert.Equal(t, now.Add(-1*time.Hour), entry.LastHeard)

	entry.Add(model.Spot{Call: call, Frequency: frequency, Time: now.Add(-30 * time.Minute)})
	assert.Equal(t, now.Add(-30*time.Minute), entry.LastHeard)

	entry.Add(model.Spot{Call: call, Frequency: frequency, Time: now.Add(-10 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)

	entry.Add(model.Spot{Call: call, Frequency: frequency, Time: now.Add(-40 * time.Minute)})
	assert.Equal(t, now.Add(-10*time.Minute), entry.LastHeard)
}

func TestEntry_Add_MaintainsFrequency(t *testing.T) {
	call := model.MustParseCallsign("dl1abc")
	frequency := model.Frequency(7035000)
	entry := Entry{BandmapEntry: model.BandmapEntry{Call: call, Frequency: frequency}}
	entry.Add(model.Spot{Call: call, Frequency: frequency})

	entry.Add(model.Spot{Call: call, Frequency: frequency + 18})
	assert.Equal(t, frequency+10, entry.Frequency)

	entry.Add(model.Spot{Call: call, Frequency: frequency - 13})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(model.Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency, entry.Frequency)

	entry.Add(model.Spot{Call: call, Frequency: frequency - 24})
	assert.Equal(t, frequency-10, entry.Frequency)
}

func TestEntry_Add_MaintainsHighestRankedSource(t *testing.T) {
	call := model.MustParseCallsign("dl1abc")
	frequency := model.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: model.BandmapEntry{Call: call, Frequency: frequency}}

	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.SkimmerSpot, Time: now})
	assert.Equal(t, model.SkimmerSpot, entry.Source)

	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.RBNSpot, Time: now})
	assert.Equal(t, model.SkimmerSpot, entry.Source)

	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.ManualSpot, Time: now})
	assert.Equal(t, model.ManualSpot, entry.Source)
}

func TestEntry_RemoveSpotsBefore(t *testing.T) {
	call := model.MustParseCallsign("dl1abc")
	frequency := model.Frequency(7035000)
	now := time.Now()
	entry := Entry{BandmapEntry: model.BandmapEntry{Call: call, Frequency: frequency}}
	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.ManualSpot, Time: now.Add(-10 * time.Hour)})
	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.SkimmerSpot, Time: now.Add(-5 * time.Hour)})
	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.RBNSpot, Time: now.Add(-1 * time.Hour)})
	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.ClusterSpot, Time: now.Add(-30 * time.Minute)})
	entry.Add(model.Spot{Call: call, Frequency: frequency, Source: model.ClusterSpot, Time: now.Add(-1 * time.Hour)})

	valid := entry.RemoveSpotsBefore(now.Add(-10 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 5, entry.Len())
	assert.Equal(t, model.ManualSpot, entry.Source, "manual")

	valid = entry.RemoveSpotsBefore(now.Add(-5 * time.Hour))
	require.True(t, valid)
	assert.Equal(t, 4, entry.Len())
	assert.Equal(t, model.SkimmerSpot, entry.Source, "skimmer")

	valid = entry.RemoveSpotsBefore(now.Add(-40 * time.Minute))
	require.True(t, valid)
	assert.Equal(t, 1, entry.Len())
	assert.Equal(t, now.Add(-30*time.Minute), entry.spots[0].Time)
	assert.Equal(t, model.ClusterSpot, entry.Source, "cluster")

	valid = entry.RemoveSpotsBefore(now.Add(-10 * time.Minute))
	require.False(t, valid)
	assert.Equal(t, 0, entry.Len())
}

func TestEntries_AddNewEntry(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	assert.Equal(t, 0, entries.Len())

	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3760000, Time: now}, now, defaultWeights)

	assert.Equal(t, 1, entries.Len())

	newEntry := entries.entries[0]
	assert.Equal(t, "DL1ABC", newEntry.Call.String())
	assert.Equal(t, model.Frequency(3760000), newEntry.Frequency)
	assert.Equal(t, now, newEntry.LastHeard)
	assert.Equal(t, model.BandmapEntryID(1), newEntry.ID)
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
			fixture:  []int{4, 3, 2},
			value:    5,
			expected: 0,
		},
		{
			desc:     "at the first",
			fixture:  []int{4, 3, 2},
			value:    4,
			expected: 0,
		},
		{
			desc:     "after the first",
			fixture:  []int{5, 3, 2},
			value:    4,
			expected: 1,
		},
		{
			desc:     "at the center",
			fixture:  []int{6, 5, 3, 2},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the existing center",
			fixture:  []int{6, 5, 4, 3, 2},
			value:    4,
			expected: 2,
		},
		{
			desc:     "before the last",
			fixture:  []int{5, 4, 2},
			value:    3,
			expected: 2,
		},
		{
			desc:     "at the last",
			fixture:  []int{4, 3, 1},
			value:    2,
			expected: 2,
		},
		{
			desc:     "after last",
			fixture:  []int{4, 3, 2},
			value:    1,
			expected: 3,
		},
	}
	newEntry := func(value int) *Entry {
		return &Entry{BandmapEntry: model.BandmapEntry{Frequency: model.Frequency(value)}}
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entries := &Entries{
				entries: make([]*Entry, 0, len(tc.fixture)+1),
				order:   model.BandmapByFrequency,
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
			fixture:  []int{4, 3, 2},
			value:    5,
			expected: 0,
		},
		{
			desc:     "at the first",
			fixture:  []int{4, 3, 2},
			value:    4,
			expected: 0,
		},
		{
			desc:     "after the first",
			fixture:  []int{5, 3, 2},
			value:    4,
			expected: 1,
		},
		{
			desc:     "at the center",
			fixture:  []int{6, 5, 3, 2},
			value:    4,
			expected: 2,
		},
		{
			desc:     "at the existing center",
			fixture:  []int{6, 5, 4, 3, 2},
			value:    4,
			expected: 2,
		},
		{
			desc:     "before the last",
			fixture:  []int{5, 4, 2},
			value:    3,
			expected: 2,
		},
		{
			desc:     "at the last",
			fixture:  []int{4, 3, 1},
			value:    2,
			expected: 2,
		},
		{
			desc:     "after last",
			fixture:  []int{4, 3, 2},
			value:    1,
			expected: 3,
		},
	}
	newEntry := func(value int) *Entry {
		return &Entry{BandmapEntry: model.BandmapEntry{Frequency: model.Frequency(value)}}
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			entries := &Entries{
				entries: make([]*Entry, 0, len(tc.fixture)+1),
				order:   model.BandmapByFrequency,
			}
			for i, value := range tc.fixture {
				entry := newEntry(value)
				entry.ID = model.BandmapEntryID(i + 1)
				entries.entries = append(entries.entries, entry)
			}
			entries.lastID = model.BandmapEntryID(len(entries.entries))

			entry := newEntry(tc.value)
			entry.Label = "inserted"
			entries.insert(entry)

			assert.Equal(t, "inserted", entries.entries[tc.expected].Label, "label")
			assert.Equal(t, len(entries.entries), int(entries.entries[tc.expected].ID), "id")
		})
	}
}

func TestEntries_QueryByFilters(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.Add(model.Spot{Call: model.MustParseCallsign("DL1ABC"), Frequency: 7020000, Band: model.Band40m, Source: model.ManualSpot}, now.Add(-1*time.Minute), defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("DL2ABC"), Frequency: 14015000, Band: model.Band20m, Source: model.ManualSpot}, now.Add(-2*time.Minute), defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("DL3ABC"), Frequency: 14030000, Band: model.Band20m, Source: model.ClusterSpot}, now.Add(-2*time.Minute), defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("DL4ABC"), Frequency: 28010000, Band: model.Band10m, Source: model.ClusterSpot}, now.Add(-2*time.Minute), defaultWeights)
	actualCallsigns := make([]string, 0, 4)

	tt := []struct {
		desc     string
		filters  []model.BandmapFilter
		expected []string
	}{
		{
			desc:     "all",
			filters:  nil,
			expected: []string{"DL1ABC", "DL2ABC", "DL3ABC", "DL4ABC"},
		},
		{
			desc:     "on 20m",
			filters:  []model.BandmapFilter{model.OnBand(model.Band20m)},
			expected: []string{"DL2ABC", "DL3ABC"},
		},
		{
			desc:     "on 20m from cluster",
			filters:  []model.BandmapFilter{model.OnBand(model.Band20m), model.FromSource(model.ClusterSpot)},
			expected: []string{"DL3ABC"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actualEntries := entries.Query(model.BandmapByFrequency, tc.filters...)
			if len(actualCallsigns) > 0 {
				actualCallsigns = actualCallsigns[:0]
			}
			for _, entry := range actualEntries {
				actualCallsigns = append(actualCallsigns, entry.Call.String())
			}
			assert.Equal(t, tc.expected, actualCallsigns)
		})
	}
}

func TestEntry_ReportedQualityWins(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now})
	entry.applyReportedQuality()
	require.Equal(t, model.UnknownSpotQuality, entry.Quality)

	_, added := entry.Add(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now.Add(time.Minute), Source: model.SkimmerSpot, Quality: model.ValidSpotQuality,
	})

	require.True(t, added)
	assert.Equal(t, model.ValidSpotQuality, entry.Quality, "one spot with a reported quality is enough")
}

func TestEntry_ReportedQualitySurvivesAnUpdate(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.SkimmerSpot, Quality: model.ValidSpotQuality,
	})
	entry.applyReportedQuality()

	entry.update()

	assert.Equal(t, 1, entry.SpotCount)
	assert.Equal(t, model.ValidSpotQuality, entry.Quality, "the bandmap does not judge a reported quality")
}

func TestEntry_ReportedQualityOfTheSourceWithTheHighestPriority(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.RBNSpot, Quality: model.ValidSpotQuality,
	})
	entry.applyReportedQuality()

	entry.Add(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now.Add(time.Minute), Source: model.SkimmerSpot, Quality: model.BustedSpotQuality,
	})

	assert.Equal(t, model.BustedSpotQuality, entry.Quality, "the skimmer is nearer than the RBN")
}

func TestEntry_LatestReportedQualityOfTheSameSource(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.SkimmerSpot, Quality: model.BustedSpotQuality,
	})
	entry.applyReportedQuality()

	entry.Add(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now.Add(time.Minute), Source: model.SkimmerSpot, Quality: model.ValidSpotQuality,
	})

	assert.Equal(t, model.ValidSpotQuality, entry.Quality, "the source corrected itself")
}

func TestEntry_SNROfTheSourceWithTheHighestPriority(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.RBNSpot, SNR: 9,
	})
	entry.applyReportedSNR()
	require.Equal(t, 9.0, entry.SNR)

	entry.Add(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now.Add(time.Minute), Source: model.SkimmerSpot, SNR: 25,
	})

	assert.Equal(t, 25.0, entry.SNR, "the receiver of the nearest source decides")
}

func TestEntry_LatestSNROfTheSameSource(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.SkimmerSpot, SNR: 25,
	})
	entry.applyReportedSNR()

	entry.Add(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now.Add(time.Minute), Source: model.SkimmerSpot, SNR: 12,
	})

	assert.Equal(t, 12.0, entry.SNR, "the signal changed")
}

func TestEntry_WithoutAReportedSNR(t *testing.T) {
	now := time.Now()
	entry := NewEntry(model.Spot{
		Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m,
		Time: now, Source: model.ClusterSpot,
	})

	entry.applyReportedSNR()

	assert.Equal(t, 0.0, entry.SNR, "a cluster spot has no signal to noise ratio")
}

func TestEntries_Remove(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 3540000, Band: model.Band80m, Time: now}, now, defaultWeights)

	removed := entries.Remove(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m})

	assert.True(t, removed)
	require.Equal(t, 1, entries.Len())
	assert.Equal(t, "DL2ABC", entries.entries[0].Call.String())
}

func TestEntries_Remove_OnlyOnTheBandOfTheSpot(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 14035000, Band: model.Band20m, Time: now}, now, defaultWeights)

	removed := entries.Remove(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m})

	assert.True(t, removed)
	require.Equal(t, 1, entries.Len())
	assert.Equal(t, model.Band20m, entries.entries[0].Band, "the same station runs on another band")
}

func TestEntries_Remove_OnAnotherFrequency(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now}, now, defaultWeights)

	removed := entries.Remove(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3545000, Band: model.Band80m})

	assert.False(t, removed, "the entry is on another frequency")
	assert.Equal(t, 1, entries.Len())
}

func TestEntries_Remove_NotifiesTheListeners(t *testing.T) {
	now := time.Now()
	notifier := &Notifier{asyncRunner: func(f func()) { f() }}
	entries := NewEntries(notifier, countAllEntries)
	listener := new(entryRemovedSpy)
	notifier.Notify(listener)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now}, now, defaultWeights)

	entries.Remove(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m})

	require.Len(t, listener.entries, 1)
	assert.Equal(t, "DL1ABC", listener.entries[0].Call.String())
}

type entryRemovedSpy struct {
	entries []model.BandmapEntry
}

func (s *entryRemovedSpy) EntryRemoved(entry model.BandmapEntry) {
	s.entries = append(s.entries, entry)
}

func TestEntries_CleanOutOldEntries(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)

	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-30 * time.Minute)}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-10 * time.Minute)}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 3535500, Time: now.Add(-10 * time.Hour)}, now, defaultWeights)
	entries.Add(model.Spot{Source: model.WorkedSpot, Call: model.MustParseCallsign("dl3abc"), Frequency: 3535500, Time: now.Add(-50 * time.Hour)}, now, defaultWeights)

	assert.Equal(t, 3, entries.Len())
	assert.Equal(t, "DL3ABC", entries.entries[0].Call.String())
	assert.Equal(t, "DL2ABC", entries.entries[1].Call.String())
	assert.Equal(t, "DL1ABC", entries.entries[2].Call.String())
	assert.Equal(t, 3, entries.entries[2].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[2].LastHeard)

	entries.CleanOut(30*time.Minute, now, defaultWeights)

	assert.Equal(t, 1, entries.Len())
	assert.Equal(t, "DL1ABC", entries.entries[0].Call.String())
	assert.Equal(t, 2, entries.entries[0].Len())
	assert.Equal(t, now.Add(-10*time.Minute), entries.entries[0].LastHeard)
}

func TestEntries_Notify(t *testing.T) {
	now := time.Now()
	notifier := &Notifier{
		asyncRunner: func(f func()) {
			f()
		},
	}
	entries := NewEntries(notifier, countAllEntries)
	listener := new(testEntryListener)
	notifier.Notify(listener)

	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-1 * time.Hour)}, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.added[0].Call.String())

	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-40 * time.Minute)}, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.updated[0].Call.String())

	entries.CleanOut(30*time.Minute, now, defaultWeights)
	assert.Equal(t, "DL1ABC", listener.removed[0].Call.String())
}

func TestEntry_Matches(t *testing.T) {
	now := time.Now()
	spot := model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Time: now.Add(-5 * time.Minute)}
	entry := NewEntry(spot)
	assert.Equal(t, model.UnknownSpotQuality, entry.Quality)

	similarSpot := model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match := entry.Matches(similarSpot)
	assert.False(t, match)
	assert.Equal(t, model.UnknownSpotQuality, quality)
	_, added := entry.Add(similarSpot)
	assert.False(t, added)

	quality, match = entry.Matches(spot)
	assert.True(t, match)
	assert.Equal(t, model.UnknownSpotQuality, quality)

	_, added = entry.Add(spot)
	assert.True(t, added)
	assert.Equal(t, model.UnknownSpotQuality, entry.Quality)

	quality, match = entry.Matches(spot)
	assert.True(t, match)
	assert.Equal(t, model.ValidSpotQuality, quality)

	_, added = entry.Add(spot)
	assert.True(t, added)
	assert.Equal(t, model.ValidSpotQuality, entry.Quality)

	qsySpot := model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3545000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(qsySpot)
	assert.False(t, match)
	assert.Equal(t, model.QSYSpotQuality, quality)
	_, added = entry.Add(qsySpot)
	assert.False(t, added)

	bustedSpot := model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(bustedSpot)
	assert.False(t, match)
	assert.Equal(t, model.BustedSpotQuality, quality)
	_, added = entry.Add(bustedSpot)
	assert.False(t, added)

	completeDifferentSpot := model.Spot{Call: model.MustParseCallsign("dl3xyz"), Frequency: 3535000, Time: now.Add(-2 * time.Minute)}
	quality, match = entry.Matches(completeDifferentSpot)
	assert.False(t, match)
	assert.Equal(t, model.UnknownSpotQuality, quality)
	_, added = entry.Add(completeDifferentSpot)
	assert.False(t, added)
}

func TestCalculateCallsignDistance(t *testing.T) {
	tt := []struct {
		call1    string
		call2    string
		expected int
	}{
		{"dl0abc", "dl0abc", 0},
		{"dl0abc", "dl0ab", 1},
		{"dl0abc", "dl0a", 2},
	}
	for _, tc := range tt {
		t.Run(tc.call1+" "+tc.call2, func(t *testing.T) {
			call1 := model.MustParseCallsign(tc.call1)
			call2 := model.MustParseCallsign(tc.call2)
			actual := calculateCallsignDistance(call1, call2)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFilterSlice(t *testing.T) {
	input := []int{1, 10, 5, 2, 9, 7, 6, 3, 4}

	output := filterSlice(input, func(i int) bool {
		return i < 6
	})

	assert.Equal(t, []int{1, 5, 2, 3, 4}, output)
}

func TestEntries_MarkAsWorked(t *testing.T) {
	tt := []struct {
		band     model.Band
		mode     model.Mode
		expected []model.SpotType
	}{
		{
			band: model.NoBand,
			mode: model.NoMode,
			expected: []model.SpotType{
				model.WorkedSpot,
				model.WorkedSpot,
				model.WorkedSpot,
				model.UnknownSpot,
				model.WorkedSpot,
			},
		},
		{
			band: model.Band20m,
			mode: model.NoMode,
			expected: []model.SpotType{
				model.WorkedSpot,
				model.WorkedSpot,
				model.UnknownSpot,
				model.UnknownSpot,
				model.UnknownSpot,
			},
		},
		{
			band: model.Band20m,
			mode: model.ModeCW,
			expected: []model.SpotType{
				model.UnknownSpot,
				model.WorkedSpot,
				model.UnknownSpot,
				model.UnknownSpot,
				model.UnknownSpot,
			},
		},
	}

	for _, tc := range tt {
		t.Run(string(tc.band)+" "+string(tc.mode), func(t *testing.T) {
			now := time.Now()
			entries := NewEntries(&Notifier{}, countAllEntries)

			// the entries will be sorted by band and time
			entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Band: model.Band20m, Frequency: 14114000, Mode: model.ModeSSB, Time: now.Add(-11 * time.Minute)}, now, defaultWeights)
			entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Band: model.Band20m, Frequency: 14014000, Mode: model.ModeCW, Time: now.Add(-10 * time.Minute)}, now, defaultWeights)
			entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Band: model.Band40m, Frequency: 7073000, Time: now.Add(-30 * time.Minute)}, now, defaultWeights)
			entries.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Band: model.Band80m, Frequency: 3535500, Time: now.Add(-12 * time.Hour)}, now, defaultWeights)
			entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Band: model.Band80m, Frequency: 3535000, Time: now.Add(-1 * time.Hour)}, now, defaultWeights)

			entries.MarkAsWorked(model.MustParseCallsign("dl1abc"), tc.band, tc.mode)

			for i, spotType := range tc.expected {
				assert.Equal(t, spotType, entries.entries[i].Source, "entry #%d", i)
			}

			for _, e := range entries.entries {
				e.update()
			}

			for i, spotType := range tc.expected {
				assert.Equal(t, spotType, entries.entries[i].Source, "entry #%d", i)
			}
		})
	}
}

type testEntryListener struct {
	added   []model.BandmapEntry
	updated []model.BandmapEntry
	removed []model.BandmapEntry
}

func (t *testEntryListener) EntryAdded(e model.BandmapEntry) {
	t.added = append(t.added, e)
}

func (t *testEntryListener) EntryUpdated(e model.BandmapEntry) {
	t.updated = append(t.updated, e)
}

func (t *testEntryListener) EntryRemoved(e model.BandmapEntry) {
	t.removed = append(t.removed, e)
}

func countAllEntries(model.BandmapEntry) bool {
	return true
}

func TestEntries_Bands_SummarizesSpotsPointsAndMultis(t *testing.T) {
	now := time.Now()
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.SetBands([]model.Band{model.Band80m, model.Band40m})

	entries.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Time: now}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 3545000, Band: model.Band80m, Time: now}, now, defaultWeights)
	entries.Add(model.Spot{Call: model.MustParseCallsign("dl3abc"), Frequency: 7035000, Band: model.Band40m, Time: now}, now, defaultWeights)
	setEntryInfo(t, entries, "DL1ABC", 2, "DL")
	setEntryInfo(t, entries, "DL2ABC", 3, "DL")
	setEntryInfo(t, entries, "DL3ABC", 10, "F")

	entries.CleanOut(time.Hour, now, defaultWeights)
	bands := entries.Bands(model.Band80m, model.Band80m)

	require.Len(t, bands, 2)

	assert.Equal(t, model.Band80m, bands[0].Band)
	assert.Equal(t, 2, bands[0].SpotCount, "80m spots")
	assert.Equal(t, 5, bands[0].Points, "80m points")
	assert.Equal(t, 1, bands[0].Multis(), "80m multis")
	assert.True(t, bands[0].MaxSpots, "80m has the most spots")
	assert.False(t, bands[0].MaxPoints, "80m has not the most points")

	assert.Equal(t, model.Band40m, bands[1].Band)
	assert.Equal(t, 1, bands[1].SpotCount, "40m spots")
	assert.Equal(t, 10, bands[1].Points, "40m points")
	assert.True(t, bands[1].MaxPoints, "40m has the most points")
	assert.False(t, bands[1].MaxSpots, "40m has not the most spots")
}

func TestEntries_Bands_NoMaximumWithoutValues(t *testing.T) {
	entries := NewEntries(&Notifier{}, countAllEntries)
	entries.SetBands([]model.Band{model.Band80m, model.Band40m})

	bands := entries.Bands(model.Band80m, model.Band80m)

	require.Len(t, bands, 2)
	for _, band := range bands {
		assert.False(t, band.MaxSpots, "no maximum of spots on %s", band.Band)
		assert.False(t, band.MaxPoints, "no maximum of points on %s", band.Band)
		assert.False(t, band.MaxMultis, "no maximum of multis on %s", band.Band)
	}
}

func setEntryInfo(t *testing.T, entries *Entries, call string, points int, multiValue string) {
	t.Helper()
	for _, entry := range entries.entries {
		if entry.Call.String() != call {
			continue
		}
		entry.Info.Points = points
		entry.Info.MultiValues = map[conval.Property]string{conval.DXCCEntityProperty: multiValue}
		return
	}
	t.Fatalf("entry %s not found", call)
}
