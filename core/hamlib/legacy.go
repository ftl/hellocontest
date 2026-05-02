package hamlib

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/rigproxy/pkg/client"
	"github.com/ftl/rigproxy/pkg/protocol"

	"github.com/ftl/hellocontest/core"
)

func NewLegacyClient(address string, bandplan bandplan.Bandplan) *LegacyClient {
	return &LegacyClient{
		address:         address,
		pollingInterval: 500 * time.Millisecond,
		pollingTimeout:  2 * time.Second,
		retryInterval:   5 * time.Second,
		requestTimeout:  500 * time.Millisecond,
		done:            make(chan struct{}),
		bandplan:        bandplan,
	}
}

type LegacyClient struct {
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

	incoming vfoState
	outgoing vfoState
}

func (c *LegacyClient) KeepOpen() {
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

func (c *LegacyClient) Connect() error {
	return c.connect(nil)
}

func (c *LegacyClient) connect(whenClosed func()) error {
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
		client.PollCommand(c.onXITActive()),
		client.PollCommand(c.onXITOffset()),
		client.PollCommand(c.onPTTActive()),
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

func (c *LegacyClient) Disconnect() {
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

func (c *LegacyClient) IsConnected() bool {
	return c.connected
}

func (c *LegacyClient) Active() bool {
	return c.connected
}

func (c *LegacyClient) withRequestTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.requestTimeout)
}

func (c *LegacyClient) setIncomingFrequency(frequency client.Frequency) {
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

func (c *LegacyClient) setIncomingModeAndPassband(mode client.Mode, _ client.Frequency) {
	incomingMode := toCoreMode(mode)
	if incomingMode == c.incoming.mode {
		return
	}
	c.incoming.mode = incomingMode
	c.emitModeChanged(c.incoming.mode)
	// log.Printf("incoming mode %v", incomingMode)
}

func (c *LegacyClient) onXITActive() (client.ResponseHandler, string, string) {
	return client.ResponseHandlerFunc(func(r protocol.Response) {
		if len(r.Data) == 0 {
			return
		}
		active := (r.Data[0] == "1")
		c.setIncomingXITActive(active)
	}), "get_func", "XIT"
}

func (c *LegacyClient) setIncomingXITActive(active bool) {
	if c.incoming.xitActive == active {
		return
	}
	c.incoming.xitActive = active
	c.emitXITChanged(c.incoming.xitActive, c.incoming.xitOffset)
	// log.Printf("incoming XIT active: %v %d", c.incoming.xitActive, c.incoming.xitOffset)
}

func (c *LegacyClient) onXITOffset() (client.ResponseHandler, string) {
	return client.ResponseHandlerFunc(func(r protocol.Response) {
		if len(r.Data) == 0 {
			return
		}
		offset, err := strconv.Atoi(r.Data[0])
		if err != nil {
			log.Printf("cannot parse XIT offset: %v", err)
			return
		}
		c.setIncomingXITOffset(offset)
	}), "get_xit"
}

func (c *LegacyClient) setIncomingXITOffset(offset int) {
	incomingXITOffset := core.Frequency(offset)
	if c.incoming.xitOffset == incomingXITOffset {
		return
	}
	c.incoming.xitOffset = incomingXITOffset
	c.emitXITChanged(c.incoming.xitActive, c.incoming.xitOffset)
	// log.Printf("incoming XIT offset: %v %d", c.incoming.xitActive, c.incoming.xitOffset)
}

func (c *LegacyClient) onPTTActive() (client.ResponseHandler, string) {
	return client.ResponseHandlerFunc(func(r protocol.Response) {
		if len(r.Data) == 0 {
			return
		}
		active := (r.Data[0] != "0")
		c.setPTTActive(active)
	}), "get_ptt"
}

func (c *LegacyClient) setPTTActive(active bool) {
	if c.incoming.ptt == active {
		return
	}
	c.incoming.ptt = active
	c.emitPTTChanged(c.incoming.ptt)
	// log.Printf("incoming PTT active: %v", c.incoming.ptt)
}

func (c *LegacyClient) SetFrequency(f core.Frequency) {
	if f == c.outgoing.frequency {
		return
	}
	c.outgoing.frequency = f
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	c.conn.SetFrequency(ctx, client.Frequency(f))

	log.Printf("outgoing frequency: %s", f)
}

func (c *LegacyClient) SetBand(band core.Band) {
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

func (c *LegacyClient) switchToBand(band bandplan.Band) error {
	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	return c.conn.SwitchToBand(ctx, band)
}

func (c *LegacyClient) switchToBandByFrequencyAndMode(band bandplan.Band) error {
	frequency := findModePortionCenter(c.bandplan, int(band.Center()), toBandplanMode(c.incoming.mode))

	ctx, cancel := c.withRequestTimeout()
	defer cancel()
	return c.conn.SetFrequency(ctx, client.Frequency(frequency))
}

func (c *LegacyClient) SetMode(mode core.Mode) {
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

func (c *LegacyClient) SetXIT(active bool, offset core.Frequency) {
	if active == c.outgoing.xitActive && offset == c.outgoing.xitOffset {
		return
	}
	c.outgoing.xitActive = active
	c.outgoing.xitOffset = offset

	if c.conn == nil || c.conn.Closed() {
		return
	}
	ctx, cancel := c.withRequestTimeout()
	defer cancel()

	activeStr := "0"
	if active {
		activeStr = "1"
	}
	err := c.conn.Set(ctx, "set_func", "XIT", activeStr)
	if err != nil {
		log.Printf("setting XTI active failed: %v", err)
		return
	}

	if active {
		err = c.conn.Set(ctx, "set_xit", strconv.Itoa(int(offset)))
	}
	if err != nil {
		log.Printf("setting the XIT offset failed: %v", err)
	}
}

func (c *LegacyClient) Refresh() {
	if c.incoming.frequency != 0 {
		log.Printf("Refreshing VFO frequency: %f", c.incoming.frequency)
		c.emitFrequencyChanged(c.incoming.frequency)
	}
	if c.incoming.band != core.NoBand {
		log.Printf("Refreshing VFO band: %s", c.incoming.band)
		c.emitBandChanged(c.incoming.band)
	}
	if c.incoming.mode != core.NoMode {
		log.Printf("Refreshing VFO mode: %s", c.incoming.mode)
		c.emitModeChanged(c.incoming.mode)
	}
	c.emitXITChanged(c.incoming.xitActive, c.incoming.xitOffset)
	c.emitPTTChanged(c.incoming.ptt)
}

func (c *LegacyClient) Speed(speed int) {
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

func (c *LegacyClient) Send(text string) {
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

func (c *LegacyClient) Abort() {
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

func (c *LegacyClient) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *LegacyClient) emitConnectionChanged(connected bool) {
	type listenerType interface {
		ConnectionChanged(bool)
	}
	for _, listener := range c.listeners {
		if typedListener, ok := listener.(listenerType); ok {
			typedListener.ConnectionChanged(connected)
		}
	}
}

func (c *LegacyClient) emitFrequencyChanged(f core.Frequency) {
	for _, listener := range c.listeners {
		if frequencyListener, ok := listener.(core.VFOFrequencyListener); ok {
			frequencyListener.VFOFrequencyChanged(f)
		}
	}
}

func (c *LegacyClient) emitBandChanged(b core.Band) {
	for _, listener := range c.listeners {
		if bandListener, ok := listener.(core.VFOBandListener); ok {
			log.Printf("triggering band change on %T", bandListener)
			bandListener.VFOBandChanged(b)
		}
	}
}

func (c *LegacyClient) emitModeChanged(m core.Mode) {
	for _, listener := range c.listeners {
		if modeListener, ok := listener.(core.VFOModeListener); ok {
			modeListener.VFOModeChanged(m)
		}
	}
}

func (c *LegacyClient) emitXITChanged(active bool, offset core.Frequency) {
	for _, listener := range c.listeners {
		if xitListener, ok := listener.(core.VFOXITListener); ok {
			xitListener.VFOXITChanged(active, offset)
		}
	}
}

func (c *LegacyClient) emitPTTChanged(active bool) {
	for _, listener := range c.listeners {
		if xitListener, ok := listener.(core.VFOPTTListener); ok {
			xitListener.VFOPTTChanged(active)
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
