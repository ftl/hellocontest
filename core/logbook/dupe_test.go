package logbook

import (
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func TestDupeIndex(t *testing.T) {
	dl1abc := callsign.MustParse("DL1ABC")
	index := make(dupeIndex)

	index.Add(dl1abc, core.NoBand, core.NoMode, 1)
	assert.Equal(t, []core.QSONumber{1}, index.Get(dl1abc, core.NoBand, core.NoMode))

	index.Add(dl1abc, core.NoBand, core.NoMode, 1)
	assert.Equal(t, []core.QSONumber{1}, index.Get(dl1abc, core.NoBand, core.NoMode))

	index.Add(dl1abc, core.NoBand, core.NoMode, 3)
	assert.Equal(t, []core.QSONumber{1, 3}, index.Get(dl1abc, core.NoBand, core.NoMode))

	index.Remove(dl1abc, core.NoBand, core.NoMode, 1)
	assert.Equal(t, []core.QSONumber{3}, index.Get(dl1abc, core.NoBand, core.NoMode))

	index.Remove(dl1abc, core.NoBand, core.NoMode, 1)
	assert.Equal(t, []core.QSONumber{3}, index.Get(dl1abc, core.NoBand, core.NoMode))

	index.Remove(dl1abc, core.NoBand, core.NoMode, 3)
	assert.Equal(t, []core.QSONumber{}, index.Get(dl1abc, core.NoBand, core.NoMode))
}
