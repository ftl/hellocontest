package bandmap

import (
	"testing"
	"time"

	"github.com/ftl/conval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/model"
)

func TestBandmap_FilterChange_DeliversAFrame(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()

	bandmap.SetFilterBand(model.FixedSpotFilterBand(model.Band20m))

	// SetView already delivered a frame with the default filter, therefore the test
	// waits for the frame that carries the new filter
	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.Filter.Band.Kind == model.SpotFilterFixed
	})
	assert.Equal(t, model.FixedSpotFilterBand(model.Band20m), frame.Filter.Band)
}

func TestBandmap_ManyFilterChanges_DoNotBlockTheGoroutine(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()

	// every change makes the filter notify the bandmap, which must not deadlock
	// with the channel that the bandmap goroutine drains itself
	for i := range 20 {
		bandmap.SetFilterBand(model.FixedSpotFilterBand(model.Bands[i%len(model.Bands)]))
		bandmap.SetFilterMode(model.FixedSpotFilterMode(model.Modes[i%len(model.Modes)]))
		bandmap.SetFilterSort(model.SpotSortColumns[i%len(model.SpotSortColumns)], i%2 == 0)
		bandmap.VFOBandChanged(model.VFO1, model.Bands[i%len(model.Bands)])
		bandmap.VFOModeChanged(model.VFO2, model.Modes[i%len(model.Modes)])
		bandmap.FocusedVFOChanged(model.VFOID(i % int(model.VFOCount)))
		bandmap.RadioChanged("radio", i%2 == 0)
	}

	waitForFrame(t, frames)

	// folding delivers no frame by design, therefore the liveness probe changes the band
	bandmap.SetFilterFolded(true)
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.Filter.Band.Kind == model.SpotFilterAll
	})
	assert.True(t, frame.Filter.Folded, "folding was stored")
}

func TestBandmap_Navigation_FollowsTheFocusedVFO(t *testing.T) {
	tt := []struct {
		desc       string
		focusedVFO model.VFOID
		wantCall   string
	}{
		{desc: "VFO1 on 80m", focusedVFO: model.VFO1, wantCall: "DL1ABC"},
		{desc: "VFO2 on 20m", focusedVFO: model.VFO2, wantCall: "DL2ABC"},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bandmap, frames := setupRunningBandmap(t)
			defer bandmap.Close()
			selections := make(chan model.BandmapEntry, 10)
			bandmap.Notify(EntrySelectedListenerFunc(func(_ model.VFOID, entry model.BandmapEntry) {
				selections <- entry
			}))

			bandmap.RadioChanged("radio", false)
			bandmap.VFOBandChanged(model.VFO1, model.Band80m)
			bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
			bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
			bandmap.VFOBandChanged(model.VFO2, model.Band20m)
			bandmap.VFOModeChanged(model.VFO2, model.ModeCW)
			bandmap.VFOFrequencyChanged(model.VFO2, 14000000)
			bandmap.FocusedVFOChanged(tc.focusedVFO)
			bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
			bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 14035000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
			waitForFrame(t, frames)

			bandmap.GotoNextEntryUp()

			select {
			case entry := <-selections:
				assert.Equal(t, tc.wantCall, entry.Call.String(), "the navigation stays on the band of the focused VFO")
			case <-time.After(2 * time.Second):
				require.Fail(t, "the bandmap selected no entry")
			}
		})
	}
}

func TestBandmap_EntryOnFrequency_IsReportedPerVFO(t *testing.T) {
	// Add does not trigger an update, therefore this test needs the update ticker
	bandmap, _ := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	onFrequency := make(chan entryOnFrequencyCall, 100)
	bandmap.Notify(&entryOnFrequencySpy{calls: onFrequency})

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.VFOFrequencyChanged(model.VFO1, 3535000)
	bandmap.VFOBandChanged(model.VFO2, model.Band20m)
	bandmap.VFOModeChanged(model.VFO2, model.ModeCW)
	bandmap.VFOFrequencyChanged(model.VFO2, 14035000)
	// the spot sits on the frequency of VFO2
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 14035000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})

	call := waitForEntryOnFrequency(t, onFrequency, model.VFO2)
	assert.Equal(t, "DL2ABC", call.entry.Call.String(), "VFO2 is on the frequency of the spot")

	calls := lastEntryOnFrequency(onFrequency)
	assert.False(t, calls[model.VFO1].available, "VFO1 is on another band")
}

func waitForEntryOnFrequency(t *testing.T, calls chan entryOnFrequencyCall, vfo model.VFOID) entryOnFrequencyCall {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case call := <-calls:
			if call.vfo == vfo && call.available {
				return call
			}
		case <-deadline:
			require.Fail(t, "the bandmap reported no entry on the frequency of the VFO")
			return entryOnFrequencyCall{}
		}
	}
}

func lastEntryOnFrequency(calls chan entryOnFrequencyCall) map[model.VFOID]entryOnFrequencyCall {
	result := make(map[model.VFOID]entryOnFrequencyCall)
	for {
		select {
		case call := <-calls:
			result[call.vfo] = call
		default:
			return result
		}
	}
}

type entryOnFrequencyCall struct {
	vfo       model.VFOID
	entry     model.BandmapEntry
	available bool
}

type entryOnFrequencySpy struct {
	calls chan entryOnFrequencyCall
}

func (s *entryOnFrequencySpy) EntryOnFrequency(vfo model.VFOID, entry model.BandmapEntry, available bool) {
	select {
	case s.calls <- entryOnFrequencyCall{vfo: vfo, entry: entry, available: available}:
	default:
	}
}

func TestBandmap_Navigation_WithoutAModeOfTheVFO(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	selections := make(chan model.BandmapEntry, 10)
	bandmap.Notify(EntrySelectedListenerFunc(func(_ model.VFOID, entry model.BandmapEntry) {
		selections <- entry
	}))

	// the rig reports no mode, which happens before the first poll and when the rig
	// answers with a mode that this application does not know
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
	waitForFrame(t, frames)

	bandmap.GotoNextEntryUp()

	select {
	case entry := <-selections:
		assert.Equal(t, "DL1ABC", entry.Call.String(), "an unknown mode restricts the navigation not")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no entry")
	}
}

func TestBandmap_BandSummaries_WithoutAModeOfTheVFO(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	// the rig reports no mode, see above
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		for _, band := range frame.Bands {
			if band.SpotCount > 0 {
				return true
			}
		}
		return false
	})

	spots := 0
	for _, band := range frame.Bands {
		spots += band.SpotCount
	}
	assert.Equal(t, 1, spots, "an unknown mode restricts the band summaries not")
}

func TestBandmap_Rows_MarkersBlendInWithTheSpotsByFrequency(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3555000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
	bandmap.AddTextMarker("beacon", 3540000, model.Band80m)
	setCQMarker(bandmap, 3570000)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 3
	})

	assert.Equal(t, []string{"beacon", "DL1ABC", "CQ"}, rowTexts(frame), "the markers stand at their frequency")

	// every row is reachable through the index, spots and markers share one ID space
	for i, row := range frame.Rows {
		index, ok := frame.IndexOf(row.ID())
		require.True(t, ok, "row %d has no index", i)
		assert.Equal(t, i, index)
	}
}

func TestBandmap_Rows_MarkersBlendInWithTheSpotsByLastSeen(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	now := time.Now()

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.SetFilterSort(model.SortSpotsByLastSeen, false)
	// the marker is created with the clock of the bandmap, which stands at now
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3555000, Band: model.Band80m, Mode: model.ModeCW, Time: now.Add(-time.Hour)})
	bandmap.AddTextMarker("beacon", 3540000, model.Band80m)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	assert.Equal(t, []string{"DL1ABC", "beacon"}, rowTexts(frame), "the older spot comes before the new marker")
}

func TestBandmap_Rows_MarkersFormABlockWithTheOtherOrders(t *testing.T) {
	for _, column := range []model.SpotSortColumn{model.SortSpotsByCallsign, model.SortSpotsByValue} {
		t.Run(column.String(), func(t *testing.T) {
			bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
			defer bandmap.Close()

			bandmap.VFOBandChanged(model.VFO1, model.Band80m)
			bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
			bandmap.SetFilterSort(column, false)
			bandmap.Add(model.Spot{Call: model.MustParseCallsign("aa1abc"), Frequency: 3555000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
			bandmap.AddTextMarker("beacon", 3540000, model.Band80m)
			setCQMarker(bandmap, 3570000)

			frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
				return len(frame.Rows) == 3
			})

			assert.Equal(t, []string{"CQ", "beacon", "AA1ABC"}, rowTexts(frame),
				"a marker has no callsign and no value, therefore the markers keep a block at the top")
		})
	}
}

func rowTexts(frame model.BandmapFrame) []string {
	result := make([]string, len(frame.Rows))
	for i, row := range frame.Rows {
		if row.Kind == model.MarkerRow {
			result[i] = row.Marker.Text
			continue
		}
		result[i] = row.Entry.Call.String()
	}
	return result
}

func TestBandmap_Rows_MarkersFollowTheBandFilter(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.SetFilterBand(model.FixedSpotFilterBand(model.Band80m))
	bandmap.AddTextMarker("eighty", 3560000, model.Band80m)
	bandmap.AddTextMarker("twenty", 14060000, model.Band20m)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) > 0 && frame.Filter.Band.Kind == model.SpotFilterFixed
	})

	require.Len(t, frame.Rows, 1, "the marker of the other band is hidden")
	assert.Equal(t, "eighty", frame.Rows[0].Marker.Text)
}

func TestBandmap_MacroSent_SetsTheCQMarker(t *testing.T) {
	tt := []struct {
		desc      string
		vfo       model.VFOID
		workmode  model.Workmode
		index     int
		wantMarks bool
	}{
		{desc: "CQ macro on VFO1 while running", vfo: model.VFO1, workmode: model.Run, index: model.CQMacroIndex, wantMarks: true},
		{desc: "CQ macro on VFO1 in search and pounce", vfo: model.VFO1, workmode: model.SearchPounce, index: model.CQMacroIndex},
		{desc: "CQ macro on VFO2 while running", vfo: model.VFO2, workmode: model.Run, index: model.CQMacroIndex},
		{desc: "another macro on VFO1 while running", vfo: model.VFO1, workmode: model.Run, index: model.CQMacroIndex + 1},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
			defer bandmap.Close()
			bandmap.VFOBandChanged(model.VFO1, model.Band80m)
			bandmap.VFOFrequencyChanged(model.VFO1, 3560000)

			bandmap.MacroSent(tc.vfo, tc.workmode, tc.index)

			if !tc.wantMarks {
				// the frame of the update ticker shows that no marker appeared
				frame := waitForFrame(t, frames)
				assert.Empty(t, frame.Rows, "only a CQ call of VFO1 marks a running frequency")
				return
			}

			frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
				return len(frame.Rows) > 0
			})
			require.Len(t, frame.Rows, 1)
			assert.Equal(t, model.MarkerRow, frame.Rows[0].Kind)
			assert.Equal(t, model.CQMarker, frame.Rows[0].Marker.Kind)
			assert.Equal(t, model.Frequency(3560000), frame.Rows[0].Marker.Frequency)
			assert.Equal(t, model.Band80m, frame.Rows[0].Marker.Band)
		})
	}
}

func TestBandmap_MacroSent_MovesTheExistingCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO1, 3560000)
	bandmap.MacroSent(model.VFO1, model.Run, model.CQMacroIndex)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) > 0
	})

	// the operator moves to another frequency and calls CQ again
	bandmap.VFOFrequencyChanged(model.VFO1, 3575000)
	bandmap.MacroSent(model.VFO1, model.Run, model.CQMacroIndex)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) > 0 && frame.Rows[0].Marker.Frequency == 3575000
	})
	assert.Len(t, frame.Rows, 1, "there is either none or only one CQ marker")
}

func TestBandmap_MacroSent_WithoutAFrequency(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	// the rig did not report a frequency yet
	bandmap.MacroSent(model.VFO1, model.Run, model.CQMacroIndex)

	frame := waitForFrame(t, frames)
	assert.Empty(t, frame.Rows, "a marker without a frequency is useless")
}

func TestBandmap_SelectMarker_UsesTheRulesOfASpot(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOBandChanged(model.VFO2, model.Band20m)
	bandmap.FocusedVFOChanged(model.VFO1)
	// the operator sees the markers of every band and clicks the one on the band of VFO2
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	bandmap.AddTextMarker("twenty", 14060000, model.Band20m)
	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) > 0
	})

	bandmap.SelectMarker(frame.Rows[0].Marker.ID)

	select {
	case selection := <-selections:
		assert.Equal(t, model.VFO2, selection.vfo, "VFO2 is already on the band of the marker")
		assert.Equal(t, model.Frequency(14060000), selection.marker.Frequency)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_Navigation_StopsAtAMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	markerSelections := make(chan model.BandmapMarker, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(_ model.VFOID, marker model.BandmapMarker) {
		markerSelections <- marker
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
	bandmap.VFOBandChanged(model.VFO2, model.Band20m)
	bandmap.FocusedVFOChanged(model.VFO1)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
	bandmap.AddTextMarker("twenty", 14020000, model.Band20m)
	bandmap.AddNumberedMarker(1, 3520000, model.Band80m)
	waitForFrame(t, frames)

	bandmap.GotoNextEntryUp()

	select {
	case marker := <-markerSelections:
		assert.Equal(t, model.Frequency(3520000), marker.Frequency, "the navigation stops at the marker on the band of the focused VFO")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_Navigation_DoesNotStopAtTheCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	selections := make(chan model.BandmapEntry, 10)
	bandmap.Notify(EntrySelectedListenerFunc(func(_ model.VFOID, entry model.BandmapEntry) {
		selections <- entry
	}))
	markerSelections := make(chan model.BandmapMarker, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(_ model.VFOID, marker model.BandmapMarker) {
		markerSelections <- marker
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.FocusedVFOChanged(model.VFO1)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})
	setCQMarker(bandmap, 3520000)
	bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
	waitForFrame(t, frames)

	bandmap.GotoNextEntryUp()

	select {
	case entry := <-selections:
		assert.Equal(t, "DL1ABC", entry.Call.String(), "the CQ frequency has its own action")
	case marker := <-markerSelections:
		require.Failf(t, "the navigation must not stop at the CQ marker", "marker: %v", marker)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected nothing")
	}
}

func TestBandmap_MarkersChanged_IsEmitted(t *testing.T) {
	bandmap, _ := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	markerSets := make(chan []model.BandmapMarker, 100)
	bandmap.Notify(MarkersChangedListenerFunc(func(markers []model.BandmapMarker) {
		select {
		case markerSets <- markers:
		default:
		}
	}))

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.AddNumberedMarker(1, 3540000, model.Band80m)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case markers := <-markerSets:
			if len(markers) == 0 {
				continue
			}
			assert.Equal(t, model.Frequency(3540000), markers[0].Frequency)
			return
		case <-deadline:
			require.Fail(t, "the bandmap emitted no markers")
		}
	}
}

func TestBandmap_DeleteMarkerOnFrequency(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO1, 3550100)
	bandmap.AddNumberedMarker(1, 3550000, model.Band80m)
	bandmap.AddTextMarker("beacon", 3500000, model.Band80m)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.DeleteMarkerOnFrequency()

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})
	assert.Equal(t, "beacon", frame.Rows[0].Marker.Text, "the marker in proximity is deleted")
}

func TestBandmap_DeleteMarkerOnFrequency_WithoutAMarkerInProximity(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
	bandmap.AddNumberedMarker(1, 3550000, model.Band80m)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})

	bandmap.DeleteMarkerOnFrequency()

	select {
	case frame := <-frames:
		assert.Len(t, frame.Rows, 1, "no marker in proximity, nothing happens")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBandmap_Remove(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	spot := model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()}
	bandmap.Add(spot)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})

	bandmap.Remove(spot)

	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 0
	})
}

func TestBandmap_GotoNumberedMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOBandChanged(model.VFO2, model.Band20m)
	bandmap.FocusedVFOChanged(model.VFO1)
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	bandmap.AddNumberedMarker(1, 3560000, model.Band80m)
	bandmap.AddNumberedMarker(2, 14060000, model.Band20m)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.GotoNumberedMarker(2)

	select {
	case selection := <-selections:
		assert.Equal(t, model.Frequency(14060000), selection.marker.Frequency, "the number selects the marker")
		assert.Equal(t, model.VFO2, selection.vfo, "the marker uses the rules of a spot")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_GotoNumberedMarker_WithoutSuchAMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, model.Band80m)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})

	bandmap.GotoNumberedMarker(3)

	select {
	case <-selections:
		require.Fail(t, "there is no marker with this number")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBandmap_GotoNumberedMarker_OnTheBandOfTheOtherVFO(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOBandChanged(model.VFO2, model.Band20m)
	bandmap.FocusedVFOChanged(model.VFO1)
	// the spot list shows only the band of the focused VFO, the action finds the marker
	// nevertheless
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterFocused})
	bandmap.AddNumberedMarker(1, 14060000, model.Band20m)
	waitForFrame(t, frames)

	bandmap.GotoNumberedMarker(1)

	select {
	case selection := <-selections:
		assert.Equal(t, model.VFO2, selection.vfo, "VFO2 is already on the band of the marker, VFO1 keeps its band")
		assert.Equal(t, model.Frequency(14060000), selection.marker.Frequency)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_GotoCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	cqMarkers := make(chan model.BandmapMarker, 10)
	bandmap.Notify(CQMarkerSelectedListenerFunc(func(marker model.BandmapMarker) {
		cqMarkers <- marker
	}))
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, model.Band80m)
	setCQMarker(bandmap, 3560000)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.GotoCQMarker()

	select {
	case marker := <-cqMarkers:
		assert.Equal(t, model.CQMarker, marker.Kind, "the generic marker is no CQ frequency")
		assert.Equal(t, model.Frequency(3560000), marker.Frequency)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no CQ marker")
	}

	select {
	case <-selections:
		require.Fail(t, "the CQ frequency has its own rules, it is no ordinary marker selection")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBandmap_GotoCQMarker_WithoutACQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	cqMarkers := make(chan model.BandmapMarker, 10)
	bandmap.Notify(CQMarkerSelectedListenerFunc(func(marker model.BandmapMarker) {
		cqMarkers <- marker
	}))

	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, model.Band80m)
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})

	bandmap.GotoCQMarker()

	select {
	case <-cqMarkers:
		require.Fail(t, "without a CQ marker there is nothing to go to")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBandmap_SelectMarker_IgnoresAnUnknownID(t *testing.T) {
	bandmap, _ := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo model.VFOID, marker model.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.SelectMarker(4711)

	select {
	case <-selections:
		require.Fail(t, "an unknown ID selects nothing")
	case <-time.After(100 * time.Millisecond):
	}
}

type markerSelection struct {
	vfo    model.VFOID
	marker model.BandmapMarker
}

func TestBandmap_MarkerOnFrequency_IsReportedPerVFO(t *testing.T) {
	bandmap, _ := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	calls := make(chan markerOnFrequencyCall, 100)
	bandmap.Notify(&markerOnFrequencySpy{calls: calls})

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO1, 3500000)
	bandmap.VFOBandChanged(model.VFO2, model.Band80m)
	bandmap.VFOFrequencyChanged(model.VFO2, 3560000)
	// the marker sits on the frequency of VFO2
	bandmap.AddTextMarker("beacon", 3560100, model.Band80m)

	call := waitForMarkerOnFrequency(t, calls, model.VFO2)
	assert.Equal(t, "beacon", call.marker.Text)

	last := lastMarkerOnFrequency(calls)
	assert.False(t, last[model.VFO1].available, "VFO1 is far away from the marker")
}

func waitForMarkerOnFrequency(t *testing.T, calls chan markerOnFrequencyCall, vfo model.VFOID) markerOnFrequencyCall {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case call := <-calls:
			if call.vfo == vfo && call.available {
				return call
			}
		case <-deadline:
			require.Fail(t, "the bandmap reported no marker on the frequency of the VFO")
			return markerOnFrequencyCall{}
		}
	}
}

func lastMarkerOnFrequency(calls chan markerOnFrequencyCall) map[model.VFOID]markerOnFrequencyCall {
	result := make(map[model.VFOID]markerOnFrequencyCall)
	for {
		select {
		case call := <-calls:
			result[call.vfo] = call
		default:
			return result
		}
	}
}

type markerOnFrequencyCall struct {
	vfo       model.VFOID
	marker    model.BandmapMarker
	available bool
}

type markerOnFrequencySpy struct {
	calls chan markerOnFrequencyCall
}

func (s *markerOnFrequencySpy) MarkerOnFrequency(vfo model.VFOID, marker model.BandmapMarker, available bool) {
	select {
	case s.calls <- markerOnFrequencyCall{vfo: vfo, marker: marker, available: available}:
	default:
	}
}

// setCQMarker places the CQ marker the way the application does it: the operator calls CQ
// on VFO1 while running.
func setCQMarker(bandmap *Bandmap, frequency model.Frequency) {
	bandmap.VFOFrequencyChanged(model.VFO1, frequency)
	bandmap.MacroSent(model.VFO1, model.Run, model.CQMacroIndex)
}

func setupRunningBandmap(t *testing.T) (*Bandmap, chan model.BandmapFrame) {
	t.Helper()
	return setupRunningBandmapWithPeriod(t, time.Hour)
}

func setupRunningBandmapWithPeriod(t *testing.T, updatePeriod time.Duration) (*Bandmap, chan model.BandmapFrame) {
	t.Helper()
	frames := make(chan model.BandmapFrame, 1000)
	bandmap := NewBandmap(
		staticClock{now: time.Now()},
		new(settingsStub),
		new(dupeCheckerStub),
		newSpotFilterStoreStub(),
		func(f func()) { f() },
		updatePeriod,
		time.Hour,
	)
	bandmap.SetView(&bandmapViewStub{frames: frames})
	return bandmap, frames
}

func waitForFrame(t *testing.T, frames chan model.BandmapFrame) model.BandmapFrame {
	t.Helper()
	return waitForFrameMatching(t, frames, func(model.BandmapFrame) bool { return true })
}

func waitForFrameMatching(t *testing.T, frames chan model.BandmapFrame, matches func(model.BandmapFrame) bool) model.BandmapFrame {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case frame := <-frames:
			if matches(frame) {
				return frame
			}
		case <-deadline:
			require.Fail(t, "the bandmap delivered no matching frame, its goroutine is blocked")
			return model.BandmapFrame{}
		}
	}
}

type staticClock struct {
	now time.Time
}

func (c staticClock) Now() time.Time {
	return c.now
}

type settingsStub struct{}

func (s *settingsStub) Station() model.Station {
	return model.Station{}
}

func (s *settingsStub) Contest() model.Contest {
	// the band summaries only know the bands of the contest
	return model.Contest{
		Definition: &conval.Definition{
			Bands: []conval.ContestBand{"80m", "40m", "20m"},
		},
	}
}

type dupeCheckerStub struct{}

func (d *dupeCheckerStub) FindWorkedQSOs(model.Callsign, model.Band, model.Mode) ([]model.QSO, bool) {
	return nil, false
}

type bandmapViewStub struct {
	frames chan model.BandmapFrame
}

func (v *bandmapViewStub) Show() {}
func (v *bandmapViewStub) Hide() {}

func (v *bandmapViewStub) ShowFrame(frame model.BandmapFrame) {
	select {
	case v.frames <- frame:
	default:
	}
}

// callinfoStub reports the value that the test sets, like the collector does after a QSO
// was logged and the spot became a duplicate.
type callinfoStub struct {
	value     int
	multis    int
	duplicate bool
}

func (s *callinfoStub) GetInfo(call model.Callsign, _ model.Band, _ model.Mode, _ []string) model.Callinfo {
	return model.Callinfo{Call: call, Value: s.value, Multis: s.multis, Duplicate: s.duplicate}
}

func (s *callinfoStub) UpdateValue(info *model.Callinfo, _ model.Band, _ model.Mode) bool {
	info.Value = s.value
	info.Multis = s.multis
	info.Duplicate = s.duplicate
	return true
}

func TestBandmap_ValueFilter_DropsASpotThatLostItsValue(t *testing.T) {
	// the periodic update recomputes the values of the entries
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	callinfo := &callinfoStub{value: 40}
	bandmap.SetCallinfo(callinfo)
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	bandmap.SetFilterMode(model.SpotFilterMode{Kind: model.SpotFilterAll})
	bandmap.SetFilterValue(model.SpotFilterScoring)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: model.Band80m, Mode: model.ModeCW, Time: time.Now()})

	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 1 })
	require.Len(t, frame.Rows, 1, "the spot scores, therefore the table shows it")

	// the QSO is logged: the contest rules still name points for the QSO, but the duplicate
	// adds nothing to the score
	callinfo.duplicate = true

	frame = waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 0 })
	assert.Empty(t, frame.Rows, "the spot lost its value, therefore the table shows it no more")
}

func TestBandmap_GotoNearestEntry_KeepsTheFocusedVFOActive(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	selections := make(chan model.VFOID, 10)
	bandmap.Notify(EntrySelectedListenerFunc(func(vfo model.VFOID, _ model.BandmapEntry) {
		selections <- vfo
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	bandmap.SetFilterMode(model.SpotFilterMode{Kind: model.SpotFilterAll})
	bandmap.VFOBandChanged(model.VFO1, model.Band20m)
	bandmap.VFOFrequencyChanged(model.VFO1, 14023000)
	bandmap.VFOBandChanged(model.VFO2, model.Band40m)
	bandmap.VFOFrequencyChanged(model.VFO2, 7030000)
	bandmap.FocusedVFOChanged(model.VFO2)
	// the table shows no spot on the band of VFO2, therefore the nearest spot belongs to
	// the band of VFO1
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 14023000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 14025000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
	waitForFrame(t, frames)

	bandmap.GotoNearestEntry(7030000)

	select {
	case vfo := <-selections:
		assert.Equal(t, model.VFO2, vfo, "the focused VFO goes to the nearest spot, it stays the active one")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no entry")
	}
}

var _ model.VFOTunedListener = (*Bandmap)(nil)

func TestBandmap_FocusedVFOChanged_ClearsTheSelection(t *testing.T) {
	bandmap, frames := setupBandmapWithSelectedSpot(t)
	defer bandmap.Close()

	bandmap.FocusedVFOChanged(model.VFO2)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.SelectedEntry.ID == model.NoEntryID
	})
	assert.Equal(t, model.NoEntryID, frame.SelectedEntry.ID, "the operator works with another VFO now")
}

func TestBandmap_VFOTuned_WithABigJump_ClearsTheSelection(t *testing.T) {
	bandmap, frames := setupBandmapWithSelectedSpot(t)
	defer bandmap.Close()

	bandmap.VFOTuned(model.VFO1, 14030000)

	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.SelectedEntry.ID == model.NoEntryID
	})
	assert.Equal(t, model.NoEntryID, frame.SelectedEntry.ID, "the operator left the frequency of the selected row")
}

func TestBandmap_VFOTuned_WithASmallJump_KeepsTheSelection(t *testing.T) {
	bandmap, frames := setupBandmapWithSelectedSpot(t)
	defer bandmap.Close()

	bandmap.VFOTuned(model.VFO1, 14020100)

	// the tuning alone delivers no frame, therefore a filter change asks for one
	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.Filter.Band.Kind == model.SpotFilterAll
	})
	assert.NotEqual(t, model.NoEntryID, frame.SelectedEntry.ID, "the operator stays on the frequency of the selected row")
}

func TestBandmap_VFOTuned_OfAnotherVFO_KeepsTheSelection(t *testing.T) {
	bandmap, frames := setupBandmapWithSelectedSpot(t)
	defer bandmap.Close()

	bandmap.VFOTuned(model.VFO2, 7030000)

	bandmap.SetFilterBand(model.SpotFilterBand{Kind: model.SpotFilterAll})
	frame := waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.Filter.Band.Kind == model.SpotFilterAll
	})
	assert.NotEqual(t, model.NoEntryID, frame.SelectedEntry.ID, "the selection belongs to the focused VFO")
}

func setupBandmapWithSelectedSpot(t *testing.T) (*Bandmap, chan model.BandmapFrame) {
	t.Helper()
	bandmap, frames := setupRunningBandmap(t)
	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(model.VFO1, model.Band20m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.VFOFrequencyChanged(model.VFO1, 14020000)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 14023000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
	bandmap.SelectByCallsign(model.MustParseCallsign("dl1abc"))
	waitForFrameMatching(t, frames, func(frame model.BandmapFrame) bool {
		return frame.SelectedEntry.ID != model.NoEntryID
	})
	return bandmap, frames
}

func TestBandmap_IgnoreEntry_TakesTheValueOfTheSpot(t *testing.T) {
	bandmap, frames := setupBandmapWithSpot(t)
	defer bandmap.Close()
	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && f.Rows[0].Entry.Info.Value > 0
	})

	bandmap.IgnoreEntry(frame.Rows[0].Entry.ID)

	frame = waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && f.Rows[0].Entry.Ignored
	})
	entry := frame.Rows[0].Entry
	assert.Zero(t, entry.Info.Value, "an ignored spot has no value")
	assert.Zero(t, entry.Info.Points, "an ignored spot scores no points")
	assert.Zero(t, entry.Info.Multis, "an ignored spot is no multi")
	assert.Zero(t, entry.Info.WeightedValue, "an ignored spot has no weight")
}

func TestBandmap_RestoreEntry_GivesTheValueBack(t *testing.T) {
	bandmap, frames := setupBandmapWithSpot(t)
	defer bandmap.Close()
	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 1 })
	id := frame.Rows[0].Entry.ID
	bandmap.IgnoreEntry(id)
	waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && f.Rows[0].Entry.Ignored
	})

	bandmap.RestoreEntry(id)

	frame = waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && !f.Rows[0].Entry.Ignored
	})
	assert.Positive(t, frame.Rows[0].Entry.Info.Value, "the spot has its value again")
}

func TestBandmap_IgnoreEntryOnFrequency_TakesTheSpotOfTheFocusedVFO(t *testing.T) {
	bandmap, frames := setupBandmapWithSpot(t)
	defer bandmap.Close()
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl2abc"), Frequency: 14050000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
	waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 2 })

	bandmap.IgnoreEntryOnFrequency()

	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 2 && f.Rows[0].Entry.Ignored
	})
	assert.Equal(t, "DL1ABC", frame.Rows[0].Entry.Call.String(), "the VFO is on this spot")
	assert.False(t, frame.Rows[1].Entry.Ignored, "the other spot keeps its value")
}

func TestBandmap_IgnoredSpot_DisappearsWithTheValueFilter(t *testing.T) {
	bandmap, frames := setupBandmapWithSpot(t)
	defer bandmap.Close()
	bandmap.SetFilterValue(model.SpotFilterScoring)
	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 1 })

	bandmap.IgnoreEntry(frame.Rows[0].Entry.ID)

	frame = waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 0 })
	assert.Empty(t, frame.Rows, "the ignored spot has no value, therefore the filter hides it")
}

func setupBandmapWithSpot(t *testing.T) (*Bandmap, chan model.BandmapFrame) {
	t.Helper()
	// the periodic update recomputes the values of the entries
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	bandmap.SetCallinfo(&callinfoStub{value: 40, multis: 1})
	bandmap.VFOBandChanged(model.VFO1, model.Band20m)
	bandmap.VFOModeChanged(model.VFO1, model.ModeCW)
	bandmap.VFOFrequencyChanged(model.VFO1, 14023000)
	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 14023000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})
	return bandmap, frames
}

func TestBandmap_IgnoredEntry_StaysIgnoredWhenANewSpotArrives(t *testing.T) {
	bandmap, frames := setupBandmapWithSpot(t)
	defer bandmap.Close()
	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 1 })
	bandmap.IgnoreEntry(frame.Rows[0].Entry.ID)
	waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && f.Rows[0].Entry.Ignored
	})

	bandmap.Add(model.Spot{Call: model.MustParseCallsign("dl1abc"), Frequency: 14023000, Band: model.Band20m, Mode: model.ModeCW, Time: time.Now()})

	frame = waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool {
		return len(f.Rows) == 1 && f.Rows[0].Entry.SpotCount == 2
	})
	assert.True(t, frame.Rows[0].Entry.Ignored, "another spot of the same station changes nothing")
	assert.Zero(t, frame.Rows[0].Entry.Info.Value, "the spot stays without value")
}

func TestBandmap_MarkWithNumber_StoresTheInputInTheMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	input := model.MarkerInput{Call: "DL1ABC", TheirExchange: []string{"599", "012"}}

	bandmap.MarkWithNumber(5, 3550000, model.Band80m, input)

	frame := waitForFrameMatching(t, frames, func(f model.BandmapFrame) bool { return len(f.Rows) == 1 })
	marker := frame.Rows[0].Marker
	assert.Equal(t, 5, marker.Number)
	assert.Equal(t, input, marker.Input, "the operator continues this QSO at the marker")
}
