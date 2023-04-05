package hamlib

import (
	"context"
	"log"
	"time"

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
		controller:      new(nullController),
	}
}

type Client struct {
	conn *client.Conn

	listeners []interface{}

	address         string
	pollingInterval time.Duration
	pollingTimeout  time.Duration
	retryInterval   time.Duration
	requestTimeout  time.Duration
	connected       bool
	closed          chan struct{}
	done            chan struct{}

	bandplan   bandplan.Bandplan
	controller VFOController

	incoming vfoSettings
	outgoing vfoSettings
}

type VFOController interface {
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
}

type vfoSettings struct {
	frequency core.Frequency
	band      core.Band
	mode      core.Mode
}

func (c *Client) SetVFOController(controller VFOController) {
	c.controller = controller
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
	c.emitStatusChanged(c.connected)

	c.conn.StartPolling(c.pollingInterval, c.pollingTimeout,
		client.PollCommand(client.OnFrequency(c.setIncomingFrequency)),
		client.PollCommand(client.OnModeAndPassband(c.setIncomingModeAndPassband)),
	)

	c.conn.WhenClosed(func() {
		c.connected = false
		c.emitStatusChanged(c.connected)

		if whenClosed != nil {
			whenClosed()
		}

		close(c.closed)
	})

	return nil
}

func (c *Client) Disconnect() {
	c.conn.Close()
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
	c.controller.SetFrequency(c.incoming.frequency)
	log.Printf("incoming frequency: %s", c.incoming.frequency)

	band := c.bandplan.ByFrequency(frequency)
	incomingBand := toCoreBand(band.Name)
	if incomingBand == c.incoming.band {
		return
	}
	c.incoming.band = incomingBand
	c.controller.SetBand(c.incoming.band)
	log.Printf("incoming band: %v", c.incoming.band)
}

func (c *Client) setIncomingModeAndPassband(mode client.Mode, _ client.Frequency) {
	incomingMode := toCoreMode(mode)
	if incomingMode == c.incoming.mode {
		return
	}
	c.incoming.mode = incomingMode
	c.controller.SetMode(c.incoming.mode)
	log.Printf("incoming mode %v", incomingMode)
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
	c.outgoing.band = band

	outgoingBandName := toBandplanBandName(c.outgoing.band)
	outgoingBand, ok := c.bandplan[outgoingBandName]
	if !ok {
		log.Printf("unknown band %v", c.outgoing.band)
		return
	}
	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	c.conn.SwitchToBand(ctx, outgoingBand)

	log.Printf("outgoing band: %v", band)
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
		c.controller.SetFrequency(c.incoming.frequency)
	}
	if c.incoming.band != core.NoBand {
		log.Printf("Refreshing VFO band")
		c.controller.SetBand(c.incoming.band)
	}
	if c.incoming.mode != core.NoMode {
		log.Printf("Refreshing VFO mode")
		c.controller.SetMode(c.incoming.mode)
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

func (c *Client) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitStatusChanged(available bool) {
	for _, listener := range c.listeners {
		if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
			serviceStatusListener.StatusChanged(core.HamlibService, available)
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

type nullController struct{}

func (c *nullController) SetFrequency(core.Frequency) {}
func (c *nullController) SetBand(core.Band)           {}
func (c *nullController) SetMode(core.Mode)           {}
