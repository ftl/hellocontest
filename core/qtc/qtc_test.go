package qtc

import (
	"testing"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	c := newController().Build()
	assert.NotNil(t, c)
}

func TestOfferQTC_HappyPath(t *testing.T) {
	view := new(fakeView)
	keyer := new(fakeKeyer)
	theirCallsign := callsign.MustParse("DL1ABC")
	c := newController().
		WithKeyer(keyer).
		WithLogbook(&fakeLogbook{nextSeriesNumber: 4, lastCallsign: theirCallsign}).
		WithQTCs(qtcsFor(core.SentQTC, theirCallsign).
			Add("0123", "DK1AB", 1).
			Add("24", "DK2AB", 2).
			Build(),
		).
		Build()
	c.SetView(view)
	c.VFOFrequencyChanged(7020000)
	c.VFOBandChanged(core.Band40m)
	c.VFOModeChanged(core.ModeCW)

	c.OfferQTC()
	assert.Equal(t, "qtc", keyer.lastTransmission)
	assert.Equal(t, core.QTCStart, c.activePhase)
	assert.True(t, view.visible)
	assert.Equal(t, core.ProvideQTC, view.mode)
	assert.Equal(t, theirCallsign, view.series.TheirCallsign)
	assert.Equal(t, "4/2", view.series.Header.String())
	assert.Equal(t, 2, len(view.series.QTCs))
	assert.Equal(t, "DK1AB", view.series.QTCs[0].QTCCallsign.String())
	assert.Equal(t, "DK2AB", view.series.QTCs[1].QTCCallsign.String())
	assert.False(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)

	c.Confirm()
	assert.Equal(t, "qtc 4/2", keyer.lastTransmission)
	assert.Equal(t, core.QTCExchangeHeader, c.activePhase)
	assert.True(t, view.visible)

	c.Confirm()
	assert.Equal(t, "0123 DK1AB 001", keyer.lastTransmission)
	assert.Equal(t, core.QTCExchangeData, c.activePhase)
	assert.True(t, view.visible)
	assert.False(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)

	c.Confirm()
	assert.Equal(t, "24 DK2AB 002", keyer.lastTransmission)
	assert.Equal(t, core.QTCExchangeData, c.activePhase)
	assert.True(t, view.visible)
	assert.True(t, view.series.QTCs[0].Confirmed)
	assert.False(t, view.series.QTCs[1].Confirmed)
	assert.Equal(t, core.Frequency(7020000), c.currentSeries.QTCs[0].Frequency)
	assert.Equal(t, core.Band40m, c.currentSeries.QTCs[0].Band)
	assert.Equal(t, core.ModeCW, c.currentSeries.QTCs[0].Mode)

	c.Confirm()
	assert.Equal(t, "24 DK2AB 002", keyer.lastTransmission)
	assert.Equal(t, core.QTCFinish, c.activePhase)
	assert.True(t, view.visible)
	assert.True(t, view.series.QTCs[0].Confirmed)
	assert.True(t, view.series.QTCs[1].Confirmed)
	assert.Equal(t, core.Frequency(7020000), c.currentSeries.QTCs[1].Frequency)
	assert.Equal(t, core.Band40m, c.currentSeries.QTCs[1].Band)
	assert.Equal(t, core.ModeCW, c.currentSeries.QTCs[1].Mode)

	c.CompleteQTCSeries()
	assert.Equal(t, "tu", keyer.lastTransmission)
	assert.False(t, view.visible)
}
