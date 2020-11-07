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
		result.prefixes = setupPrefixes()
		log.Print("DXCC prefix database available")
		close(result.available)
	}()

	return result
}

type Finder struct {
	prefixes  *dxcc.Prefixes
	available chan struct{}
}

func (f *Finder) WhenAvailable(callback func()) {
	go func() {
		<-f.available
		callback()
	}()
}

func (f *Finder) Find(s string) (prefix dxcc.Prefix, found bool) {
	if prefixes := f.FindAll(s); len(prefixes) > 0 {
		prefix = prefixes[0]
		found = true
	}
	return
}

func (f *Finder) FindAll(s string) []dxcc.Prefix {
	if f.prefixes == nil {
		return []dxcc.Prefix{}
	}
	result, _ := f.prefixes.Find(s)
	return result
}

func setupPrefixes() *dxcc.Prefixes {
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
