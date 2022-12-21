package core

import (
	"testing"

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
