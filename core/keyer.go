package core

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ftl/hamradio/callsign"
)

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	MyCall    callsign.Callsign
	TheirCall string
	MyNumber  QSONumber
	MyReport  RST
}

// KeyerValueProvider provides the variable values for the Keyer templates on demand.
type KeyerValueProvider func() KeyerValues

// CWClient defines the interface used by the Keyer to output the CW.
type CWClient interface {
	Send(text string)
}

// Keyer represents the component that sends prepared CW texts using text/templates.
type Keyer interface {
	SetTemplate(index int, pattern string) error
	GetTemplate(index int) string
	GetText(index int) (string, error)
	Send(index int) error
}

// NewKeyer returns a new Keyer that provides len(patterns) templates, based on the given patterns.
func NewKeyer(patterns []string, client CWClient, values KeyerValueProvider) (Keyer, error) {
	templates := make([]*template.Template, len(patterns))
	for i, pattern := range patterns {
		name := fmt.Sprintf("%d", i)
		var err error
		templates[i], err = template.New(name).Parse(pattern)
		if err != nil {
			return nil, err
		}
	}
	return &keyer{patterns, templates, client, values}, nil
}

type keyer struct {
	patterns  []string
	templates []*template.Template
	client    CWClient
	values    KeyerValueProvider
}

func (k *keyer) SetTemplate(index int, pattern string) error {
	var err error
	k.templates[index], err = k.templates[index].Parse(pattern)
	return err
}

func (k *keyer) GetTemplate(index int) string {
	return k.patterns[index]
}

func (k *keyer) GetText(index int) (string, error) {
	buffer := bytes.NewBufferString("")
	err := k.templates[index].Execute(buffer, k.values())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (k *keyer) Send(index int) error {
	message, err := k.GetText(index)
	if err != nil {
		return err
	}
	k.client.Send(message)
	return nil
}
