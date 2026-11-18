package bandmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core/model"
)

func TestBandMatrix_Show(t *testing.T) {
	matrix := NewBandMatrix(new(bandSwitcherStub), new(vfoFocuserStub), nil)
	view := new(bandMatrixViewStub)
	matrix.SetView(view)

	matrix.Show()

	assert.True(t, view.shown)
}

func TestBandMatrix_BandsChanged_EmitsFrame(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.BandsChanged([]model.BandSummary{{Band: model.Band80m, SpotCount: 3}})

	require.NotEmpty(t, view.frames)
	require.Len(t, view.lastFrame().Bands, 1)
	assert.Equal(t, model.Band80m, view.lastFrame().Bands[0].Band)
	assert.Equal(t, 3, view.lastFrame().Bands[0].SpotCount)
}

func TestBandMatrix_VFOBandChanged_EmitsBandPerVFO(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.VFOBandChanged(model.VFO1, model.Band40m)
	matrix.VFOBandChanged(model.VFO2, model.Band20m)

	assert.Equal(t, model.Band40m, view.lastFrame().VFOBands[model.VFO1])
	assert.Equal(t, model.Band20m, view.lastFrame().VFOBands[model.VFO2])
}

func TestBandMatrix_FocusedVFOChanged_EmitsFocusedVFO(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.FocusedVFOChanged(model.VFO2)

	assert.Equal(t, model.VFO2, view.lastFrame().FocusedVFO)
}

func TestBandMatrix_RadioChanged_TogglesVFO2Availability(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.RadioChanged("dual", false)
	assert.True(t, view.lastFrame().VFO2Available, "dual VFO radio")

	matrix.RadioChanged("single", true)
	assert.False(t, view.lastFrame().VFO2Available, "single VFO radio")
}

func TestBandMatrix_SelectBand(t *testing.T) {
	tt := []struct {
		desc          string
		vfo2Available bool
		vfo           model.VFOID
		wantSwitched  bool
	}{
		{desc: "VFO1", vfo: model.VFO1, wantSwitched: true},
		{desc: "VFO2 available", vfo2Available: true, vfo: model.VFO2, wantSwitched: true},
		{desc: "VFO2 not available", vfo: model.VFO2, wantSwitched: false},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, switcher, _, _ := setupBandMatrix()
			matrix.RadioChanged("radio", !tc.vfo2Available)

			matrix.SelectBand(tc.vfo, model.Band20m)

			if !tc.wantSwitched {
				assert.Empty(t, switcher.calls)
				return
			}
			require.Len(t, switcher.calls, 1)
			assert.Equal(t, tc.vfo, switcher.calls[0].vfo)
			assert.Equal(t, model.Band20m, switcher.calls[0].band)
		})
	}
}

func TestBandMatrix_SelectBand_SwitchesTheVFOOfTheColumn(t *testing.T) {
	tt := []struct {
		desc       string
		focusedVFO model.VFOID
		vfo1Band   model.Band
		vfo2Band   model.Band
		vfo        model.VFOID
		band       model.Band
	}{
		{
			desc:       "the other VFO goes to a free band",
			focusedVFO: model.VFO1,
			vfo1Band:   model.Band80m,
			vfo2Band:   model.Band40m,
			vfo:        model.VFO2,
			band:       model.Band20m,
		},
		{
			desc:       "the other VFO goes to the band of the focused VFO",
			focusedVFO: model.VFO1,
			vfo1Band:   model.Band80m,
			vfo2Band:   model.Band40m,
			vfo:        model.VFO2,
			band:       model.Band80m,
		},
		{
			desc:       "the other VFO stays on its own band",
			focusedVFO: model.VFO1,
			vfo1Band:   model.Band80m,
			vfo2Band:   model.Band40m,
			vfo:        model.VFO2,
			band:       model.Band40m,
		},
		{
			desc:       "the focused VFO goes to the band of the other VFO",
			focusedVFO: model.VFO2,
			vfo1Band:   model.Band80m,
			vfo2Band:   model.Band40m,
			vfo:        model.VFO2,
			band:       model.Band80m,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, switcher, focuser, _ := setupBandMatrix()
			matrix.RadioChanged("radio", false)
			matrix.VFOBandChanged(model.VFO1, tc.vfo1Band)
			matrix.VFOBandChanged(model.VFO2, tc.vfo2Band)
			matrix.FocusedVFOChanged(tc.focusedVFO)

			matrix.SelectBand(tc.vfo, tc.band)

			require.Len(t, switcher.calls, 1, "the column of a VFO always switches that VFO")
			assert.Equal(t, tc.vfo, switcher.calls[0].vfo)
			assert.Equal(t, tc.band, switcher.calls[0].band)
			assert.Empty(t, focuser.calls, "the column of a VFO does not move the focus")
		})
	}
}

func TestBandMatrix_ActivateBand(t *testing.T) {
	tt := []struct {
		desc          string
		vfo2Available bool
		vfo1Band      model.Band
		vfo2Band      model.Band
		focusedVFO    model.VFOID
		band          model.Band
		wantSwitched  bool
		wantFocused   bool
		wantVFO       model.VFOID
	}{
		{
			desc:         "band on no VFO switches the focused VFO",
			vfo1Band:     model.Band80m,
			focusedVFO:   model.VFO1,
			band:         model.Band20m,
			wantSwitched: true,
			wantVFO:      model.VFO1,
		},
		{
			desc:       "band on the focused VFO does nothing",
			vfo1Band:   model.Band20m,
			focusedVFO: model.VFO1,
			band:       model.Band20m,
		},
		{
			desc:          "band on the other VFO focuses that VFO",
			vfo2Available: true,
			vfo1Band:      model.Band80m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO1,
			band:          model.Band20m,
			wantFocused:   true,
			wantVFO:       model.VFO2,
		},
		{
			desc:          "band on VFO1 focuses VFO1",
			vfo2Available: true,
			vfo1Band:      model.Band80m,
			vfo2Band:      model.Band20m,
			focusedVFO:    model.VFO2,
			band:          model.Band80m,
			wantFocused:   true,
			wantVFO:       model.VFO1,
		},
		{
			desc:         "band on VFO2, but VFO2 not available, switches the focused VFO",
			vfo1Band:     model.Band80m,
			vfo2Band:     model.Band20m,
			focusedVFO:   model.VFO1,
			band:         model.Band20m,
			wantSwitched: true,
			wantVFO:      model.VFO1,
		},
		{
			desc:       "no band does nothing",
			vfo1Band:   model.Band80m,
			focusedVFO: model.VFO1,
			band:       model.NoBand,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, switcher, focuser, _ := setupBandMatrix()
			matrix.RadioChanged("radio", !tc.vfo2Available)
			matrix.VFOBandChanged(model.VFO1, tc.vfo1Band)
			matrix.VFOBandChanged(model.VFO2, tc.vfo2Band)
			matrix.FocusedVFOChanged(tc.focusedVFO)

			matrix.ActivateBand(tc.band)

			if tc.wantSwitched {
				require.Len(t, switcher.calls, 1)
				assert.Equal(t, tc.wantVFO, switcher.calls[0].vfo)
				assert.Equal(t, tc.band, switcher.calls[0].band)
			} else {
				assert.Empty(t, switcher.calls, "no band switch")
			}

			if tc.wantFocused {
				require.Equal(t, []model.VFOID{tc.wantVFO}, focuser.calls)
			} else {
				assert.Empty(t, focuser.calls, "no focus change")
			}
		})
	}
}

func TestBandMatrix_FocusVFO(t *testing.T) {
	tt := []struct {
		desc          string
		vfo2Available bool
		vfo           model.VFOID
		wantFocused   bool
	}{
		{desc: "VFO1", vfo: model.VFO1, wantFocused: true},
		{desc: "VFO2 available", vfo2Available: true, vfo: model.VFO2, wantFocused: true},
		{desc: "VFO2 not available", vfo: model.VFO2, wantFocused: false},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, _, focuser, _ := setupBandMatrix()
			matrix.RadioChanged("radio", !tc.vfo2Available)

			matrix.FocusVFO(tc.vfo)

			if !tc.wantFocused {
				assert.Empty(t, focuser.calls)
				return
			}
			require.Equal(t, []model.VFOID{tc.vfo}, focuser.calls)
		})
	}
}

func setupBandMatrix() (*BandMatrix, *bandSwitcherStub, *vfoFocuserStub, *bandMatrixViewStub) {
	switcher := new(bandSwitcherStub)
	focuser := new(vfoFocuserStub)
	view := new(bandMatrixViewStub)
	matrix := NewBandMatrix(switcher, focuser, func(f func()) { f() })
	matrix.SetView(view)
	return matrix, switcher, focuser, view
}

type bandSwitcherCall struct {
	vfo  model.VFOID
	band model.Band
}

type bandSwitcherStub struct {
	calls []bandSwitcherCall
}

func (s *bandSwitcherStub) SetVFOBand(vfo model.VFOID, band model.Band) {
	s.calls = append(s.calls, bandSwitcherCall{vfo: vfo, band: band})
}

type vfoFocuserStub struct {
	calls []model.VFOID
}

func (s *vfoFocuserStub) SetFocusedVFO(vfo model.VFOID) {
	s.calls = append(s.calls, vfo)
}

type bandMatrixViewStub struct {
	frames []model.BandMatrixFrame
	shown  bool
}

func (s *bandMatrixViewStub) Show() {
	s.shown = true
}

func (s *bandMatrixViewStub) ShowFrame(frame model.BandMatrixFrame) {
	s.frames = append(s.frames, frame)
}

func (s *bandMatrixViewStub) lastFrame() model.BandMatrixFrame {
	if len(s.frames) == 0 {
		return model.BandMatrixFrame{}
	}
	return s.frames[len(s.frames)-1]
}

func TestBandMatrix_EventsFromManyGoroutines_StayOnTheUIThread(t *testing.T) {
	// the radio backends emit from their own goroutines, the band matrix must not touch
	// its state or its view there
	uiThread := make(chan func(), 1000)
	done := make(chan struct{})
	go func() {
		for f := range uiThread {
			f()
		}
		close(done)
	}()

	matrix := NewBandMatrix(new(bandSwitcherStub), new(vfoFocuserStub), func(f func()) { uiThread <- f })
	view := new(bandMatrixViewStub)
	matrix.SetView(view)

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			matrix.VFOBandChanged(model.VFOID(i%int(model.VFOCount)), model.Bands[i%len(model.Bands)])
			matrix.FocusedVFOChanged(model.VFOID(i % int(model.VFOCount)))
			matrix.RadioChanged("radio", i%2 == 0)
			matrix.BandsChanged([]model.BandSummary{{Band: model.Band20m, SpotCount: i}})
		}(i)
	}
	wg.Wait()
	close(uiThread)
	<-done

	assert.NotEmpty(t, view.frames, "the frames arrive on the UI thread")
}
