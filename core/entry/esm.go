package entry

import (
	"log"
	"strings"

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
	c.esmView.SetMessage(c.esmMessage[c.focusedVFO])
}

func (c *Controller) ESMEnabled() bool {
	return c.esmEnabled
}

func (c *Controller) SetESMEnabled(enabled bool) {
	c.esmEnabled = enabled
	c.esmView.SetESMEnabled(enabled)
	c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])
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
	if !c.canTransmit() {
		return
	}
	c.updateESM()
	c.ignoreVFOChange = true
	c.vfoSwitcher.SetTXVFO(c.focusedVFO)
	c.ignoreVFOChange = false
	if index := c.esmMacroIndex[c.focusedVFO]; index >= 0 {
		c.keyer.Send(index)
	} else {
		c.keyer.SendText(c.esmMessage[c.focusedVFO])
	}
	switch {
	case c.esmState[c.focusedVFO] == core.ESMCallsignValid && c.workmode == core.Run:
		c.GotoNextField()
		if c.activeField[c.focusedVFO] == c.theirReportExchangeField.Field {
			c.GotoNextField()
		}
	case c.esmState[c.focusedVFO] == core.ESMExchangeValid:
		c.Log()
	}
}

func (c *Controller) updateESM() {
	c.esmState[c.focusedVFO] = c.currentESMState()

	switch c.workmode {
	case core.SearchPounce:
		c.esmMessage[c.focusedVFO], c.esmMacroIndex[c.focusedVFO] = c.updateSPMessage()
	case core.Run:
		c.esmMessage[c.focusedVFO], c.esmMacroIndex[c.focusedVFO] = c.updateRunMessage()
	default:
		c.esmMessage[c.focusedVFO] = ""
		c.esmMacroIndex[c.focusedVFO] = -1
	}
	c.esmView.SetMessage(c.esmMessage[c.focusedVFO])
}

func (c *Controller) currentESMState() core.ESMState {
	switch {
	case c.activeField[c.focusedVFO] == core.CallsignField:
		if c.input[c.focusedVFO].callsign == "" {
			return core.ESMCallsignEmpty
		}
		_, err := core.ParseCallsign(c.input[c.focusedVFO].callsign)
		if err != nil {
			return core.ESMCallsignInvalid
		}
		return core.ESMCallsignValid
	case c.activeField[c.focusedVFO].IsTheirExchange():
		_, err := c.parseTheirExchange(nil, nil, nil)
		if err != nil {
			return core.ESMExchangeInvalid
		}
		return core.ESMExchangeValid
	}
	return core.ESMUnknown
}

func (c *Controller) updateSPMessage() (string, int) {
	switch c.esmState[c.focusedVFO] {
	case core.ESMCallsignEmpty, core.ESMCallsignInvalid:
		return callsignRequest(c.input[c.focusedVFO].callsign), -1
	case core.ESMCallsignValid:
		return c.getKeyerText(0), 0
	case core.ESMExchangeInvalid:
		return "nr?", -1
	case core.ESMExchangeValid:
		return c.getKeyerText(1), 1
	default:
		return "", -1
	}
}

func (c *Controller) updateRunMessage() (string, int) {
	switch c.esmState[c.focusedVFO] {
	case core.ESMCallsignEmpty:
		return c.getKeyerText(0), 0
	case core.ESMCallsignInvalid:
		return callsignRequest(c.input[c.focusedVFO].callsign), -1
	case core.ESMCallsignValid:
		return c.getKeyerText(1), 1
	case core.ESMExchangeInvalid:
		return "nr?", -1
	case core.ESMExchangeValid:
		return c.getKeyerText(2), 2
	default:
		return "", -1
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
