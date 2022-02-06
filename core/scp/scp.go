package scp

import (
	"log"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

func New() *Finder {
	result := &Finder{
		available: make(chan struct{}),
	}

	go func() {
		result.database = setupDatabase()
		if result.database == nil {
			return
		}
		log.Print("Supercheck database available")
		close(result.available)
	}()

	return result
}

type Finder struct {
	database  *scp.Database
	available chan struct{}
}

func (f *Finder) Available() bool {
	select {
	case <-f.available:
		return true
	default:
		return false
	}
}

func (f *Finder) WhenAvailable(callback func()) {
	go func() {
		<-f.available
		callback()
	}()
}

func (f *Finder) FindStrings(s string) ([]string, error) {
	if !f.Available() {
		return nil, nil
	}
	return f.database.FindStrings(s)
}

func (f *Finder) Find(s string) ([]core.AnnotatedCallsign, error) {
	if !f.Available() {
		return nil, nil
	}

	matches, err := f.database.Find(s)
	if err != nil {
		return nil, err
	}

	return toAnnotatedCallsigns(matches), nil
}

func toAnnotatedCallsigns(matches []scp.Match) []core.AnnotatedCallsign {
	result := make([]core.AnnotatedCallsign, 0, len(matches))

	for _, match := range matches {
		annotatedCallsign, err := toAnnotatedCallsign(match)
		if err != nil {
			log.Print(err)
			continue
		}
		result = append(result, annotatedCallsign)
	}

	return result
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

func setupDatabase() *scp.Database {
	localFilename, err := scp.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := scp.Update(scp.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of Supercheck database failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of Supercheck database: %v", localFilename)
	}

	result, err := scp.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load Supercheck database: %v", err)
		return nil
	}
	return result
}
