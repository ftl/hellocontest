package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryField_ExchangeField(t *testing.T) {
	myExchange := MyExchangeField(1)
	assert.True(t, myExchange.IsMyExchange())
	assert.Equal(t, 1, myExchange.ExchangeIndex())

	theirExchange := TheirExchangeField(2)
	assert.True(t, theirExchange.IsTheirExchange())
	assert.Equal(t, 2, theirExchange.ExchangeIndex())

	assert.False(t, CallsignField.IsMyExchange())
	assert.False(t, CallsignField.IsTheirExchange())
	assert.Equal(t, -1, CallsignField.ExchangeIndex())
}
