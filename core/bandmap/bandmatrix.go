package bandmap

import (
	"github.com/ftl/hellocontest/core"
)

type BandMatrixView interface {
	ShowFrame(core.BandMatrixFrame)
}

type BandSwitcher interface {
	SetVFOBand(core.VFOID, core.Band)
}

type VFOFocuser interface {
	SetFocusedVFO(core.VFOID)
}

type BandMatrix struct {
	view     BandMatrixView
	switcher BandSwitcher
	focuser  VFOFocuser

	bands         []core.BandSummary
	vfoBands      [core.VFOCount]core.Band
	focusedVFO    core.VFOID
	vfo2Available bool
}

func NewBandMatrix(switcher BandSwitcher, focuser VFOFocuser) *BandMatrix {
	return &BandMatrix{
		view:     new(nullBandMatrixView),
		switcher: switcher,
		focuser:  focuser,
	}
}

func (m *BandMatrix) SetView(view BandMatrixView) {
	if view == nil {
		m.view = new(nullBandMatrixView)
		return
	}
	m.view = view
	m.emitFrame()
}

func (m *BandMatrix) BandsChanged(bands []core.BandSummary) {
	m.bands = bands
	m.emitFrame()
}

func (m *BandMatrix) VFOBandChanged(vfo core.VFOID, band core.Band) {
	if vfo < 0 || vfo >= core.VFOCount {
		return
	}
	if m.vfoBands[vfo] == band {
		return
	}
	m.vfoBands[vfo] = band
	m.emitFrame()
}

func (m *BandMatrix) FocusedVFOChanged(vfo core.VFOID) {
	if m.focusedVFO == vfo {
		return
	}
	m.focusedVFO = vfo
	m.emitFrame()
}

func (m *BandMatrix) RadioChanged(_ string, singleVFO bool) {
	vfo2Available := !singleVFO
	if m.vfo2Available == vfo2Available {
		return
	}
	m.vfo2Available = vfo2Available
	m.emitFrame()
}

func (m *BandMatrix) SelectBand(vfo core.VFOID, band core.Band) {
	if !m.vfoAvailable(vfo) {
		return
	}
	m.switcher.SetVFOBand(vfo, band)
}

func (m *BandMatrix) ActivateBand(band core.Band) {
	if band == core.NoBand {
		return
	}
	vfo, onVFO := m.vfoOnBand(band)
	switch {
	case !onVFO:
		m.SelectBand(m.focusedVFO, band)
	case vfo != m.focusedVFO:
		m.FocusVFO(vfo)
	}
}

func (m *BandMatrix) FocusVFO(vfo core.VFOID) {
	if !m.vfoAvailable(vfo) {
		return
	}
	m.focuser.SetFocusedVFO(vfo)
}

func (m *BandMatrix) vfoOnBand(band core.Band) (core.VFOID, bool) {
	for vfo, vfoBand := range m.vfoBands {
		if !m.vfoAvailable(core.VFOID(vfo)) {
			continue
		}
		if vfoBand == band {
			return core.VFOID(vfo), true
		}
	}
	return core.VFO1, false
}

func (m *BandMatrix) vfoAvailable(vfo core.VFOID) bool {
	switch vfo {
	case core.VFO1:
		return true
	case core.VFO2:
		return m.vfo2Available
	default:
		return false
	}
}

func (m *BandMatrix) emitFrame() {
	m.view.ShowFrame(core.BandMatrixFrame{
		Bands:         m.bands,
		VFOBands:      m.vfoBands,
		FocusedVFO:    m.focusedVFO,
		VFO2Available: m.vfo2Available,
	})
}

type nullBandMatrixView struct{}

func (v *nullBandMatrixView) ShowFrame(core.BandMatrixFrame) {}
