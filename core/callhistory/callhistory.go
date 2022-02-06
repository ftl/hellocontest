package callhistory

import (
	"fmt"
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
	callback AvailabilityCallback

	filename string
	field    string
}

type AvailabilityCallback func(service core.Service, available bool)

func (f *Finder) Available() bool {
	return f.database != nil
}

func (f *Finder) ContestChanged(contest core.Contest) {
	if contest.CallHistoryFilename == f.filename {
		return
	}

	f.filename = contest.CallHistoryFilename
	f.field = contest.CallHistoryField

	f.database = loadCallHistory(f.filename)
	f.callback(core.CallHistoryService, f.Available())
	if f.Available() {
		log.Printf("Using call history from %s", f.filename)
	}
}

func (f *Finder) FindEntry(s string) (core.AnnotatedCallsign, bool) {
	if !f.Available() {
		return core.AnnotatedCallsign{}, false
	}
	searchCallsign, err := callsign.Parse(s)
	if err != nil {
		log.Print(err)
		return core.AnnotatedCallsign{}, false
	}
	searchString := searchCallsign.String()

	entries, err := f.database.Find(searchString)
	if err != nil {
		log.Print(err)
		return core.AnnotatedCallsign{}, false
	}

	for _, entry := range entries {
		if entry.Key() == searchString {
			result, err := toAnnotatedCallsign(entry)
			if err != nil {
				log.Print(err)
				return core.AnnotatedCallsign{}, false
			}
			result.PredictedXchange = entry.Get(scp.FieldName(f.field))
			return result, true
		}
	}

	return core.AnnotatedCallsign{}, false
}

func (f *Finder) Find(s string) ([]core.AnnotatedCallsign, error) {
	if !f.Available() {
		return nil, fmt.Errorf("the call history is currently not available")
	}

	matches, err := f.database.Find(s)
	if err != nil {
		return nil, err
	}

	result := make([]core.AnnotatedCallsign, 0, len(matches))
	for _, match := range matches {
		annotatedCallsign, err := toAnnotatedCallsign(match)
		if err != nil {
			log.Print(err)
			continue
		}
		annotatedCallsign.PredictedXchange = match.Get(scp.FieldName(f.field))
		result = append(result, annotatedCallsign)
	}

	return result, nil
}

func toAnnotatedCallsign(match scp.Match) (core.AnnotatedCallsign, error) {
	cs, err := callsign.Parse(match.Key())
	if err != nil {
		return core.AnnotatedCallsign{}, nil
	}
	return core.AnnotatedCallsign{
		Callsign:   cs,
		Assembly:   toMatchingAssembly(match),
		Comparable: match,
		Compare: func(a interface{}, b interface{}) bool {
			aMatch, aOk := a.(scp.Match)
			bMatch, bOk := b.(scp.Match)
			if !aOk || !bOk {
				return false
			}
			return aMatch.LessThan(bMatch)
		},
	}, nil
}

func toMatchingAssembly(match scp.Match) core.MatchingAssembly {
	result := make(core.MatchingAssembly, len(match.Assembly))

	for i, part := range match.Assembly {
		result[i] = core.MatchingPart{OP: core.MatchingOperation(part.OP), Value: part.Value}
	}

	return result
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
