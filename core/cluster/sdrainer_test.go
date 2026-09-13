package cluster

import (
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/ftl/hamradio/bandplan"
	sdrcore "github.com/ftl/sdrainer/core"
	sdrainer "github.com/ftl/sdrainer/scope/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

var testSpotTime = time.Date(2026, 9, 12, 16, 51, 0, 0, time.UTC)

func TestSDRainerSpot(t *testing.T) {
	channel := sdrainer.Channel{
		Frequency: 3560000,
		Callsign:  core.MustParseCallsign("dl1abc"),
		SNR:       25,
	}

	spot, ok := sdrainerSpot(channel, core.SkimmerSpot, bandplan.IARURegion1, testSpotTime)

	require.True(t, ok)
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), spot.Call)
	assert.Equal(t, core.Frequency(3560000), spot.Frequency, "SDRainer reports the frequency in Hz")
	assert.Equal(t, core.Band80m, spot.Band)
	assert.Equal(t, core.ModeCW, spot.Mode, "SDRainer decodes CW")
	assert.Equal(t, core.SkimmerSpot, spot.Source, "the type of the source decides")
	assert.Equal(t, 25.0, spot.SNR, "SDRainer measures the signal to noise ratio")
	assert.Equal(t, testSpotTime, spot.Time, "the spot is as new as the event")
}

func TestSDRainerSpot_Quality(t *testing.T) {
	tt := []struct {
		desc    string
		quality sdrcore.ChannelQuality
		want    core.SpotQuality
	}{
		{desc: "valid", quality: sdrcore.ValidQuality, want: core.ValidSpotQuality},
		{desc: "qsy", quality: sdrcore.QSYQuality, want: core.QSYSpotQuality},
		{desc: "busted", quality: sdrcore.BustedQuality, want: core.BustedSpotQuality},
		{desc: "unverified", quality: sdrcore.UnverifiedQuality, want: core.UnknownSpotQuality},
		{desc: "none", quality: sdrcore.NoQuality, want: core.UnknownSpotQuality},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			channel := testChannel("1", 3560000, "dl1abc")
			channel.Quality = tc.quality

			spot, ok := sdrainerSpot(channel, core.SkimmerSpot, bandplan.IARURegion1, testSpotTime)

			require.True(t, ok)
			assert.Equal(t, tc.want, spot.Quality, "SDRainer and the bandmap use the same tags")
		})
	}
}

func TestSDRainerSpot_WithoutACallsign(t *testing.T) {
	channel := sdrainer.Channel{Frequency: 3560000}

	_, ok := sdrainerSpot(channel, core.SkimmerSpot, bandplan.IARURegion1, testSpotTime)

	assert.False(t, ok, "a channel without a callsign is no spot")
}

func TestSDRainerSpot_OutsideOfTheBands(t *testing.T) {
	channel := sdrainer.Channel{
		Frequency: 5000000,
		Callsign:  core.MustParseCallsign("dl1abc"),
	}

	_, ok := sdrainerSpot(channel, core.SkimmerSpot, bandplan.IARURegion1, testSpotTime)

	assert.False(t, ok, "a frequency of no band is no spot")
}

func TestSDRainerClient_SpotsARunningCallsign(t *testing.T) {
	client, bandmap, _ := setupSDRainerClient()
	events := make(chan any, 3)
	events <- sdrainer.ChannelCreated{Channel: testChannel("1", 3560000, "")}
	events <- sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")}
	events <- sdrainer.ChannelDestroyed{Channel: testChannel("1", 3560000, "dl1abc")}
	close(events)

	client.run(events, nil)

	require.Len(t, bandmap.Spots(), 1, "the channel without a callsign is no spot")
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), bandmap.Spots()[0].Call)
}

func TestSDRainerClient_LearnsAChannelFromEveryEvent(t *testing.T) {
	tt := []struct {
		desc  string
		event any
	}{
		{desc: "a character", event: sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560000, "dl1abc")}},
		{desc: "a state", event: sdrainer.ChannelStateChanged{Channel: testChannel("1", 3560000, "dl1abc")}},
		{desc: "a quality", event: sdrainer.ChannelQualityChanged{Channel: testChannel("1", 3560000, "dl1abc")}},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			client, bandmap, _ := setupSDRainerClient()

			// SDRainer sends no ChannelCreated for a channel that it had before this client
			// connected
			client.handleEvent(tc.event)

			require.Len(t, bandmap.Spots(), 1, "every event of a channel brings the spot")
			assert.Equal(t, core.MustParseCallsign("DL1ABC"), bandmap.Spots()[0].Call)
		})
	}
}

func TestSDRainerClient_IgnoresTheCharactersDuringTheGracePeriod(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	character := sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560000, "dl1abc")}

	client.handleEvent(character)
	clock.Add(characterGracePeriod - time.Second)
	client.handleEvent(sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560100, "dl1abc")})

	assert.Equal(t, 3560000.0, client.channels["1"].channel.Frequency,
		"the stream reports each single character")
	require.Len(t, bandmap.Spots(), 1)
}

func TestSDRainerClient_TakesTheCharactersAfterTheGracePeriod(t *testing.T) {
	client, _, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560000, "dl1abc")})

	clock.Add(characterGracePeriod)
	client.handleEvent(sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560100, "dl1abc")})

	assert.Equal(t, 3560100.0, client.channels["1"].channel.Frequency,
		"one character in the grace period shows that the station still sends")
}

func TestSDRainerClient_TakesEveryOtherEventDuringTheGracePeriod(t *testing.T) {
	client, _, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelCharacterReceived{Channel: testChannel("1", 3560000, "dl1abc")})

	clock.Add(time.Second)
	client.handleEvent(sdrainer.ChannelStateChanged{Channel: testChannel("1", 3560100, "dl1abc")})

	assert.Equal(t, 3560100.0, client.channels["1"].channel.Frequency,
		"the grace period holds only for the characters")
}

func TestSDRainerClient_SpotsOneChannelOnlyEveryMinute(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	stateChanged := sdrainer.ChannelStateChanged{Channel: testChannel("1", 3560000, "dl1abc")}

	client.handleEvent(stateChanged)
	clock.Add(minSpotInterval - time.Second)
	client.handleEvent(stateChanged)
	assert.Len(t, bandmap.Spots(), 1, "a spot of the same callsign waits for the interval")

	clock.Add(time.Second)
	client.handleEvent(stateChanged)
	assert.Len(t, bandmap.Spots(), 2, "the spot says how long ago the station was heard")
	assert.Equal(t, clock.Now(), bandmap.Spots()[1].Time)
}

func TestSDRainerClient_SpotsACorrectedCallsignAtOnce(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	clock.Add(time.Second)

	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abd")})

	require.Len(t, bandmap.Spots(), 2, "a corrected callsign waits for no interval")
	assert.Equal(t, core.MustParseCallsign("DL1ABD"), bandmap.Spots()[1].Call)
}

func TestSDRainerClient_RefreshesTheChannelsThatStillExist(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("2", 3570000, "dl2abc")})
	client.handleEvent(sdrainer.ChannelDestroyed{Channel: testChannel("2", 3570000, "dl2abc")})
	bandmap.Clear()
	clock.Add(minSpotInterval)

	client.refreshSpots()

	require.Len(t, bandmap.Spots(), 1, "the destroyed channel is gone")
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), bandmap.Spots()[0].Call)
}

func TestSDRainerClient_RefreshesAlsoAnIdleChannel(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	idle := testChannel("1", 3560000, "dl1abc")
	idle.State = sdrcore.IdleChannel
	client.handleEvent(sdrainer.ChannelStateChanged{Channel: idle})
	bandmap.Clear()
	clock.Add(minSpotInterval)

	client.refreshSpots()

	assert.Len(t, bandmap.Spots(), 1, "a channel lives until SDRainer destroys it")
}

func TestSDRainerClient_RefreshesTheLastStateOfAChannel(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	// SDRainer corrects the callsign and follows the signal
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560100, "dl1abd")})
	bandmap.Clear()
	clock.Add(minSpotInterval)

	client.refreshSpots()

	require.Len(t, bandmap.Spots(), 1, "one record for each channel")
	assert.Equal(t, core.MustParseCallsign("DL1ABD"), bandmap.Spots()[0].Call)
	assert.Equal(t, core.Frequency(3560100), bandmap.Spots()[0].Frequency)
}

func TestSDRainerClient_RefreshesNoChannelWithoutACallsign(t *testing.T) {
	client, bandmap, _ := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelCreated{Channel: testChannel("1", 3560000, "")})

	client.refreshSpots()

	assert.Empty(t, bandmap.Spots(), "a channel without a callsign is no spot")
}

func TestSDRainerClient_RefreshesOnTheTick(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	events := make(chan any, 1)
	refresh := make(chan time.Time, 1)
	events <- sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")}
	done := make(chan struct{})
	go func() {
		defer close(done)
		client.run(events, refresh)
	}()

	require.Eventually(t, func() bool { return len(bandmap.Spots()) == 1 }, time.Second, 10*time.Millisecond)
	clock.Add(minSpotInterval)
	refresh <- testSpotTime
	require.Eventually(t, func() bool { return len(bandmap.Spots()) == 2 }, time.Second, 10*time.Millisecond)

	close(events)
	<-done
}

func TestSDRainerClient_TakesTheSpotOfADestroyedChannelBack(t *testing.T) {
	client, bandmap, _ := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})

	client.handleEvent(sdrainer.ChannelDestroyed{Channel: testChannel("1", 3560000, "dl1abc")})

	require.Len(t, bandmap.RemovedSpots(), 1, "SDRainer hears the station no more")
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), bandmap.RemovedSpots()[0].Call)
	assert.Equal(t, core.Frequency(3560000), bandmap.RemovedSpots()[0].Frequency)
	assert.Empty(t, client.channels, "the channel is gone")
}

func TestSDRainerClient_TakesTheSpotOfACorrectedCallsignBack(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	clock.Add(time.Second)

	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abd")})

	require.Len(t, bandmap.RemovedSpots(), 1, "the callsign before is obsolete")
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), bandmap.RemovedSpots()[0].Call)
	require.Len(t, bandmap.Spots(), 2)
	assert.Equal(t, core.MustParseCallsign("DL1ABD"), bandmap.Spots()[1].Call)
}

func TestSDRainerClient_TakesNoSpotBackWithoutASpot(t *testing.T) {
	client, bandmap, _ := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelCreated{Channel: testChannel("1", 3560000, "")})

	client.handleEvent(sdrainer.ChannelDestroyed{Channel: testChannel("1", 3560000, "")})

	assert.Empty(t, bandmap.RemovedSpots(), "this channel made no spot")
}

func TestSDRainerClient_TakesASpotBackOnlyOneTime(t *testing.T) {
	client, bandmap, clock := setupSDRainerClient()
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abc")})
	clock.Add(time.Second)
	client.handleEvent(sdrainer.ChannelRunningCallsignDetected{Channel: testChannel("1", 3560000, "dl1abd")})

	client.handleEvent(sdrainer.ChannelDestroyed{Channel: testChannel("1", 3560000, "dl1abd")})

	removed := bandmap.RemovedSpots()
	require.Len(t, removed, 2)
	assert.Equal(t, core.MustParseCallsign("DL1ABC"), removed[0].Call)
	assert.Equal(t, core.MustParseCallsign("DL1ABD"), removed[1].Call, "each spot goes back one time")
}

func TestSDRainerClient_ReportsTheEndOfTheStream(t *testing.T) {
	client, _, _ := setupSDRainerClient()
	listener := new(connectionSpy)
	client.Notify(listener)
	events := make(chan any)
	close(events)

	client.run(events, nil)

	assert.False(t, client.Connected())
	assert.Equal(t, []bool{false}, listener.states, "SDRainer stopped, the source is disconnected")
}

func setupSDRainerClient() (*sdrainerClient, *bandmapSpy, *testClock) {
	bandmap := new(bandmapSpy)
	clock := &testClock{now: testSpotTime}
	client := &sdrainerClient{
		parent:    &Clusters{bandplan: bandplan.IARURegion1},
		source:    core.SpotSource{Type: core.SkimmerSpot},
		bandmap:   bandmap,
		clock:     clock,
		connected: true,
		channels:  make(map[string]sdrainerChannel),
	}
	return client, bandmap, clock
}

func testChannel(id string, frequency float64, call string) sdrainer.Channel {
	result := sdrainer.Channel{
		ID:        sdrcore.ChannelID(id),
		Frequency: frequency,
	}
	if call != "" {
		result.Callsign = core.MustParseCallsign(call)
	}
	return result
}

// bandmapSpy is used from the test goroutine and from the goroutine of the client, therefore it
// needs the lock
type bandmapSpy struct {
	lock         sync.RWMutex
	spots        []core.Spot
	removedSpots []core.Spot
}

func (b *bandmapSpy) Add(spot core.Spot) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.spots = append(b.spots, spot)
}

func (b *bandmapSpy) Remove(spot core.Spot) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.removedSpots = append(b.removedSpots, spot)
}

func (b *bandmapSpy) RemovedSpots() []core.Spot {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return slices.Clone(b.removedSpots)
}

func (b *bandmapSpy) Spots() []core.Spot {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return slices.Clone(b.spots)
}

func (b *bandmapSpy) Clear() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.spots = nil
}

type connectionSpy struct {
	states []bool
}

func (c *connectionSpy) Connected(connected bool) {
	c.states = append(c.states, connected)
}

// testClock is used from the test goroutine and from the goroutine of the client, therefore it
// needs the lock
type testClock struct {
	lock sync.RWMutex
	now  time.Time
}

func (c *testClock) Now() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.now
}

func (c *testClock) Add(d time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.now = c.now.Add(d)
}
