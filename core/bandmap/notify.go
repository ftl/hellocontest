package bandmap

import (
	"github.com/ftl/hellocontest/core/model"
)

type EntryAddedListener interface {
	EntryAdded(model.BandmapEntry)
}

type EntryUpdatedListener interface {
	EntryUpdated(model.BandmapEntry)
}

type EntryRemovedListener interface {
	EntryRemoved(model.BandmapEntry)
}

type EntrySelectedListener interface {
	EntrySelected(model.VFOID, model.BandmapEntry)
}

type EntrySelectedListenerFunc func(model.VFOID, model.BandmapEntry)

func (f EntrySelectedListenerFunc) EntrySelected(vfo model.VFOID, entry model.BandmapEntry) {
	f(vfo, entry)
}

type MarkerSelectedListener interface {
	MarkerSelected(model.VFOID, model.BandmapMarker)
}

type MarkerSelectedListenerFunc func(model.VFOID, model.BandmapMarker)

func (f MarkerSelectedListenerFunc) MarkerSelected(vfo model.VFOID, marker model.BandmapMarker) {
	f(vfo, marker)
}

type CQMarkerSelectedListener interface {
	CQMarkerSelected(model.BandmapMarker)
}

type CQMarkerSelectedListenerFunc func(model.BandmapMarker)

func (f CQMarkerSelectedListenerFunc) CQMarkerSelected(marker model.BandmapMarker) {
	f(marker)
}

type MarkersChangedListener interface {
	MarkersChanged([]model.BandmapMarker)
}

type MarkersChangedListenerFunc func([]model.BandmapMarker)

func (f MarkersChangedListenerFunc) MarkersChanged(markers []model.BandmapMarker) {
	f(markers)
}

type EntryOnFrequencyListener interface {
	EntryOnFrequency(model.VFOID, model.BandmapEntry, bool)
}

type MarkerOnFrequencyListener interface {
	MarkerOnFrequency(model.VFOID, model.BandmapMarker, bool)
}

type BandsChangedListener interface {
	BandsChanged([]model.BandSummary)
}

type Notifier struct {
	listeners   []any
	asyncRunner model.AsyncRunner
}

func (n *Notifier) Notify(listener any) {
	n.listeners = append(n.listeners, listener)
}

func (n *Notifier) emitEntryAdded(e model.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryAddedListener, ok := listener.(EntryAddedListener); ok {
			n.asyncRunner(func() {
				entryAddedListener.EntryAdded(e)
			})
		}
	}
}

func (n *Notifier) emitEntryUpdated(e model.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryUpdatedListener, ok := listener.(EntryUpdatedListener); ok {
			n.asyncRunner(func() {
				entryUpdatedListener.EntryUpdated(e)
			})
		}
	}
}

func (n *Notifier) emitEntryRemoved(e model.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryRemovedListener, ok := listener.(EntryRemovedListener); ok {
			n.asyncRunner(func() {
				entryRemovedListener.EntryRemoved(e)
			})
		}
	}
}

func (n *Notifier) emitEntrySelected(vfo model.VFOID, e model.BandmapEntry) {
	for _, listener := range n.listeners {
		if entrySelectedListener, ok := listener.(EntrySelectedListener); ok {
			n.asyncRunner(func() {
				entrySelectedListener.EntrySelected(vfo, e)
			})
		}
	}
}

func (n *Notifier) emitEntryOnFrequency(vfo model.VFOID, e model.BandmapEntry, available bool) {
	for _, listener := range n.listeners {
		if nearestEntryListener, ok := listener.(EntryOnFrequencyListener); ok {
			n.asyncRunner(func() {
				nearestEntryListener.EntryOnFrequency(vfo, e, available)
			})
		}
	}
}

func (n *Notifier) emitBandsChanged(bands []model.BandSummary) {
	for _, listener := range n.listeners {
		if bandsChangedListener, ok := listener.(BandsChangedListener); ok {
			n.asyncRunner(func() {
				bandsChangedListener.BandsChanged(bands)
			})
		}
	}
}

func (n *Notifier) emitMarkerSelected(vfo model.VFOID, marker model.BandmapMarker) {
	for _, listener := range n.listeners {
		if markerSelectedListener, ok := listener.(MarkerSelectedListener); ok {
			n.asyncRunner(func() {
				markerSelectedListener.MarkerSelected(vfo, marker)
			})
		}
	}
}

func (n *Notifier) emitCQMarkerSelected(marker model.BandmapMarker) {
	for _, listener := range n.listeners {
		if cqMarkerSelectedListener, ok := listener.(CQMarkerSelectedListener); ok {
			n.asyncRunner(func() {
				cqMarkerSelectedListener.CQMarkerSelected(marker)
			})
		}
	}
}

func (n *Notifier) emitMarkersChanged(markers []model.BandmapMarker) {
	for _, listener := range n.listeners {
		if markersChangedListener, ok := listener.(MarkersChangedListener); ok {
			n.asyncRunner(func() {
				markersChangedListener.MarkersChanged(markers)
			})
		}
	}
}

func (n *Notifier) emitMarkerOnFrequency(vfo model.VFOID, marker model.BandmapMarker, available bool) {
	for _, listener := range n.listeners {
		if markerOnFrequencyListener, ok := listener.(MarkerOnFrequencyListener); ok {
			n.asyncRunner(func() {
				markerOnFrequencyListener.MarkerOnFrequency(vfo, marker, available)
			})
		}
	}
}
