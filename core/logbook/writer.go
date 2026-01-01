package logbook

import "github.com/ftl/hellocontest/core"

type Writer interface {
	WriteQSO(core.QSO) error
	WriteQTC(core.QTC) error
}

var _ Writer = new(nullWriter)

type nullWriter struct{}

func (d *nullWriter) WriteQSO(core.QSO) error {
	return nil
}

func (d *nullWriter) WriteQTC(core.QTC) error {
	return nil
}
