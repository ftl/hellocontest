package logbook

import (
	"slices"

	"github.com/ftl/hellocontest/core"
)

type dupeKey struct {
	callsign core.Callsign
	band     core.Band
	mode     core.Mode
}

type dupeIndex map[dupeKey][]core.QSONumber

func (i *dupeIndex) Add(callsign core.Callsign, band core.Band, mode core.Mode, number core.QSONumber) {
	key := dupeKey{callsign, band, mode}
	entry := (*i)[key]
	if slices.Contains(entry, number) {
		return
	}

	entry = append(entry, number)
	(*i)[key] = entry
}

func (i *dupeIndex) Remove(callsign core.Callsign, band core.Band, mode core.Mode, number core.QSONumber) {
	key := dupeKey{callsign, band, mode}
	entry := (*i)[key]
	for i, n := range entry {
		if n == number {
			if len(entry) > 1 {
				entry[len(entry)-1], entry[i] = entry[i], entry[len(entry)-1]
				entry = entry[:len(entry)-1]
			} else {
				entry = []core.QSONumber{}
			}
			break
		}
	}
	(*i)[key] = entry
}

func (i *dupeIndex) Get(callsign core.Callsign, band core.Band, mode core.Mode) []core.QSONumber {
	return (*i)[dupeKey{callsign, band, mode}]
}
