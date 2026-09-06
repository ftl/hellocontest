package bandmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestBandmap_FilterChange_DeliversAFrame(t *testing.T) {
	bandmap, frames := setupRunningBandmap(t)
	defer bandmap.Close()

	bandmap.SetFilterBand(core.FixedSpotFilterBand(core.Band20m))

	frame := waitForFrame(t, frames)
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
	frame := waitForFrame(t, frames)
	assert.Equal(t, core.SpotFilterAll, frame.Filter.Band.Kind, "the goroutine survived the storm")
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

func setupRunningBandmap(t *testing.T) (*Bandmap, chan core.BandmapFrame) {
	t.Helper()
	frames := make(chan core.BandmapFrame, 1000)
	bandmap := NewBandmap(
		staticClock{now: time.Now()},
		new(settingsStub),
		new(dupeCheckerStub),
		newSpotFilterStoreStub(),
		func(f func()) { f() },
		time.Hour,
		time.Hour,
	)
	bandmap.SetView(&bandmapViewStub{frames: frames})
	return bandmap, frames
}

func waitForFrame(t *testing.T, frames chan core.BandmapFrame) core.BandmapFrame {
	t.Helper()
	select {
	case frame := <-frames:
		// the newest frame that is already available holds the current state
		for {
			select {
			case newer := <-frames:
				frame = newer
			default:
				return frame
			}
		}
	case <-time.After(2 * time.Second):
		require.Fail(t, "the bandmap delivered no frame, its goroutine is blocked")
		return core.BandmapFrame{}
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
	return core.Contest{}
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
