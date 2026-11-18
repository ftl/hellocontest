package logbook

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core/model"
)

func TestDupeIndex(t *testing.T) {
	dl1abc := model.MustParseCallsign("DL1ABC")
	index := make(dupeIndex)

	index.Add(dl1abc, model.NoBand, model.NoMode, 1)
	assert.Equal(t, []model.QSONumber{1}, index.Get(dl1abc, model.NoBand, model.NoMode))

	index.Add(dl1abc, model.NoBand, model.NoMode, 1)
	assert.Equal(t, []model.QSONumber{1}, index.Get(dl1abc, model.NoBand, model.NoMode))

	index.Add(dl1abc, model.NoBand, model.NoMode, 3)
	assert.Equal(t, []model.QSONumber{1, 3}, index.Get(dl1abc, model.NoBand, model.NoMode))

	index.Remove(dl1abc, model.NoBand, model.NoMode, 1)
	assert.Equal(t, []model.QSONumber{3}, index.Get(dl1abc, model.NoBand, model.NoMode))

	index.Remove(dl1abc, model.NoBand, model.NoMode, 1)
	assert.Equal(t, []model.QSONumber{3}, index.Get(dl1abc, model.NoBand, model.NoMode))

	index.Remove(dl1abc, model.NoBand, model.NoMode, 3)
	assert.Equal(t, []model.QSONumber{}, index.Get(dl1abc, model.NoBand, model.NoMode))
}
