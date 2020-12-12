package dxcc

import (
	"log"

	"github.com/ftl/hamradio/dxcc"
)

func New() *Finder {
	result := &Finder{
		available: make(chan struct{}),
	}

	go func() {
		result.entities = setupEntities()
		log.Print("DXCC prefix database available")
		close(result.available)
	}()

	return result
}

type Finder struct {
	entities  *dxcc.Prefixes
	available chan struct{}
}

func (f *Finder) WhenAvailable(callback func()) {
	go func() {
		<-f.available
		callback()
	}()
}

func (f *Finder) Find(s string) (entity dxcc.Prefix, found bool) {
	if entities := f.FindAll(s); len(entities) > 0 {
		entity = entities[0]
		found = true
	}
	return
}

func (f *Finder) FindAll(s string) []dxcc.Prefix {
	if f.entities == nil {
		return []dxcc.Prefix{}
	}
	result, _ := f.entities.Find(s)
	return result
}

func setupEntities() *dxcc.Prefixes {
	localFilename, err := dxcc.LocalFilename()
	if err != nil {
		log.Print(err)
		return nil
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		log.Printf("update of local copy of DXCC prefixes failed: %v", err)
	}
	if updated {
		log.Printf("updated local copy of DXCC prefixes: %v", localFilename)
	}

	result, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Printf("cannot load DXCC prefixes: %v", err)
		return nil
	}
	return result
}
