package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/pb"
)

type fileFormat interface {
	ReadAll(pbReader) ([]core.QSO, core.Station, core.Contest, error)
	WriteQSO(pbWriter, core.QSO) error
	Clear(pbWriter) error
}

func formatFromFile(filename string) fileFormat {
	file, err := os.Open(filename)
	if err != nil {
		return &unknownFormat{err}
	}
	defer file.Close()
	pbReader := &pbReadWriter{reader: file}

	if fileInfo, err := file.Stat(); err == nil {
		if fileInfo.Size() == 0 {
			return new(latestFormat)
		}
	}

	var preamble int32
	err = binary.Read(file, binary.LittleEndian, &preamble)
	if err != nil {
		return &unknownFormat{err}
	}
	if preamble != 0 {
		return new(v0Format)
	}

	var formatInfo pb.FileInfo
	err = pbReader.Read(&formatInfo)
	if err != nil {
		return &unknownFormat{err}
	}

	switch formatInfo.FormatVersion {
	default:
		return &unknownFormat{fmt.Errorf("%s has an unknown file format", filename)}
	}
}

type unknownFormat struct {
	err error
}

func (f *unknownFormat) ReadAll(pbReader) ([]core.QSO, core.Station, core.Contest, error) {
	return []core.QSO{}, core.Station{}, core.Contest{}, f.err
}

func (f *unknownFormat) WriteQSO(pbWriter, core.QSO) error {
	return f.err
}

func (f *unknownFormat) Clear(pbWriter) error {
	return f.err
}

type v0Format struct {
	filename string
}

func (f *v0Format) ReadAll(r pbReader) ([]core.QSO, core.Station, core.Contest, error) {
	qsos := []core.QSO{}
	var pbQSO pb.QSO
	for {
		err := r.Read(&pbQSO)
		if err == io.EOF {
			return qsos, core.Station{}, core.Contest{}, nil
		} else if err != nil {
			return nil, core.Station{}, core.Contest{}, err
		}
		qso, err := pbToQSO(pbQSO)
		if err != nil {
			return nil, core.Station{}, core.Contest{}, err
		}
		qsos = append(qsos, qso)
	}
}

func (f *v0Format) WriteQSO(w pbWriter, qso core.QSO) error {
	pbQSO := qsoToPB(qso)
	return w.Write(&pbQSO)
}

func (f *v0Format) Clear(pbWriter) error {
	return nil
}
