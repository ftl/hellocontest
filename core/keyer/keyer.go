package keyer

import (
	"bytes"
	"log"
	"text/template"

	"github.com/ftl/hellocontest/core"
)

// NewController returns a new Keyer that has no patterns or templates defined yet.
func NewController(client core.CWClient, values core.KeyerValueProvider) core.KeyerController {
	return &keyer{
		patterns:  make(map[int]string),
		templates: make(map[int]*template.Template),
		client:    client,
		values:    values}
}

type keyer struct {
	patterns  map[int]string
	templates map[int]*template.Template
	client    core.CWClient
	values    core.KeyerValueProvider
	view      core.KeyerView
}

func (k *keyer) SetView(view core.KeyerView) {
	k.view = view
	k.view.SetKeyerController(k)
}

func (k *keyer) EnterPattern(index int, pattern string) {
	k.patterns[index] = pattern
	var err error
	k.templates[index], err = template.New("").Parse(pattern)
	k.view.ShowMessage(err)
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
	err := template.Execute(buffer, k.values())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
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
