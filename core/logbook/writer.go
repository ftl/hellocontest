package logbook

import "github.com/ftl/hellocontest/core/model"

type Writer interface {
	WriteQSO(model.QSO) error
	WriteQTC(model.QTC) error
}

var _ Writer = new(nullWriter)

type nullWriter struct{}

func (d *nullWriter) WriteQSO(model.QSO) error {
	return nil
}

func (d *nullWriter) WriteQTC(model.QTC) error {
	return nil
}
