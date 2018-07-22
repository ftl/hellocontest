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

// Keyer represents the component that sends prepared CW texts using text/templates.
type Keyer interface {
	SetTemplate(index int, pattern string) error
	GetTemplate(index int) string
	GetText(index int, values KeyerValues) string
}

// NewKeyer returns a new Keyer that provides len(patterns) templates, based on the given patterns.
func NewKeyer(patterns []string) (Keyer, error) {
	templates := make([]*template.Template, len(patterns))
	for i, pattern := range patterns {
		name := fmt.Sprintf("%d", i)
		var err error
		templates[i], err = template.New(name).Parse(pattern)
		if err != nil {
			return nil, err
		}
	}
	return &keyer{patterns, templates}, nil
}

type keyer struct {
	patterns  []string
	templates []*template.Template
}

func (k *keyer) SetTemplate(index int, pattern string) error {
	var err error
	k.templates[index], err = k.templates[index].Parse(pattern)
	return err
}

func (k *keyer) GetTemplate(index int) string {
	return k.patterns[index]
}

func (k *keyer) GetText(index int, values KeyerValues) string {
	buffer := bytes.NewBufferString("")
	k.templates[index].Execute(buffer, values)
	return buffer.String()
}
