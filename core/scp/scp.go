package scp

import (
	"log"

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

func (f *Finder) FindAnnotated(s string) ([]core.AnnotatedMatch, error) {
	if !f.Available() {
		return nil, nil
	}

	annotatedMatches, err := f.database.FindAnnotated(s)
	if err != nil {
		return nil, err
	}

	return annotateMatches(annotatedMatches), nil
}

func annotateMatches(annotatedMatches []scp.AnnotatedMatch) []core.AnnotatedMatch {
	result := make([]core.AnnotatedMatch, len(annotatedMatches))

	for i, annotatedMatch := range annotatedMatches {
		result[i] = annotateMatch(annotatedMatch)
	}

	return result
}

func annotateMatch(annotatedMatch scp.AnnotatedMatch) core.AnnotatedMatch {
	result := make(core.AnnotatedMatch, len(annotatedMatch))

	for i, part := range annotatedMatch {
		result[i] = core.MatchAnnotation{OP: core.MatchingOperation(part.OP), Value: part.Value}
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
