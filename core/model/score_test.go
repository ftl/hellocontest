package model

import (
	"testing"

	"github.com/ftl/conval"
	"github.com/stretchr/testify/assert"
)

func TestBandScore_Copy_DetachesTheMultisPerProperty(t *testing.T) {
	original := BandScore{}
	original.AddQSO(QSOScore{Multis: 5, MultisPerProperty: map[conval.Property]int{"ct_district": 5}})

	copied := original.Copy()
	original.AddQSO(QSOScore{Multis: 5, MultisPerProperty: map[conval.Property]int{"ct_district": 5}})

	assert.Equal(t, map[conval.Property]int{"ct_district": 5}, copied.MultisPerProperty, "the copy keeps its values")
	assert.Equal(t, map[conval.Property]int{"ct_district": 10}, original.MultisPerProperty, "the original counts on")
}

func TestBandScore_AddQSO_DoesNotShareTheMapOfTheQSO(t *testing.T) {
	multisPerProperty := map[conval.Property]int{"ct_district": 5}
	first := BandScore{}
	second := BandScore{}

	first.AddQSO(QSOScore{Multis: 5, MultisPerProperty: multisPerProperty})
	second.AddQSO(QSOScore{Multis: 5, MultisPerProperty: multisPerProperty})

	assert.Equal(t, map[conval.Property]int{"ct_district": 5}, first.MultisPerProperty)
	assert.Equal(t, map[conval.Property]int{"ct_district": 5}, second.MultisPerProperty)
	assert.Equal(t, map[conval.Property]int{"ct_district": 5}, multisPerProperty, "the QSO keeps its own map")
}
