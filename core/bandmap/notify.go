package bandmap

import (
	"github.com/ftl/hellocontest/core"
)

type EntryAddedListener interface {
	EntryAdded(core.BandmapEntry)
}

type EntryUpdatedListener interface {
	EntryUpdated(core.BandmapEntry)
}

type EntryRemovedListener interface {
	EntryRemoved(core.BandmapEntry)
}

type EntrySelectedListener interface {
	EntrySelected(core.VFOID, core.BandmapEntry)
}

type EntrySelectedListenerFunc func(core.VFOID, core.BandmapEntry)

func (f EntrySelectedListenerFunc) EntrySelected(vfo core.VFOID, entry core.BandmapEntry) {
	f(vfo, entry)
}

type MarkerSelectedListener interface {
	MarkerSelected(core.VFOID, core.BandmapMarker)
}

type MarkerSelectedListenerFunc func(core.VFOID, core.BandmapMarker)

func (f MarkerSelectedListenerFunc) MarkerSelected(vfo core.VFOID, marker core.BandmapMarker) {
	f(vfo, marker)
}

type CQMarkerSelectedListener interface {
	CQMarkerSelected(core.BandmapMarker)
}

type CQMarkerSelectedListenerFunc func(core.BandmapMarker)

func (f CQMarkerSelectedListenerFunc) CQMarkerSelected(marker core.BandmapMarker) {
	f(marker)
}

type MarkersChangedListener interface {
	MarkersChanged([]core.BandmapMarker)
}

type MarkersChangedListenerFunc func([]core.BandmapMarker)

func (f MarkersChangedListenerFunc) MarkersChanged(markers []core.BandmapMarker) {
	f(markers)
}

type EntryOnFrequencyListener interface {
	EntryOnFrequency(core.VFOID, core.BandmapEntry, bool)
}

type MarkerOnFrequencyListener interface {
	MarkerOnFrequency(core.VFOID, core.BandmapMarker, bool)
}

type BandsChangedListener interface {
	BandsChanged([]core.BandSummary)
}

type Notifier struct {
	listeners   []any
	asyncRunner core.AsyncRunner
}

func (n *Notifier) Notify(listener any) {
	n.listeners = append(n.listeners, listener)
}

func (n *Notifier) emitEntryAdded(e core.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryAddedListener, ok := listener.(EntryAddedListener); ok {
			n.asyncRunner(func() {
				entryAddedListener.EntryAdded(e)
			})
		}
	}
}

func (n *Notifier) emitEntryUpdated(e core.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryUpdatedListener, ok := listener.(EntryUpdatedListener); ok {
			n.asyncRunner(func() {
				entryUpdatedListener.EntryUpdated(e)
			})
		}
	}
}

func (n *Notifier) emitEntryRemoved(e core.BandmapEntry) {
	for _, listener := range n.listeners {
		if entryRemovedListener, ok := listener.(EntryRemovedListener); ok {
			n.asyncRunner(func() {
				entryRemovedListener.EntryRemoved(e)
			})
		}
	}
}

func (n *Notifier) emitEntrySelected(vfo core.VFOID, e core.BandmapEntry) {
	for _, listener := range n.listeners {
		if entrySelectedListener, ok := listener.(EntrySelectedListener); ok {
			n.asyncRunner(func() {
				entrySelectedListener.EntrySelected(vfo, e)
			})
		}
	}
}

func (n *Notifier) emitEntryOnFrequency(vfo core.VFOID, e core.BandmapEntry, available bool) {
	for _, listener := range n.listeners {
		if nearestEntryListener, ok := listener.(EntryOnFrequencyListener); ok {
			n.asyncRunner(func() {
				nearestEntryListener.EntryOnFrequency(vfo, e, available)
			})
		}
	}
}

func (n *Notifier) emitBandsChanged(bands []core.BandSummary) {
	for _, listener := range n.listeners {
		if bandsChangedListener, ok := listener.(BandsChangedListener); ok {
			n.asyncRunner(func() {
				bandsChangedListener.BandsChanged(bands)
			})
		}
	}
}

func (n *Notifier) emitMarkerSelected(vfo core.VFOID, marker core.BandmapMarker) {
	for _, listener := range n.listeners {
		if markerSelectedListener, ok := listener.(MarkerSelectedListener); ok {
			n.asyncRunner(func() {
				markerSelectedListener.MarkerSelected(vfo, marker)
			})
		}
	}
}

func (n *Notifier) emitCQMarkerSelected(marker core.BandmapMarker) {
	for _, listener := range n.listeners {
		if cqMarkerSelectedListener, ok := listener.(CQMarkerSelectedListener); ok {
			n.asyncRunner(func() {
				cqMarkerSelectedListener.CQMarkerSelected(marker)
			})
		}
	}
}

func (n *Notifier) emitMarkersChanged(markers []core.BandmapMarker) {
	for _, listener := range n.listeners {
		if markersChangedListener, ok := listener.(MarkersChangedListener); ok {
			n.asyncRunner(func() {
				markersChangedListener.MarkersChanged(markers)
			})
		}
	}
}

func (n *Notifier) emitMarkerOnFrequency(vfo core.VFOID, marker core.BandmapMarker, available bool) {
	for _, listener := range n.listeners {
		if markerOnFrequencyListener, ok := listener.(MarkerOnFrequencyListener); ok {
			n.asyncRunner(func() {
				markerOnFrequencyListener.MarkerOnFrequency(vfo, marker, available)
			})
		}
	}
}
