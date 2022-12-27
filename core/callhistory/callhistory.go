// Package callhistory provides access to call history files. Those can be used to predict the exchange data.
// For more information on the supported file format, see https://n1mmwp.hamdocs.com/setup/call-history/
package callhistory

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

func New(settings core.Settings, callback AvailabilityCallback) *Finder {
	result := &Finder{
		callback:   callback,
		filename:   settings.Contest().CallHistoryFilename,
		fieldNames: settings.Contest().CallHistoryFieldNames,
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

	filename   string
	fieldNames []string
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
	f.fieldNames = contest.CallHistoryFieldNames

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

func Export(w io.Writer, qsos ...core.QSO) error {
	callsignToXchange := make(map[string]string)
	// TODO use the new exchange fields
	// for _, qso := range qsos {
	// 	if qso.TheirXchange == "" {
	// 		continue
	// 	}
	// 	callsignToXchange[qso.Callsign.String()] = strings.ToUpper(qso.TheirXchange)
	// }
	entries := make([]string, 0, len(callsignToXchange))
	for callsign, xchange := range callsignToXchange {
		entries = append(entries, fmt.Sprintf("%s,%s", callsign, xchange))
	}
	sort.Strings(entries)

	_, err := fmt.Fprintln(w, "!!Order!!,Call,Exch1\n# Call history created with Hello Contest\n# Enter some additional information here")
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
