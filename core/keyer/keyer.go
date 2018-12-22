package keyer

import (
	"bytes"
	"log"
	"strings"
	"text/template"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

// NewController returns a new Keyer that has no patterns or templates defined yet.
func NewController(client core.CWClient, myCall callsign.Callsign, values core.KeyerValueProvider) core.KeyerController {
	return &keyer{
		myCall:    myCall,
		patterns:  make(map[int]string),
		templates: make(map[int]*template.Template),
		client:    client,
		values:    values}
}

type keyer struct {
	myCall    callsign.Callsign
	patterns  map[int]string
	templates map[int]*template.Template
	client    core.CWClient
	values    core.KeyerValueProvider
	view      core.KeyerView
}

func (k *keyer) SetView(view core.KeyerView) {
	k.view = view
	k.view.SetKeyerController(k)
	for i, pattern := range k.patterns {
		k.view.SetPattern(i, pattern)
	}
}

func (k *keyer) EnterPattern(index int, pattern string) {
	k.patterns[index] = pattern
	var err error
	k.templates[index], err = template.New("").Parse(pattern)
	if err != nil {
		k.view.ShowMessage(err)
	} else {
		k.view.ShowMessage()
	}
}

func (k *keyer) SetPatterns(patterns []string) {
	for i, pattern := range patterns {
		k.patterns[i] = pattern
		k.templates[i], _ = template.New("").Parse(patterns[i])
	}
}

func (k *keyer) GetPattern(index int) string {
	return k.patterns[index]
}

func (k *keyer) GetText(index int) (string, error) {
	buffer := bytes.NewBufferString("")
	template, ok := k.templates[index]
	if !ok {
		return "", nil
	}
	err := template.Execute(buffer, k.fillins())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (k *keyer) fillins() map[string]string {
	values := k.values()
	return map[string]string{
		"MyCall":    k.myCall.String(),
		"MyReport":  softcut(values.MyReport.String()),
		"MyNumber":  softcut(values.MyNumber.String()),
		"MyXchange": values.MyXchange,
		"TheirCall": values.TheirCall,
	}
}

func (k *keyer) Send(index int) {
	message, err := k.GetText(index)
	if err != nil {
		k.view.ShowMessage(err)
		return
	}

	if !k.client.IsConnected() {
		err := k.client.Connect()
		if err != nil {
			k.view.ShowMessage(err)
			return
		}
	}

	log.Printf("sending %s\n", message)
	k.client.Send(message)
}

// Softcut replaces 0 and 9 with their "cut" counterparts t and n.
func softcut(s string) string {
	cuts := map[string]string{
		"0": "t",
		"9": "n",
	}
	result := s
	for digit, cut := range cuts {
		result = strings.Replace(result, digit, cut, -1)
	}
	return result
}

// Cut replaces digits with the "cut" counterparts. (see http://wiki.bavarian-contest-club.de/wiki/Contest-FAQ#Was_sind_.22Cut_Numbers.22.3F)
func cut(s string) string {
	cuts := map[string]string{
		"0": "t",
		"1": "a",
		"2": "u",
		"3": "v",
		"5": "e",
		"7": "g",
		"8": "d",
		"9": "n",
	}
	result := s
	for digit, cut := range cuts {
		result = strings.Replace(result, digit, cut, -1)
	}
	return result
}
