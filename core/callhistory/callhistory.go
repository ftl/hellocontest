package callhistory

import (
	"log"
	"os"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

func New(settings core.Settings, callback AvailabilityCallback) *Finder {
	result := &Finder{
		// available: make(chan struct{}),
		callback: callback,
		filename: settings.Contest().CallHistoryFilename,
		field:    settings.Contest().CallHistoryField,
	}

	result.database = loadCallHistory(result.filename)
	result.callback(core.CallHistoryService, result.Available())
	if result.Available() {
		log.Printf("Using call history from %s", result.filename)
	}

	return result
}

type Finder struct {
	database *scp.Database
	// available chan struct{}
	callback AvailabilityCallback

	filename string
	field    string
}

type AvailabilityCallback func(service core.Service, available bool)

func (f *Finder) Available() bool {
	return f.database != nil
	// select {
	// case <-f.available:
	// 	return true
	// default:
	// 	return false
	// }
}

func (f *Finder) SettingsChanged(settings core.Settings) {
	if settings.Contest().CallHistoryFilename == f.filename {
		return
	}

	f.filename = settings.Contest().CallHistoryFilename
	f.field = settings.Contest().CallHistoryField

	f.database = loadCallHistory(f.filename)
	f.callback(core.CallHistoryService, f.Available())
	if f.Available() {
		log.Printf("Using call history from %s", f.filename)
	}
}

func (f *Finder) FindEntry(s string) (core.CallHistoryEntry, bool) {
	if !f.Available() {
		return core.CallHistoryEntry{}, false
	}
	searchCallsign, err := callsign.Parse(s)
	if err != nil {
		log.Print(err)
		return core.CallHistoryEntry{}, false
	}
	searchString := searchCallsign.String()

	entries, err := f.database.FindEntries(searchString)
	if err != nil {
		log.Print(err)
		return core.CallHistoryEntry{}, false
	}

	for _, entry := range entries {
		if entry.Key() == searchString {
			return core.CallHistoryEntry{
				Callsign: searchCallsign,
				// TODO include the fields or at least the f.CallHistoryField
			}, true
		}
	}

	return core.CallHistoryEntry{}, false
}

func loadCallHistory(filename string) *scp.Database {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("cannot open call history file: %v", err)
		return nil
	}
	defer file.Close()
	result, err := scp.ReadCallHistory(file)
	if err != nil {
		log.Printf("cannot load call history: %v", err)
		return nil
	}
	return result
}
