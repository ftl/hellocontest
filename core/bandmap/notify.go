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
	EntrySelected(core.BandmapEntry)
}

type EntryOnFrequencyListener interface {
	EntryOnFrequency(core.BandmapEntry, bool)
}

type Notifier struct {
	listeners []any
}

func (n *Notifier) Notify(listener any) {
	n.listeners = append(n.listeners, listener)
}

func (n *Notifier) emitEntryAdded(e Entry) {
	for _, listener := range n.listeners {
		if entryAddedListener, ok := listener.(EntryAddedListener); ok {
			entryAddedListener.EntryAdded(e.BandmapEntry)
		}
	}
}

func (n *Notifier) emitEntryUpdated(e Entry) {
	for _, listener := range n.listeners {
		if entryUpdatedListener, ok := listener.(EntryUpdatedListener); ok {
			entryUpdatedListener.EntryUpdated(e.BandmapEntry)
		}
	}
}

func (n *Notifier) emitEntryRemoved(e Entry) {
	for _, listener := range n.listeners {
		if entryRemovedListener, ok := listener.(EntryRemovedListener); ok {
			entryRemovedListener.EntryRemoved(e.BandmapEntry)
		}
	}
}

func (n *Notifier) emitEntrySelected(e Entry) {
	for _, listener := range n.listeners {
		if entrySelectedListener, ok := listener.(EntrySelectedListener); ok {
			entrySelectedListener.EntrySelected(e.BandmapEntry)
		}
	}
}

func (n *Notifier) emitEntryOnFrequency(e core.BandmapEntry, available bool) {
	for _, listener := range n.listeners {
		if nearestEntryListener, ok := listener.(EntryOnFrequencyListener); ok {
			nearestEntryListener.EntryOnFrequency(e, available)
		}
	}
}
