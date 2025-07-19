package entry

import (
	"log"
	"strings"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

type ESMView interface {
	SetESMEnabled(enabled bool)
	SetMessage(message string)
}

type ESMListener interface {
	ESMEnabled(enabled bool)
}

type ESMListenerFunc func(enabled bool)

func (f ESMListenerFunc) ESMEnabled(enabled bool) {
	f(enabled)
}

func (c *Controller) SetESMView(esmView ESMView) {
	if esmView == nil {
		c.esmView = new(nullESMView)
	}
	log.Printf("setting esmView: %t", c.esmEnabled)
	c.esmView = esmView
	c.esmView.SetESMEnabled(c.esmEnabled)
	c.esmView.SetMessage(c.esmMessage)
}

func (c *Controller) ESMEnabled() bool {
	return c.esmEnabled
}

func (c *Controller) SetESMEnabled(enabled bool) {
	c.esmEnabled = enabled
	c.esmView.SetESMEnabled(enabled)
	c.view.SetActiveField(c.activeField)
	c.emitESMEnabled(enabled)
}

func (c *Controller) emitESMEnabled(enabled bool) {
	for _, l := range c.listeners {
		if listener, ok := l.(ESMListener); ok {
			listener.ESMEnabled(enabled)
		}
	}
}

func (c *Controller) NextESMStep() {
	c.updateESM()
	c.keyer.SendText(c.esmMessage)
	switch {
	case c.esmState == core.ESMCallsignValid && c.workmode == core.Run:
		c.GotoNextField()
		if c.activeField == c.theirReportExchangeField.Field {
			c.GotoNextField()
		}
	case c.esmState == core.ESMExchangeValid:
		c.Log()
	}
}

func (c *Controller) updateESM() {
	c.esmState = c.currentESMState()

	switch c.workmode {
	case core.SearchPounce:
		c.esmMessage = c.updateSPMessage()
	case core.Run:
		c.esmMessage = c.updateRunMessage()
	default:
		c.esmMessage = ""
	}
	c.esmView.SetMessage(c.esmMessage)
}

func (c *Controller) currentESMState() core.ESMState {
	switch {
	case c.activeField == core.CallsignField:
		if c.input.callsign == "" {
			return core.ESMCallsignEmpty
		}
		_, err := callsign.Parse(c.input.callsign)
		if err != nil {
			return core.ESMCallsignInvalid
		}
		return core.ESMCallsignValid
	case c.activeField.IsTheirExchange():
		_, err := c.parseTheirExchange(nil, nil, nil)
		if err != nil {
			return core.ESMExchangeInvalid
		}
		return core.ESMExchangeValid
	}
	return core.ESMUnknown
}

func (c *Controller) updateSPMessage() string {
	switch c.esmState {
	case core.ESMCallsignEmpty, core.ESMCallsignInvalid:
		return callsignRequest(c.input.callsign)
	case core.ESMCallsignValid:
		return c.getKeyerText(0)
	case core.ESMExchangeInvalid:
		return "nr?"
	case core.ESMExchangeValid:
		return c.getKeyerText(1)
	default:
		return ""
	}
}

func (c *Controller) updateRunMessage() string {
	switch c.esmState {
	case core.ESMCallsignEmpty:
		return c.getKeyerText(0)
	case core.ESMCallsignInvalid:
		return callsignRequest(c.input.callsign)
	case core.ESMCallsignValid:
		return c.getKeyerText(1)
	case core.ESMExchangeInvalid:
		return "nr?"
	case core.ESMExchangeValid:
		return c.getKeyerText(2)
	default:
		return ""
	}
}

func (c *Controller) getKeyerText(index int) string {
	if c.keyer == nil {
		return ""
	}
	text, err := c.keyer.GetText(c.workmode, index)
	if err != nil {
		return ""
	}
	return text
}

func callsignRequest(input string) string {
	input = strings.ToUpper(input)
	result := make([]byte, 0, len(input)+1)
	for _, b := range input {
		if b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '/' {
			result = append(result, byte(b))
		} else {
			break
		}
	}
	result = append(result, '?')
	return string(result)
}
