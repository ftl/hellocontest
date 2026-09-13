package cluster

import (
	"context"
	"log"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	sdrcore "github.com/ftl/sdrainer/core"
	sdrainer "github.com/ftl/sdrainer/scope/client"

	"github.com/ftl/hellocontest/core"
)

// refreshDivisor gives the period of the refresh: SDRainer reports a running callsign one time,
// therefore its spots need a refresh before they reach the end of their lifetime.
const refreshDivisor = 2

// sdrainerRetryInterval repeats the connection to SDRainer. SDRainer runs on the same computer as
// hellocontest, and the operator restarts it during a contest, therefore the source connects again
// by itself, like the radio.
const sdrainerRetryInterval = 10 * time.Second

// minSpotInterval is the time between two spots of the same callsign on the same channel. Each
// spot costs memory in the bandmap. The DX cluster server of SDRainer uses the same period for its
// repetitions (--spot-every).
const minSpotInterval = 1 * time.Minute

// characterGracePeriod ignores the characters of a channel for a while. The stream reports each
// decoded character, which gives approximately 100 events for each second on a full band, and one
// character for each channel in this period is enough to see that the station still sends.
const characterGracePeriod = 1 * time.Minute

func newSDRainerCluster(parent *Clusters, source core.SpotSource, bandmap Bandmap, bandplan bandplan.Bandplan, clock core.Clock, spotLifetime time.Duration) *cluster {
	result := newCluster(parent, source, bandmap, bandplan, clock, openSDRainer(parent, source, bandmap, clock, spotLifetime/refreshDivisor))
	result.retryInterval = sdrainerRetryInterval
	return result
}

func openSDRainer(parent *Clusters, source core.SpotSource, bandmap Bandmap, clock core.Clock, refreshPeriod time.Duration) openClusterFunc {
	return func(host *net.TCPAddr, username string, password string, trace bool) (clusterClient, error) {
		client := sdrainer.NewClient(host.String())
		err := client.Open()
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(context.Background())
		events, err := client.GetChannelEvents(ctx)
		if err != nil {
			cancel()
			client.Close()
			return nil, err
		}

		result := &sdrainerClient{
			parent:    parent,
			source:    source,
			bandmap:   bandmap,
			clock:     clock,
			client:    client,
			cancel:    cancel,
			connected: true,
			channels:  make(map[string]sdrainerChannel),
		}

		refresh := time.NewTicker(refreshPeriod)
		go func() {
			defer refresh.Stop()
			result.run(events, refresh.C)
		}()

		return result, nil
	}
}

// sdrainerClient receives the channel events of SDRainer and puts the callsigns of the running
// stations into the bandmap. The goroutine of the event stream writes the spots, and the goroutine
// of the UI calls the methods of the clusterClient interface, therefore the lock protects the
// state of the connection.
type sdrainerClient struct {
	parent  *Clusters
	source  core.SpotSource
	bandmap Bandmap
	clock   core.Clock

	client *sdrainer.Client
	cancel context.CancelFunc

	// channels holds the last known state of each channel that SDRainer still has. Only the
	// goroutine of run uses this map.
	channels map[string]sdrainerChannel

	lock      sync.RWMutex
	listeners []any
	connected bool
}

func (c *sdrainerClient) run(events chan any, refresh <-chan time.Time) {
	for {
		select {
		case event, ok := <-events:
			if !ok {
				// the stream ends when SDRainer stops, or when Disconnect cancels its context
				log.Printf("The channel stream of %s ended", c.source.Name)
				c.emitConnected(false)
				return
			}
			c.handleEvent(event)
		case <-refresh:
			c.refreshSpots()
		}
	}
}

// sdrainerChannel is the last known state of one channel, together with the spot that this client
// made of it.
type sdrainerChannel struct {
	channel           sdrainer.Channel
	lastCharacterTime time.Time
	// lastSpot is the spot that this client made of the channel. It takes that spot back when
	// SDRainer destroys the channel or corrects its callsign.
	lastSpot core.Spot
}

// handleEvent keeps one record for each channel, and every event that carries a channel
// overwrites it: SDRainer corrects the callsign of a channel and follows the frequency of its
// signal. Every event also renews the spot, so the bandmap sees how long ago the station was
// really heard. SDRainer reports no ChannelCreated for a channel that it had before this client
// connected, therefore each of these events may bring a new channel.
func (c *sdrainerClient) handleEvent(event any) {
	switch e := event.(type) {
	case sdrainer.ChannelCreated:
		c.channelSeen(e.Channel)
	case sdrainer.ChannelStateChanged:
		c.channelSeen(e.Channel)
	case sdrainer.ChannelQualityChanged:
		c.channelSeen(e.Channel)
	case sdrainer.ChannelCharacterReceived:
		c.characterReceived(e.Channel)
	case sdrainer.ChannelRunningCallsignDetected:
		c.channelSeen(e.Channel)
	case sdrainer.ChannelDestroyed:
		c.channelGone(e.Channel)
	}
}

// characterReceived takes the characters of one channel only after the grace period, because the
// stream reports each single character.
func (c *sdrainerClient) characterReceived(channel sdrainer.Channel) {
	id := string(channel.ID)
	if id == "" {
		return
	}

	record, known := c.channels[id]
	now := c.clock.Now()
	if known && now.Before(record.lastCharacterTime.Add(characterGracePeriod)) {
		return
	}

	record.lastCharacterTime = now
	c.channels[id] = record

	c.channelSeen(channel)
}

func (c *sdrainerClient) channelSeen(channel sdrainer.Channel) {
	id := string(channel.ID)
	if id == "" {
		return
	}

	record := c.channels[id]
	record.channel = channel
	c.channels[id] = record

	// SDRainer corrected the callsign, therefore the spot of the callsign before is obsolete
	if record.lastSpot.Call != channel.Callsign {
		c.removeLastSpot(id)
	}

	c.spot(id)
}

// channelGone takes the spot of a channel back. SDRainer holds a channel as long as it hears the
// station, therefore a destroyed channel means that the station is gone.
func (c *sdrainerClient) channelGone(channel sdrainer.Channel) {
	id := string(channel.ID)
	c.removeLastSpot(id)
	delete(c.channels, id)
}

func (c *sdrainerClient) removeLastSpot(id string) {
	record, ok := c.channels[id]
	if !ok || record.lastSpot.Time.IsZero() {
		return
	}

	c.bandmap.Remove(record.lastSpot)

	record.lastSpot = core.Spot{}
	c.channels[id] = record
}

// refreshSpots spots each channel that SDRainer still has, so that a station stays in the bandmap
// as long as it really runs, also when its channel sends no events.
func (c *sdrainerClient) refreshSpots() {
	for id := range c.channels {
		c.spot(id)
	}
}

func (c *sdrainerClient) spot(id string) {
	record, ok := c.channels[id]
	if !ok {
		return
	}
	now := c.clock.Now()
	if !shouldSpot(record, now) {
		return
	}

	// the bandplan of the parent follows a change of the contest, the same way as for the
	// spots of a DX cluster
	spot, ok := sdrainerSpot(record.channel, c.source.Type, c.parent.bandplan, now)
	if !ok {
		return
	}
	c.bandmap.Add(spot)

	record.lastSpot = spot
	c.channels[id] = record
}

// shouldSpot limits the spots of one channel, but a callsign that SDRainer corrected goes to the
// bandmap at once.
func shouldSpot(record sdrainerChannel, now time.Time) bool {
	if record.lastSpot.Time.IsZero() {
		return true
	}
	if record.channel.Callsign != record.lastSpot.Call {
		return true
	}
	return !now.Before(record.lastSpot.Time.Add(minSpotInterval))
}

func (c *sdrainerClient) Connected() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.connected
}

func (c *sdrainerClient) Notify(listener any) {
	c.lock.Lock()
	c.listeners = append(c.listeners, listener)
	connected := c.connected
	c.lock.Unlock()

	// the stream may end before the cluster registers its listener
	if connected {
		return
	}
	if connectionListener, ok := listener.(connectionListener); ok {
		connectionListener.Connected(false)
	}
}

func (c *sdrainerClient) Disconnect() {
	c.cancel()
	err := c.client.Close()
	if err != nil {
		log.Printf("cannot close the connection to SDRainer: %v", err)
	}
}

func (c *sdrainerClient) emitConnected(connected bool) {
	c.lock.Lock()
	c.connected = connected
	listeners := slices.Clone(c.listeners)
	c.lock.Unlock()

	for _, l := range listeners {
		if listener, ok := l.(connectionListener); ok {
			listener.Connected(connected)
		}
	}
}

type connectionListener interface {
	Connected(bool)
}

// sdrainerSpot makes a spot of a channel that SDRainer detected. SDRainer decodes CW, and it
// reports the frequency in Hz.
func sdrainerSpot(channel sdrainer.Channel, source core.SpotType, plan bandplan.Bandplan, now time.Time) (core.Spot, bool) {
	if channel.Callsign == core.NoCallsign {
		return core.Spot{}, false
	}
	frequency := core.Frequency(channel.Frequency)
	band := toCoreBand(plan.ByFrequency(hamradio.Frequency(frequency)).Name)
	if band == core.NoBand {
		return core.Spot{}, false
	}

	return core.Spot{
		Call:      channel.Callsign,
		Frequency: frequency,
		Band:      band,
		Mode:      core.ModeCW,
		Time:      now,
		Source:    source,
		Quality:   toSpotQuality(channel.Quality),
		SNR:       channel.SNR,
	}, true
}

// toSpotQuality maps the quality of a channel to the quality of a spot. Both use the tags of the
// algorithm of CT1BOH.
func toSpotQuality(quality sdrcore.ChannelQuality) core.SpotQuality {
	switch quality {
	case sdrcore.ValidQuality:
		return core.ValidSpotQuality
	case sdrcore.QSYQuality:
		return core.QSYSpotQuality
	case sdrcore.BustedQuality:
		return core.BustedSpotQuality
	default:
		return core.UnknownSpotQuality
	}
}
