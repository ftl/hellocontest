package bandmap

import (
	"testing"
	"time"

	"github.com/ftl/conval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestBandmap_FilterChange_DeliversAFrame(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()

	bandmap.SetFilterBand(core.FixedSpotFilterBand(core.Band20m))

	// SetView already delivered a frame with the default filter, therefore the test
	// waits for the frame that carries the new filter
	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return frame.Filter.Band.Kind == core.SpotFilterFixed
	})
	assert.Equal(t, core.FixedSpotFilterBand(core.Band20m), frame.Filter.Band)
}

func TestBandmap_ManyFilterChanges_DoNotBlockTheGoroutine(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()

	// every change makes the filter notify the bandmap, which must not deadlock
	// with the channel that the bandmap goroutine drains itself
	for i := range 20 {
		bandmap.SetFilterBand(core.FixedSpotFilterBand(core.Bands[i%len(core.Bands)]))
		bandmap.SetFilterMode(core.FixedSpotFilterMode(core.Modes[i%len(core.Modes)]))
		bandmap.SetFilterSort(core.SpotSortColumns[i%len(core.SpotSortColumns)], i%2 == 0)
		bandmap.VFOBandChanged(core.VFO1, core.Bands[i%len(core.Bands)])
		bandmap.VFOModeChanged(core.VFO2, core.Modes[i%len(core.Modes)])
		bandmap.FocusedVFOChanged(core.VFOID(i % int(core.VFOCount)))
		bandmap.RadioChanged("radio", i%2 == 0)
	}

	waitForFrame(t, frames)

	// folding delivers no frame by design, therefore the liveness probe changes the band
	bandmap.SetFilterFolded(true)
	bandmap.SetFilterBand(core.SpotFilterBand{Kind: core.SpotFilterAll})
	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return frame.Filter.Band.Kind == core.SpotFilterAll
	})
	assert.True(t, frame.Filter.Folded, "folding was stored")
}

func TestBandmap_Navigation_FollowsTheFocusedVFO(t *testing.T) {
	tt := []struct {
		desc       string
		focusedVFO core.VFOID
		wantCall   string
	}{
		{desc: "VFO1 on 80m", focusedVFO: core.VFO1, wantCall: "DL1ABC"},
		{desc: "VFO2 on 20m", focusedVFO: core.VFO2, wantCall: "DL2ABC"},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bandmap, frames := setupRunningBandmap(t)
			defer bandmap.Close()
			selections := make(chan core.BandmapEntry, 10)
			bandmap.Notify(EntrySelectedListenerFunc(func(_ core.VFOID, entry core.BandmapEntry) {
				selections <- entry
			}))

			bandmap.RadioChanged("radio", false)
			bandmap.VFOBandChanged(core.VFO1, core.Band80m)
			bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
			bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
			bandmap.VFOBandChanged(core.VFO2, core.Band20m)
			bandmap.VFOModeChanged(core.VFO2, core.ModeCW)
			bandmap.VFOFrequencyChanged(core.VFO2, 14000000)
			bandmap.FocusedVFOChanged(tc.focusedVFO)
			bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
			bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl2abc"), Frequency: 14035000, Band: core.Band20m, Mode: core.ModeCW, Time: time.Now()})
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
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	bandmap.VFOFrequencyChanged(core.VFO1, 3535000)
	bandmap.VFOBandChanged(core.VFO2, core.Band20m)
	bandmap.VFOModeChanged(core.VFO2, core.ModeCW)
	bandmap.VFOFrequencyChanged(core.VFO2, 14035000)
	// the spot sits on the frequency of VFO2
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl2abc"), Frequency: 14035000, Band: core.Band20m, Mode: core.ModeCW, Time: time.Now()})

	call := waitForEntryOnFrequency(t, onFrequency, core.VFO2)
	assert.Equal(t, "DL2ABC", call.entry.Call.String(), "VFO2 is on the frequency of the spot")

	calls := lastEntryOnFrequency(onFrequency)
	assert.False(t, calls[core.VFO1].available, "VFO1 is on another band")
}

func waitForEntryOnFrequency(t *testing.T, calls chan entryOnFrequencyCall, vfo core.VFOID) entryOnFrequencyCall {
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

func lastEntryOnFrequency(calls chan entryOnFrequencyCall) map[core.VFOID]entryOnFrequencyCall {
	result := make(map[core.VFOID]entryOnFrequencyCall)
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
	vfo       core.VFOID
	entry     core.BandmapEntry
	available bool
}

type entryOnFrequencySpy struct {
	calls chan entryOnFrequencyCall
}

func (s *entryOnFrequencySpy) EntryOnFrequency(vfo core.VFOID, entry core.BandmapEntry, available bool) {
	select {
	case s.calls <- entryOnFrequencyCall{vfo: vfo, entry: entry, available: available}:
	default:
	}
}

func TestBandmap_Navigation_WithoutAModeOfTheVFO(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	selections := make(chan core.BandmapEntry, 10)
	bandmap.Notify(EntrySelectedListenerFunc(func(_ core.VFOID, entry core.BandmapEntry) {
		selections <- entry
	}))

	// the rig reports no mode, which happens before the first poll and when the rig
	// answers with a mode that this application does not know
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
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
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
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

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3555000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
	bandmap.AddTextMarker("beacon", 3540000, core.Band80m)
	setCQMarker(bandmap, 3570000)

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
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

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	bandmap.SetFilterSort(core.SortSpotsByLastSeen, false)
	// the marker is created with the clock of the bandmap, which stands at now
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3555000, Band: core.Band80m, Mode: core.ModeCW, Time: now.Add(-time.Hour)})
	bandmap.AddTextMarker("beacon", 3540000, core.Band80m)

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	assert.Equal(t, []string{"DL1ABC", "beacon"}, rowTexts(frame), "the older spot comes before the new marker")
}

func TestBandmap_Rows_MarkersFormABlockWithTheOtherOrders(t *testing.T) {
	for _, column := range []core.SpotSortColumn{core.SortSpotsByCallsign, core.SortSpotsByValue} {
		t.Run(column.String(), func(t *testing.T) {
			bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
			defer bandmap.Close()

			bandmap.VFOBandChanged(core.VFO1, core.Band80m)
			bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
			bandmap.SetFilterSort(column, false)
			bandmap.Add(core.Spot{Call: core.MustParseCallsign("aa1abc"), Frequency: 3555000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
			bandmap.AddTextMarker("beacon", 3540000, core.Band80m)
			setCQMarker(bandmap, 3570000)

			frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
				return len(frame.Rows) == 3
			})

			assert.Equal(t, []string{"CQ", "beacon", "AA1ABC"}, rowTexts(frame),
				"a marker has no callsign and no value, therefore the markers keep a block at the top")
		})
	}
}

func rowTexts(frame core.BandmapFrame) []string {
	result := make([]string, len(frame.Rows))
	for i, row := range frame.Rows {
		if row.Kind == core.MarkerRow {
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

	bandmap.SetFilterBand(core.FixedSpotFilterBand(core.Band80m))
	bandmap.AddTextMarker("eighty", 3560000, core.Band80m)
	bandmap.AddTextMarker("twenty", 14060000, core.Band20m)

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) > 0 && frame.Filter.Band.Kind == core.SpotFilterFixed
	})

	require.Len(t, frame.Rows, 1, "the marker of the other band is hidden")
	assert.Equal(t, "eighty", frame.Rows[0].Marker.Text)
}

func TestBandmap_MacroSent_SetsTheCQMarker(t *testing.T) {
	tt := []struct {
		desc      string
		vfo       core.VFOID
		workmode  core.Workmode
		index     int
		wantMarks bool
	}{
		{desc: "CQ macro on VFO1 while running", vfo: core.VFO1, workmode: core.Run, index: core.CQMacroIndex, wantMarks: true},
		{desc: "CQ macro on VFO1 in search and pounce", vfo: core.VFO1, workmode: core.SearchPounce, index: core.CQMacroIndex},
		{desc: "CQ macro on VFO2 while running", vfo: core.VFO2, workmode: core.Run, index: core.CQMacroIndex},
		{desc: "another macro on VFO1 while running", vfo: core.VFO1, workmode: core.Run, index: core.CQMacroIndex + 1},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
			defer bandmap.Close()
			bandmap.VFOBandChanged(core.VFO1, core.Band80m)
			bandmap.VFOFrequencyChanged(core.VFO1, 3560000)

			bandmap.MacroSent(tc.vfo, tc.workmode, tc.index)

			if !tc.wantMarks {
				// the frame of the update ticker shows that no marker appeared
				frame := waitForFrame(t, frames)
				assert.Empty(t, frame.Rows, "only a CQ call of VFO1 marks a running frequency")
				return
			}

			frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
				return len(frame.Rows) > 0
			})
			require.Len(t, frame.Rows, 1)
			assert.Equal(t, core.MarkerRow, frame.Rows[0].Kind)
			assert.Equal(t, core.CQMarker, frame.Rows[0].Marker.Kind)
			assert.Equal(t, core.Frequency(3560000), frame.Rows[0].Marker.Frequency)
			assert.Equal(t, core.Band80m, frame.Rows[0].Marker.Band)
		})
	}
}

func TestBandmap_MacroSent_MovesTheExistingCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO1, 3560000)
	bandmap.MacroSent(core.VFO1, core.Run, core.CQMacroIndex)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) > 0
	})

	// the operator moves to another frequency and calls CQ again
	bandmap.VFOFrequencyChanged(core.VFO1, 3575000)
	bandmap.MacroSent(core.VFO1, core.Run, core.CQMacroIndex)

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) > 0 && frame.Rows[0].Marker.Frequency == 3575000
	})
	assert.Len(t, frame.Rows, 1, "there is either none or only one CQ marker")
}

func TestBandmap_MacroSent_WithoutAFrequency(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	// the rig did not report a frequency yet
	bandmap.MacroSent(core.VFO1, core.Run, core.CQMacroIndex)

	frame := waitForFrame(t, frames)
	assert.Empty(t, frame.Rows, "a marker without a frequency is useless")
}

func TestBandmap_SelectMarker_UsesTheRulesOfASpot(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOBandChanged(core.VFO2, core.Band20m)
	bandmap.FocusedVFOChanged(core.VFO1)
	// the operator sees the markers of every band and clicks the one on the band of VFO2
	bandmap.SetFilterBand(core.SpotFilterBand{Kind: core.SpotFilterAll})
	bandmap.AddTextMarker("twenty", 14060000, core.Band20m)
	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) > 0
	})

	bandmap.SelectMarker(frame.Rows[0].Marker.ID)

	select {
	case selection := <-selections:
		assert.Equal(t, core.VFO2, selection.vfo, "VFO2 is already on the band of the marker")
		assert.Equal(t, core.Frequency(14060000), selection.marker.Frequency)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_Navigation_StopsAtAMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	markerSelections := make(chan core.BandmapMarker, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(_ core.VFOID, marker core.BandmapMarker) {
		markerSelections <- marker
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
	bandmap.VFOBandChanged(core.VFO2, core.Band20m)
	bandmap.FocusedVFOChanged(core.VFO1)
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
	bandmap.AddTextMarker("twenty", 14020000, core.Band20m)
	bandmap.AddNumberedMarker(1, 3520000, core.Band80m)
	waitForFrame(t, frames)

	bandmap.GotoNextEntryUp()

	select {
	case marker := <-markerSelections:
		assert.Equal(t, core.Frequency(3520000), marker.Frequency, "the navigation stops at the marker on the band of the focused VFO")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_Navigation_DoesNotStopAtTheCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()
	selections := make(chan core.BandmapEntry, 10)
	bandmap.Notify(EntrySelectedListenerFunc(func(_ core.VFOID, entry core.BandmapEntry) {
		selections <- entry
	}))
	markerSelections := make(chan core.BandmapMarker, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(_ core.VFOID, marker core.BandmapMarker) {
		markerSelections <- marker
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	bandmap.FocusedVFOChanged(core.VFO1)
	bandmap.Add(core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()})
	setCQMarker(bandmap, 3520000)
	bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
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
	markerSets := make(chan []core.BandmapMarker, 100)
	bandmap.Notify(MarkersChangedListenerFunc(func(markers []core.BandmapMarker) {
		select {
		case markerSets <- markers:
		default:
		}
	}))

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.AddNumberedMarker(1, 3540000, core.Band80m)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case markers := <-markerSets:
			if len(markers) == 0 {
				continue
			}
			assert.Equal(t, core.Frequency(3540000), markers[0].Frequency)
			return
		case <-deadline:
			require.Fail(t, "the bandmap emitted no markers")
		}
	}
}

func TestBandmap_DeleteMarkerOnFrequency(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO1, 3550100)
	bandmap.AddNumberedMarker(1, 3550000, core.Band80m)
	bandmap.AddTextMarker("beacon", 3500000, core.Band80m)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.DeleteMarkerOnFrequency()

	frame := waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})
	assert.Equal(t, "beacon", frame.Rows[0].Marker.Text, "the marker in proximity is deleted")
}

func TestBandmap_DeleteMarkerOnFrequency_WithoutAMarkerInProximity(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
	bandmap.AddNumberedMarker(1, 3550000, core.Band80m)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
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

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOModeChanged(core.VFO1, core.ModeCW)
	spot := core.Spot{Call: core.MustParseCallsign("dl1abc"), Frequency: 3535000, Band: core.Band80m, Mode: core.ModeCW, Time: time.Now()}
	bandmap.Add(spot)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 1
	})

	bandmap.Remove(spot)

	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 0
	})
}

func TestBandmap_GotoNumberedMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOBandChanged(core.VFO2, core.Band20m)
	bandmap.FocusedVFOChanged(core.VFO1)
	bandmap.SetFilterBand(core.SpotFilterBand{Kind: core.SpotFilterAll})
	bandmap.AddNumberedMarker(1, 3560000, core.Band80m)
	bandmap.AddNumberedMarker(2, 14060000, core.Band20m)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.GotoNumberedMarker(2)

	select {
	case selection := <-selections:
		assert.Equal(t, core.Frequency(14060000), selection.marker.Frequency, "the number selects the marker")
		assert.Equal(t, core.VFO2, selection.vfo, "the marker uses the rules of a spot")
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_GotoNumberedMarker_WithoutSuchAMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, core.Band80m)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
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
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOBandChanged(core.VFO2, core.Band20m)
	bandmap.FocusedVFOChanged(core.VFO1)
	// the spot list shows only the band of the focused VFO, the action finds the marker
	// nevertheless
	bandmap.SetFilterBand(core.SpotFilterBand{Kind: core.SpotFilterFocused})
	bandmap.AddNumberedMarker(1, 14060000, core.Band20m)
	waitForFrame(t, frames)

	bandmap.GotoNumberedMarker(1)

	select {
	case selection := <-selections:
		assert.Equal(t, core.VFO2, selection.vfo, "VFO2 is already on the band of the marker, VFO1 keeps its band")
		assert.Equal(t, core.Frequency(14060000), selection.marker.Frequency)
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap selected no marker")
	}
}

func TestBandmap_GotoCQMarker(t *testing.T) {
	bandmap, frames := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	cqMarkers := make(chan core.BandmapMarker, 10)
	bandmap.Notify(CQMarkerSelectedListenerFunc(func(marker core.BandmapMarker) {
		cqMarkers <- marker
	}))
	selections := make(chan markerSelection, 10)
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
		selections <- markerSelection{vfo: vfo, marker: marker}
	}))

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, core.Band80m)
	setCQMarker(bandmap, 3560000)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
		return len(frame.Rows) == 2
	})

	bandmap.GotoCQMarker()

	select {
	case marker := <-cqMarkers:
		assert.Equal(t, core.CQMarker, marker.Kind, "the generic marker is no CQ frequency")
		assert.Equal(t, core.Frequency(3560000), marker.Frequency)
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
	cqMarkers := make(chan core.BandmapMarker, 10)
	bandmap.Notify(CQMarkerSelectedListenerFunc(func(marker core.BandmapMarker) {
		cqMarkers <- marker
	}))

	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.AddTextMarker("beacon", 3540000, core.Band80m)
	waitForFrameMatching(t, frames, func(frame core.BandmapFrame) bool {
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
	bandmap.Notify(MarkerSelectedListenerFunc(func(vfo core.VFOID, marker core.BandmapMarker) {
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
	vfo    core.VFOID
	marker core.BandmapMarker
}

func TestBandmap_MarkerOnFrequency_IsReportedPerVFO(t *testing.T) {
	bandmap, _ := setupRunningBandmapWithPeriod(t, 10*time.Millisecond)
	defer bandmap.Close()
	calls := make(chan markerOnFrequencyCall, 100)
	bandmap.Notify(&markerOnFrequencySpy{calls: calls})

	bandmap.RadioChanged("radio", false)
	bandmap.VFOBandChanged(core.VFO1, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO1, 3500000)
	bandmap.VFOBandChanged(core.VFO2, core.Band80m)
	bandmap.VFOFrequencyChanged(core.VFO2, 3560000)
	// the marker sits on the frequency of VFO2
	bandmap.AddTextMarker("beacon", 3560100, core.Band80m)

	call := waitForMarkerOnFrequency(t, calls, core.VFO2)
	assert.Equal(t, "beacon", call.marker.Text)

	last := lastMarkerOnFrequency(calls)
	assert.False(t, last[core.VFO1].available, "VFO1 is far away from the marker")
}

func waitForMarkerOnFrequency(t *testing.T, calls chan markerOnFrequencyCall, vfo core.VFOID) markerOnFrequencyCall {
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

func lastMarkerOnFrequency(calls chan markerOnFrequencyCall) map[core.VFOID]markerOnFrequencyCall {
	result := make(map[core.VFOID]markerOnFrequencyCall)
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
	vfo       core.VFOID
	marker    core.BandmapMarker
	available bool
}

type markerOnFrequencySpy struct {
	calls chan markerOnFrequencyCall
}

func (s *markerOnFrequencySpy) MarkerOnFrequency(vfo core.VFOID, marker core.BandmapMarker, available bool) {
	select {
	case s.calls <- markerOnFrequencyCall{vfo: vfo, marker: marker, available: available}:
	default:
	}
}

// setCQMarker places the CQ marker the way the application does it: the operator calls CQ
// on VFO1 while running.
func setCQMarker(bandmap *Bandmap, frequency core.Frequency) {
	bandmap.VFOFrequencyChanged(core.VFO1, frequency)
	bandmap.MacroSent(core.VFO1, core.Run, core.CQMacroIndex)
}

func setupRunningBandmap(t *testing.T) (*Bandmap, chan core.BandmapFrame) {
	t.Helper()
	return setupRunningBandmapWithPeriod(t, time.Hour)
}

func setupRunningBandmapWithPeriod(t *testing.T, updatePeriod time.Duration) (*Bandmap, chan core.BandmapFrame) {
	t.Helper()
	frames := make(chan core.BandmapFrame, 1000)
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

func waitForFrame(t *testing.T, frames chan core.BandmapFrame) core.BandmapFrame {
	t.Helper()
	return waitForFrameMatching(t, frames, func(core.BandmapFrame) bool { return true })
}

func waitForFrameMatching(t *testing.T, frames chan core.BandmapFrame, matches func(core.BandmapFrame) bool) core.BandmapFrame {
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
			return core.BandmapFrame{}
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

func (s *settingsStub) Station() core.Station {
	return core.Station{}
}

func (s *settingsStub) Contest() core.Contest {
	// the band summaries only know the bands of the contest
	return core.Contest{
		Definition: &conval.Definition{
			Bands: []conval.ContestBand{"80m", "40m", "20m"},
		},
	}
}

type dupeCheckerStub struct{}

func (d *dupeCheckerStub) FindWorkedQSOs(core.Callsign, core.Band, core.Mode) ([]core.QSO, bool) {
	return nil, false
}

type bandmapViewStub struct {
	frames chan core.BandmapFrame
}

func (v *bandmapViewStub) Show() {}
func (v *bandmapViewStub) Hide() {}

func (v *bandmapViewStub) ShowFrame(frame core.BandmapFrame) {
	select {
	case v.frames <- frame:
	default:
	}
}
