package hamlib

import (
	"context"
	"log"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hellocontest/core"
)

func New(address string, bandplan bandplan.Bandplan) *Client {
	return &Client{
		address:         address,
		pollingInterval: 500 * time.Millisecond,
		pollingTimeout:  2 * time.Second,
		retryInterval:   5 * time.Second,
		requestTimeout:  500 * time.Millisecond,
		done:            make(chan struct{}),
		bandplan:        bandplan,
	}
}

type Client struct {
	conn *client.Conn

	listeners []any

	address         string
	pollingInterval time.Duration
	pollingTimeout  time.Duration
	retryInterval   time.Duration
	requestTimeout  time.Duration
	connected       bool
	closed          chan struct{}
	done            chan struct{}

	bandplan bandplan.Bandplan

	incoming vfoSettings
	outgoing vfoSettings
}

type vfoSettings struct {
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
}

func (c *Client) KeepOpen() {
	go func() {
		disconnected := make(chan bool, 1)
		for {
			err := c.connect(func() {
				disconnected <- true
			})
			if err == nil {
				select {
				case <-disconnected:
					log.Print("Connection lost to Hamlib, waiting for retry.")
				case <-c.done:
					log.Print("Connection to Hamlib closed.")
					return
				}
			} else {
				log.Printf("Cannot connect to Hamlib, waiting for retry: %v", err)
			}

			select {
			case <-time.After(c.retryInterval):
				log.Print("Retrying to connect to Hamlib")
			case <-c.done:
				log.Print("Connection to Hamlib closed.")
				return
			}
		}
	}()
}

func (c *Client) Connect() error {
	return c.connect(nil)
}

func (c *Client) connect(whenClosed func()) error {
	var err error

	c.conn, err = client.Open(c.address)
	if err != nil {
		return err
	}

	c.closed = make(chan struct{})
	c.connected = true
	c.emitConnectionChanged(c.connected)

	c.conn.StartPolling(c.pollingInterval, c.pollingTimeout,
		client.PollCommand(client.OnFrequency(c.setIncomingFrequency)),
		client.PollCommand(client.OnModeAndPassband(c.setIncomingModeAndPassband)),
	)

	c.conn.WhenClosed(func() {
		c.connected = false
		c.emitConnectionChanged(c.connected)

		if whenClosed != nil {
			whenClosed()
		}

		close(c.closed)
	})

	return nil
}

func (c *Client) Disconnect() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Active() bool {
	return c.connected
}

func (c *Client) withRequestTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.requestTimeout)
}

func (c *Client) setIncomingFrequency(frequency client.Frequency) {
	incomingFrequency := core.Frequency(frequency)
	if c.incoming.frequency == incomingFrequency {
		return
	}
	c.incoming.frequency = incomingFrequency
	c.emitFrequencyChanged(c.incoming.frequency)
	// log.Printf("incoming frequency: %s", c.incoming.frequency)

	band := c.bandplan.ByFrequency(frequency)
	incomingBand := toCoreBand(band.Name)
	if incomingBand == c.incoming.band {
		return
	}
	c.incoming.band = incomingBand
	c.emitBandChanged(c.incoming.band)
	// log.Printf("incoming band: %v", c.incoming.band)
}

func (c *Client) setIncomingModeAndPassband(mode client.Mode, _ client.Frequency) {
	incomingMode := toCoreMode(mode)
	if incomingMode == c.incoming.mode {
		return
	}
	c.incoming.mode = incomingMode
	c.emitModeChanged(c.incoming.mode)
	// log.Printf("incoming mode %v", incomingMode)
}

func (c *Client) SetFrequency(f core.Frequency) {
	if f == c.outgoing.frequency {
		return
	}
	c.outgoing.frequency = f
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	c.conn.SetFrequency(ctx, client.Frequency(f))

	log.Printf("outgoing frequency: %s", f)
}

func (c *Client) SetBand(band core.Band) {
	if band == c.outgoing.band {
		return
	}
	if c.conn == nil || c.conn.Closed() {
		return
	}

	outgoingBandName := toBandplanBandName(band)
	outgoingBand, ok := c.bandplan[outgoingBandName]
	if !ok {
		log.Printf("unknown band %v", band)
		return
	}
	c.outgoing.band = band
	log.Printf("outgoing band: %v", band)

	err := c.switchToBand(outgoingBand)
	if err == nil {
		return
	}
	log.Printf("cannot switch to band %s directly: %v", outgoingBand, err)

	err = c.switchToBandByFrequencyAndMode(outgoingBand)
	if err != nil {
		log.Printf("cannot switch to band %s by frequency: %v", band, err)
		return
	}
}

func (c *Client) switchToBand(band bandplan.Band) error {
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	return c.conn.SwitchToBand(ctx, band)
}

func (c *Client) switchToBandByFrequencyAndMode(band bandplan.Band) error {
	frequency := findModePortionCenter(c.bandplan, int(band.Center()), toBandplanMode(c.incoming.mode))

	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	return c.conn.SetFrequency(ctx, client.Frequency(frequency))
}

func (c *Client) SetMode(mode core.Mode) {
	if mode == c.outgoing.mode {
		return
	}
	c.outgoing.mode = mode

	outgoingMode := toClientMode(c.outgoing.mode)
	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	c.conn.SetModeAndPassband(ctx, outgoingMode, 0)

	log.Printf("outgoing mode: %v", mode)
}

func (c *Client) Refresh() {
	if c.incoming.frequency != 0 {
		log.Printf("Refreshing VFO frequency")
		c.emitFrequencyChanged(c.incoming.frequency)
	}
	if c.incoming.band != core.NoBand {
		log.Printf("Refreshing VFO band")
		c.emitBandChanged(c.incoming.band)
	}
	if c.incoming.mode != core.NoMode {
		log.Printf("Refreshing VFO mode")
		c.emitModeChanged(c.incoming.mode)
	}
}

func (c *Client) Speed(speed int) {
	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	err := c.conn.SetMorseSpeed(ctx, speed)
	if err != nil {
		log.Printf("setting the morse speed failed: %v", err)
	}
}

func (c *Client) Send(text string) {
	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	err := c.conn.SendMorse(ctx, text)
	if err != nil {
		log.Printf("sending the morse code failed: %v", err)
	}
}

func (c *Client) Abort() {
	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	err := c.conn.StopMorse(ctx)
	if err != nil {
		log.Printf("stopping the morse code transmission failed: %v", err)
	}
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	type listenerType interface {
		ConnectionChanged(bool)
	}
	for _, listener := range c.listeners {
		if typedListener, ok := listener.(listenerType); ok {
			typedListener.ConnectionChanged(connected)
		}
	}
}

func (c *Client) emitFrequencyChanged(f core.Frequency) {
	for _, listener := range c.listeners {
		if frequencyListener, ok := listener.(core.VFOFrequencyListener); ok {
			frequencyListener.VFOFrequencyChanged(f)
		}
	}
}

func (c *Client) emitBandChanged(b core.Band) {
	for _, listener := range c.listeners {
		if bandListener, ok := listener.(core.VFOBandListener); ok {
			bandListener.VFOBandChanged(b)
		}
	}
}

func (c *Client) emitModeChanged(m core.Mode) {
	for _, listener := range c.listeners {
		if modeListener, ok := listener.(core.VFOModeListener); ok {
			modeListener.VFOModeChanged(m)
		}
	}
}

func toCoreBand(bandName bandplan.BandName) core.Band {
	if bandName == bandplan.BandUnknown {
		return core.NoBand
	}
	return core.Band(bandName)
}

func toBandplanBandName(band core.Band) bandplan.BandName {
	if band == core.NoBand {
		return bandplan.BandUnknown
	}
	return bandplan.BandName(band)
}

func toCoreMode(mode client.Mode) core.Mode {
	switch mode {
	case client.ModeUSB, client.ModeLSB:
		return core.ModeSSB
	case client.ModeCW, client.ModeCWR:
		return core.ModeCW
	case client.ModeRTTY, client.ModeRTTYR:
		return core.ModeRTTY
	case client.ModeFM, client.ModeWFM:
		return core.ModeFM
	case client.ModePKTLSB, client.ModePKTUSB, client.ModePKTFM, client.ModeECSSLSB, client.ModeECSSUSB, client.ModeFAX, client.ModeSAM, client.ModeSAL, client.ModeSAH:
		return core.ModeDigital
	default:
		return core.NoMode
	}
}

func toClientMode(mode core.Mode) client.Mode {
	switch mode {
	case core.ModeCW:
		return client.ModeCW
	case core.ModeSSB:
		return client.ModeUSB // TODO make this dependent of the current frequency either LSB or USB
	case core.ModeFM:
		return client.ModeFM
	case core.ModeRTTY:
		return client.ModeRTTY
	case core.ModeDigital:
		return client.ModePKTUSB
	default:
		return client.ModeNone
	}
}

func toBandplanMode(mode core.Mode) bandplan.Mode {
	log.Printf("to bandplan mode: %s", mode)
	switch mode {
	case core.ModeCW:
		return bandplan.ModeCW
	case core.ModeSSB, core.ModeFM:
		return bandplan.ModePhone
	case core.ModeDigital, core.ModeRTTY:
		return bandplan.ModeDigital
	default:
		return bandplan.ModeDigital
	}
}

func findModePortionCenter(bp bandplan.Bandplan, f int, mode bandplan.Mode) int {
	log.Printf("find mode portion center: %d %s", f, mode)
	frequency := hamradio.Frequency(f)
	band := bp.ByFrequency(frequency)
	var modePortion bandplan.Portion
	var currentPortion bandplan.Portion
	for _, portion := range band.Portions {
		if (portion.Mode == mode && portion.From < frequency) || modePortion.Mode != mode {
			modePortion = portion
		}
		if portion.Contains(frequency) {
			currentPortion = portion
		}
		if modePortion.Mode == mode && currentPortion.Mode != "" {
			break
		}
	}
	if currentPortion.Mode == mode {
		return int(currentPortion.Center())
	}
	if modePortion.Mode == mode {
		return int(modePortion.Center())
	}
	return int(band.Center())
}
