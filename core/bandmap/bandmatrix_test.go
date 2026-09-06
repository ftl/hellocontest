package bandmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestBandMatrix_BandsChanged_EmitsFrame(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.BandsChanged([]core.BandSummary{{Band: core.Band80m, SpotCount: 3}})

	require.NotEmpty(t, view.frames)
	require.Len(t, view.lastFrame().Bands, 1)
	assert.Equal(t, core.Band80m, view.lastFrame().Bands[0].Band)
	assert.Equal(t, 3, view.lastFrame().Bands[0].SpotCount)
}

func TestBandMatrix_VFOBandChanged_EmitsBandPerVFO(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.VFOBandChanged(core.VFO1, core.Band40m)
	matrix.VFOBandChanged(core.VFO2, core.Band20m)

	assert.Equal(t, core.Band40m, view.lastFrame().VFOBands[core.VFO1])
	assert.Equal(t, core.Band20m, view.lastFrame().VFOBands[core.VFO2])
}

func TestBandMatrix_FocusedVFOChanged_EmitsFocusedVFO(t *testing.T) {
	matrix, _, _, view := setupBandMatrix()

	matrix.FocusedVFOChanged(core.VFO2)

	assert.Equal(t, core.VFO2, view.lastFrame().FocusedVFO)
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
		vfo           core.VFOID
		wantSwitched  bool
	}{
		{desc: "VFO1", vfo: core.VFO1, wantSwitched: true},
		{desc: "VFO2 available", vfo2Available: true, vfo: core.VFO2, wantSwitched: true},
		{desc: "VFO2 not available", vfo: core.VFO2, wantSwitched: false},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, switcher, _, _ := setupBandMatrix()
			matrix.RadioChanged("radio", !tc.vfo2Available)

			matrix.SelectBand(tc.vfo, core.Band20m)

			if !tc.wantSwitched {
				assert.Empty(t, switcher.calls)
				return
			}
			require.Len(t, switcher.calls, 1)
			assert.Equal(t, tc.vfo, switcher.calls[0].vfo)
			assert.Equal(t, core.Band20m, switcher.calls[0].band)
		})
	}
}

func TestBandMatrix_ActivateBand(t *testing.T) {
	tt := []struct {
		desc          string
		vfo2Available bool
		vfo1Band      core.Band
		vfo2Band      core.Band
		focusedVFO    core.VFOID
		band          core.Band
		wantSwitched  bool
		wantFocused   bool
		wantVFO       core.VFOID
	}{
		{
			desc:         "band on no VFO switches the focused VFO",
			vfo1Band:     core.Band80m,
			focusedVFO:   core.VFO1,
			band:         core.Band20m,
			wantSwitched: true,
			wantVFO:      core.VFO1,
		},
		{
			desc:       "band on the focused VFO does nothing",
			vfo1Band:   core.Band20m,
			focusedVFO: core.VFO1,
			band:       core.Band20m,
		},
		{
			desc:          "band on the other VFO focuses that VFO",
			vfo2Available: true,
			vfo1Band:      core.Band80m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO1,
			band:          core.Band20m,
			wantFocused:   true,
			wantVFO:       core.VFO2,
		},
		{
			desc:          "band on VFO1 focuses VFO1",
			vfo2Available: true,
			vfo1Band:      core.Band80m,
			vfo2Band:      core.Band20m,
			focusedVFO:    core.VFO2,
			band:          core.Band80m,
			wantFocused:   true,
			wantVFO:       core.VFO1,
		},
		{
			desc:         "band on VFO2, but VFO2 not available, switches the focused VFO",
			vfo1Band:     core.Band80m,
			vfo2Band:     core.Band20m,
			focusedVFO:   core.VFO1,
			band:         core.Band20m,
			wantSwitched: true,
			wantVFO:      core.VFO1,
		},
		{
			desc:       "no band does nothing",
			vfo1Band:   core.Band80m,
			focusedVFO: core.VFO1,
			band:       core.NoBand,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			matrix, switcher, focuser, _ := setupBandMatrix()
			matrix.RadioChanged("radio", !tc.vfo2Available)
			matrix.VFOBandChanged(core.VFO1, tc.vfo1Band)
			matrix.VFOBandChanged(core.VFO2, tc.vfo2Band)
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
				require.Equal(t, []core.VFOID{tc.wantVFO}, focuser.calls)
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
		vfo           core.VFOID
		wantFocused   bool
	}{
		{desc: "VFO1", vfo: core.VFO1, wantFocused: true},
		{desc: "VFO2 available", vfo2Available: true, vfo: core.VFO2, wantFocused: true},
		{desc: "VFO2 not available", vfo: core.VFO2, wantFocused: false},
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
			require.Equal(t, []core.VFOID{tc.vfo}, focuser.calls)
		})
	}
}

func setupBandMatrix() (*BandMatrix, *bandSwitcherStub, *vfoFocuserStub, *bandMatrixViewStub) {
	switcher := new(bandSwitcherStub)
	focuser := new(vfoFocuserStub)
	view := new(bandMatrixViewStub)
	matrix := NewBandMatrix(switcher, focuser)
	matrix.SetView(view)
	return matrix, switcher, focuser, view
}

type bandSwitcherCall struct {
	vfo  core.VFOID
	band core.Band
}

type bandSwitcherStub struct {
	calls []bandSwitcherCall
}

func (s *bandSwitcherStub) SetVFOBand(vfo core.VFOID, band core.Band) {
	s.calls = append(s.calls, bandSwitcherCall{vfo: vfo, band: band})
}

type vfoFocuserStub struct {
	calls []core.VFOID
}

func (s *vfoFocuserStub) SetFocusedVFO(vfo core.VFOID) {
	s.calls = append(s.calls, vfo)
}

type bandMatrixViewStub struct {
	frames []core.BandMatrixFrame
}

func (s *bandMatrixViewStub) ShowFrame(frame core.BandMatrixFrame) {
	s.frames = append(s.frames, frame)
}

func (s *bandMatrixViewStub) lastFrame() core.BandMatrixFrame {
	if len(s.frames) == 0 {
		return core.BandMatrixFrame{}
	}
	return s.frames[len(s.frames)-1]
}
