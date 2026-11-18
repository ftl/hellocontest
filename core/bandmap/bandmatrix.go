package bandmap

import (
	"log"

	"github.com/ftl/hellocontest/core/model"
)

type BandMatrixView interface {
	Show()
	ShowFrame(model.BandMatrixFrame)
}

type BandSwitcher interface {
	SetVFOBand(model.VFOID, model.Band)
}

type VFOFocuser interface {
	SetFocusedVFO(model.VFOID)
}

type BandMatrix struct {
	view        BandMatrixView
	switcher    BandSwitcher
	focuser     VFOFocuser
	asyncRunner model.AsyncRunner

	bands         []model.BandSummary
	vfoBands      [model.VFOCount]model.Band
	focusedVFO    model.VFOID
	vfo2Available bool
}

func NewBandMatrix(switcher BandSwitcher, focuser VFOFocuser, asyncRunner model.AsyncRunner) *BandMatrix {
	return &BandMatrix{
		view:        new(nullBandMatrixView),
		switcher:    switcher,
		focuser:     focuser,
		asyncRunner: asyncRunner,
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

func (m *BandMatrix) Show() {
	m.onUIThread(func() {
		m.view.Show()
	})
}

func (m *BandMatrix) onUIThread(f func()) {
	// the state and the view of the band matrix belong to the UI thread, but the radio
	// backends emit their events from their own goroutines, see docs/concurrency_guide.md
	if m.asyncRunner == nil {
		f()
		return
	}
	m.asyncRunner(f)
}

func (m *BandMatrix) BandsChanged(bands []model.BandSummary) {
	m.onUIThread(func() {
		m.bands = bands
		m.emitFrame()
	})
}

func (m *BandMatrix) VFOBandChanged(vfo model.VFOID, band model.Band) {
	m.onUIThread(func() {
		if vfo < 0 || vfo >= model.VFOCount {
			return
		}
		if m.vfoBands[vfo] == band {
			return
		}
		m.vfoBands[vfo] = band
		m.emitFrame()
	})
}

func (m *BandMatrix) FocusedVFOChanged(vfo model.VFOID) {
	m.onUIThread(func() {
		if m.focusedVFO == vfo {
			return
		}
		m.focusedVFO = vfo
		m.emitFrame()
	})
}

func (m *BandMatrix) RadioChanged(_ string, singleVFO bool) {
	m.onUIThread(func() {
		vfo2Available := !singleVFO
		if m.vfo2Available == vfo2Available {
			return
		}
		m.vfo2Available = vfo2Available
		m.emitFrame()
	})
}

func (m *BandMatrix) SelectBand(vfo model.VFOID, band model.Band) {
	if !m.vfoAvailable(vfo) {
		log.Printf("band matrix: cannot switch VFO%d to %s, this VFO is not available", vfo+1, band)
		return
	}
	log.Printf("band matrix: switching VFO%d to %s", vfo+1, band)
	m.switcher.SetVFOBand(vfo, band)
}

func (m *BandMatrix) ActivateBand(band model.Band) {
	if band == model.NoBand {
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

func (m *BandMatrix) FocusVFO(vfo model.VFOID) {
	if !m.vfoAvailable(vfo) {
		return
	}
	m.focuser.SetFocusedVFO(vfo)
}

func (m *BandMatrix) vfoOnBand(band model.Band) (model.VFOID, bool) {
	for vfo, vfoBand := range m.vfoBands {
		if !m.vfoAvailable(model.VFOID(vfo)) {
			continue
		}
		if vfoBand == band {
			return model.VFOID(vfo), true
		}
	}
	return model.VFO1, false
}

func (m *BandMatrix) vfoAvailable(vfo model.VFOID) bool {
	switch vfo {
	case model.VFO1:
		return true
	case model.VFO2:
		return m.vfo2Available
	default:
		return false
	}
}

func (m *BandMatrix) emitFrame() {
	m.view.ShowFrame(model.BandMatrixFrame{
		Bands:         m.bands,
		VFOBands:      m.vfoBands,
		FocusedVFO:    m.focusedVFO,
		VFO2Available: m.vfo2Available,
	})
}

type nullBandMatrixView struct{}

func (v *nullBandMatrixView) Show()                           {}
func (v *nullBandMatrixView) ShowFrame(model.BandMatrixFrame) {}
