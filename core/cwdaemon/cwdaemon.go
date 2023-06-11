package cwdaemon

import (
	"log"

	"github.com/ftl/hamradio/cwclient"

	"github.com/ftl/hellocontest/core/network"
)

type Client struct {
	client    *cwclient.Client
	listeners []any

	connected bool
	wpm       int
}

func NewClient(address string) (*Client, error) {
	host, err := network.ParseTCPAddr(address)
	if err != nil {
		return nil, err
	}

	client, err := cwclient.New(host.IP.String(), host.Port)
	if err != nil {
		return nil, err
	}

	result := &Client{
		client: client,
	}

	err = result.connect()
	if err != nil {
		return nil, err
	}

	return result, nil
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

func (c *Client) connect() error {
	if c.connected && c.client.IsConnected() {
		return nil
	}
	if c.client.IsConnected() {
		c.connected = true
		c.emitConnectionChanged(true)
		return nil
	}

	err := c.client.Connect()
	if err != nil {
		c.connected = false
		c.emitConnectionChanged(false)
		return err
	}

	c.connected = true
	c.emitConnectionChanged(true)
	c.Speed(c.wpm)
	return nil
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *Client) Speed(wpm int) {
	c.wpm = wpm
	if !c.client.IsConnected() {
		if c.connected {
			c.connected = false
			c.emitConnectionChanged(false)
		}
		return
	}
	c.client.Speed(wpm)
}

func (c *Client) Send(text string) {
	err := c.connect()
	if err != nil {
		log.Printf("cannot send %q: %v", text, err)
	}
}

func (c *Client) Abort() {
	if !c.client.IsConnected() {
		if c.connected {
			c.connected = false
			c.emitConnectionChanged(false)
		}
		return
	}
	c.client.Abort()
}
