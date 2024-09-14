// Package callhistory provides access to call history files. Those can be used to predict the exchange data.
// For more information on the supported file format, see https://n1mmwp.hamdocs.com/setup/call-history/
package callhistory

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

const (
	NameField  = "Name"
	Exch1Field = "Exch1"
)

func New(availabilityCallback AvailabilityCallback, asyncRunner core.AsyncRunner) *Finder {
	return &Finder{
		dataLock:             new(sync.Mutex),
		availabilityCallback: availabilityCallback,
		asyncRunner:          asyncRunner,
	}
}

type Finder struct {
	database *scp.Database
	cache    map[string][]scp.Match
	dataLock *sync.Mutex

	availabilityCallback AvailabilityCallback
	asyncRunner          core.AsyncRunner

	listeners []any

	filename   string
	fieldNames []string
}

type Settings interface {
	core.Settings
}

type AvailabilityCallback func(service core.Service, available bool)

type AvailableFieldNamesListener interface {
	SetAvailableCallHistoryFieldNames(fieldNames []string)
}

func (f *Finder) Notify(listener any) {
	f.listeners = append(f.listeners, listener)
}

func (f *Finder) Activate(filename string) {
	if f.filename == filename {
		return
	}
	f.filename = filename
	f.activateCallHistory()
}

func (f *Finder) Deactivate() {
	f.filename = ""
	clear(f.fieldNames)
}

func (f *Finder) available() bool {
	f.dataLock.Lock()
	defer f.dataLock.Unlock()

	return f.database != nil
}

func (f *Finder) activateCallHistory() {
	f.dataLock.Lock()
	defer f.dataLock.Unlock()

	f.database = loadCallHistory(f.filename)
	f.cache = make(map[string][]scp.Match)

	available := f.database != nil
	f.availabilityCallback(core.CallHistoryService, available)
	if available {
		log.Printf("Using call history from %s with available field names %v", f.filename, f.database.FieldSet().UsableNames())
		f.emitAvailableCallHistoryFieldNames(toFieldNames(f.database.FieldSet().UsableNames()))
	} else {
		log.Printf("No call history available from %s", f.filename)
		f.emitAvailableCallHistoryFieldNames([]string{})
	}
}

func (f *Finder) findInDatabase(s string) ([]scp.Match, error) {
	f.dataLock.Lock()
	defer f.dataLock.Unlock()

	cached, hit := f.cache[s]
	if hit {
		return cached, nil
	}

	matches, err := f.database.Find(s)
	if err != nil {
		return nil, err
	}

	f.cache[s] = matches
	return matches, nil
}

func (f *Finder) emitAvailableCallHistoryFieldNames(fieldNames []string) {
	for _, listener := range f.listeners {
		if fieldNamesListener, ok := listener.(AvailableFieldNamesListener); ok {
			f.asyncRunner(func() {
				fieldNamesListener.SetAvailableCallHistoryFieldNames(fieldNames)
			})
		}
	}
}

func (f *Finder) SelectFieldNames(fieldNames []string) {
	f.fieldNames = fieldNames
}

func (f *Finder) FindEntry(s string) (core.AnnotatedCallsign, bool) {
	if !f.available() {
		return core.AnnotatedCallsign{}, false
	}

	searchCallsign, err := callsign.Parse(s)
	if err != nil {
		log.Print(err)
		return core.AnnotatedCallsign{}, false
	}
	searchString := searchCallsign.String()

	entries, err := f.findInDatabase(searchString)
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
			result.PredictedExchange = make([]string, len(f.fieldNames))
			for i := range f.fieldNames {
				result.PredictedExchange[i] = entry.Get(scp.FieldName(f.fieldNames[i]))
			}
			return result, true
		}
	}

	return core.AnnotatedCallsign{}, false
}

func (f *Finder) Find(s string) ([]core.AnnotatedCallsign, error) {
	if !f.available() {
		return nil, fmt.Errorf("the call history is currently not available")
	}

	matches, err := f.findInDatabase(s)
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
		annotatedCallsign.PredictedExchange = make([]string, len(f.fieldNames))
		for i := range f.fieldNames {
			annotatedCallsign.PredictedExchange[i] = match.Get(scp.FieldName(f.fieldNames[i]))
		}
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
		Name:     match.Get(scp.FieldUserName),
		UserText: match.Get(scp.FieldUserText),
	}, nil
}

func toMatchingAssembly(match scp.Match) core.MatchingAssembly {
	result := make(core.MatchingAssembly, len(match.Assembly))

	for i, part := range match.Assembly {
		result[i] = core.MatchingPart{OP: core.MatchingOperation(part.OP), Value: part.Value}
	}

	return result
}

func toFieldNames(fieldSet scp.FieldSet) []string {
	result := make([]string, len(fieldSet))
	for i, fieldName := range fieldSet {
		result[i] = string(fieldName)
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

func Export(w io.Writer, fieldNames []string, qsos ...core.QSO) error {
	usedFieldNames := make([]string, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		if fieldName != "" {
			usedFieldNames = append(usedFieldNames, fieldName)
		}
	}
	if len(usedFieldNames) == 0 {
		return fmt.Errorf("no field names configured for this contest")
	}

	callsignToExchange := make(map[string][]string)
	for _, qso := range qsos {
		callsignToExchange[qso.Callsign.String()] = qso.TheirExchange
	}

	entries := make([]string, 0, len(callsignToExchange))
	for callsign, exchange := range callsignToExchange {
		usedValues := make([]string, 0, len(fieldNames))
		for i, fieldName := range fieldNames {
			if fieldName == "" {
				continue
			}
			usedValues = append(usedValues, exchange[i])
		}

		entry := fmt.Sprintf("%s,%s", callsign, strings.Join(usedValues, ","))
		entries = append(entries, entry)
	}
	sort.Strings(entries)

	_, err := fmt.Fprintf(w, "!!Order!!,Call,%s\n# Call history created with Hello Contest\n# Enter some additional information here\n", strings.Join(usedFieldNames, ","))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		_, err = fmt.Fprintln(w, entry)
		if err != nil {
			return err
		}
	}
	return nil
}
