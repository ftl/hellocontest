package callhistory

import (
	"sync"
	"testing"

	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
)

func TestFinder_AvailableFor(t *testing.T) {
	loaded := &scp.Database{}
	tt := []struct {
		name       string
		database   *scp.Database
		fieldNames []string
		fieldIndex int
		expected   bool
	}{
		{name: "no file loaded", database: nil, fieldNames: []string{"member"}, fieldIndex: 0, expected: false},
		{name: "loaded and mapped", database: loaded, fieldNames: []string{"member"}, fieldIndex: 0, expected: true},
		{name: "loaded but field unmapped", database: loaded, fieldNames: []string{""}, fieldIndex: 0, expected: false},
		{name: "loaded but index out of range", database: loaded, fieldNames: []string{"member"}, fieldIndex: 1, expected: false},
		{name: "loaded but no field names at all", database: loaded, fieldNames: nil, fieldIndex: 0, expected: false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := &Finder{database: tc.database, fieldNames: tc.fieldNames}
			if got := f.AvailableFor(tc.fieldIndex); got != tc.expected {
				t.Errorf("AvailableFor(%d) = %v, want %v", tc.fieldIndex, got, tc.expected)
			}
		})
	}
}

func TestFinder_Deactivate_DropsTheLoadedCallHistory(t *testing.T) {
	available := true
	f := &Finder{
		database:             &scp.Database{},
		dataLock:             new(sync.Mutex),
		filename:             "/home/dl1abc/contest.txt",
		fieldNames:           []string{"name"},
		availabilityCallback: func(_ core.Service, a bool) { available = a },
	}

	f.Deactivate()

	if f.database != nil {
		t.Error("the call history is still loaded")
	}
	if available {
		t.Error("the call history is still reported as available")
	}
}
