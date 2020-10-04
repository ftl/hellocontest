package dxcc

import (
	"log"

	"github.com/ftl/hamradio/dxcc"
)

func New() *Finder {
	result := &Finder{}

	go func() {
		result.prefixes = setupPrefixes()
		log.Print("DXCC prefix database available")
	}()

	return result
}

type Finder struct {
	prefixes *dxcc.Prefixes
}

func (f *Finder) Find(s string) ([]dxcc.Prefix, bool) {
	if f.prefixes == nil {
		return []dxcc.Prefix{}, false
	}
	return f.prefixes.Find(s)
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
