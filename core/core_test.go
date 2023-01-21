package core

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntryField_ExchangeField(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.True(t, myExchange.IsMyExchange())

	theirExchange := TheirExchangeField(2)
	assert.True(t, theirExchange.IsTheirExchange())

	assert.False(t, CallsignField.IsMyExchange())
	assert.False(t, CallsignField.IsTheirExchange())
}

func TestEntryField_ExchangeIndex(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.Equal(t, 1, myExchange.ExchangeIndex())

	theirExchange := TheirExchangeField(2)
	assert.Equal(t, 2, theirExchange.ExchangeIndex())

	assert.Equal(t, -1, CallsignField.ExchangeIndex())
}

func TestEntryField_NextExchangeField(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.Equal(t, MyExchangeField(2), myExchange.NextExchangeField())

	theirExchange := TheirExchangeField(2)
	assert.Equal(t, TheirExchangeField(3), theirExchange.NextExchangeField())

	assert.Equal(t, EntryField(""), CallsignField.NextExchangeField())
}

func TestBandGraph_Bindex(t *testing.T) {
	tt := []struct {
		duration time.Duration
		value    time.Duration
		expected int
	}{
		{0, 1 * time.Second, 0},
		{2 * time.Hour, -1 * time.Second, -1},
		{2 * time.Hour, 0, 0},
		{2 * time.Hour, 1 * time.Second, 0},
		{2 * time.Hour, 1*time.Hour - 1*time.Second, 23},
		{2 * time.Hour, 1 * time.Hour, 24},
		{2 * time.Hour, 1*time.Hour + 1*time.Second, 24},
		{2 * time.Hour, 2*time.Hour - 1*time.Second, 47},
		{2 * time.Hour, 2 * time.Hour, -1},
		{2 * time.Hour, 2*time.Hour + 1*time.Second, -1},
	}
	startTime := time.Now()
	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			graph := NewBandGraph(NoBand, startTime, tc.duration)
			actual := graph.Bindex(startTime.Add(tc.value))
			assert.Equal(t, tc.expected, actual)
		})
	}
}
