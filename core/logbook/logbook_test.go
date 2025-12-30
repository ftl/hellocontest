package logbook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/clock"
)

func TestLogbook_AddQSO(t *testing.T) {
	logbook := NewLogbook(clock.New())
	qso1 := core.QSO{MyNumber: 1}
	qso2 := core.QSO{MyNumber: 2}
	qso3 := core.QSO{MyNumber: 3}

	require.Empty(t, 0, logbook.AllQSOs())

	logbook.AddQSO(qso1)
	assert.Equal(t, []core.QSO{qso1}, logbook.AllQSOs())

	logbook.AddQSO(qso3)
	assert.Equal(t, []core.QSO{qso1, qso3}, logbook.AllQSOs())

	logbook.AddQSO(qso2)
	assert.Equal(t, []core.QSO{qso1, qso2, qso3}, logbook.AllQSOs())
}
