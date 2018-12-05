package keyer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ftl/hellocontest/core"
)

// New returns a new Keyer that provides len(patterns) templates, based on the given patterns.
func New(patterns []string, client core.CWClient, values core.KeyerValueProvider) (core.Keyer, error) {
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
	client    core.CWClient
	values    core.KeyerValueProvider
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
