package scp

import (
	"log"

	"github.com/ftl/hamradio/scp"
)

func New() *Finder {
	result := &Finder{
		available: make(chan struct{}),
	}

	go func() {
		result.database = setupDatabase()
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

func (f *Finder) Find(s string) ([]string, error) {
	if f.database == nil {
		return []string{}, nil
	}
	return f.database.Find(s)
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
