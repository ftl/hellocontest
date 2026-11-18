package logbook

import (
	"slices"

	"github.com/ftl/hellocontest/core/model"
)

type dupeKey struct {
	callsign model.Callsign
	band     model.Band
	mode     model.Mode
}

type dupeIndex map[dupeKey][]model.QSONumber

func (i *dupeIndex) Add(callsign model.Callsign, band model.Band, mode model.Mode, number model.QSONumber) {
	key := dupeKey{callsign, band, mode}
	entry := (*i)[key]
	if slices.Contains(entry, number) {
		return
	}

	entry = append(entry, number)
	(*i)[key] = entry
}

func (i *dupeIndex) Remove(callsign model.Callsign, band model.Band, mode model.Mode, number model.QSONumber) {
	key := dupeKey{callsign, band, mode}
	entry := (*i)[key]
	for i, n := range entry {
		if n == number {
			if len(entry) > 1 {
				entry[len(entry)-1], entry[i] = entry[i], entry[len(entry)-1]
				entry = entry[:len(entry)-1]
			} else {
				entry = []model.QSONumber{}
			}
			break
		}
	}
	(*i)[key] = entry
}

func (i *dupeIndex) Get(callsign model.Callsign, band model.Band, mode model.Mode) []model.QSONumber {
	return (*i)[dupeKey{callsign, band, mode}]
}
