package store

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/pb"
)

type fileFormat interface {
	ReadAll(pbReader) ([]core.QSO, *core.Station, *core.Contest, error)
	WriteQSO(pbWriter, core.QSO) error
	WriteStation(pbWriter, core.Station) error
	WriteContest(pbWriter, core.Contest) error
	Clear(pbWriter) error
}

func formatFromFile(filename string) fileFormat {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return new(latestFormat)
	} else if err != nil {
		return &unknownFormat{err}
	}
	defer file.Close()
	pbReader := &pbReadWriter{reader: file}

	if fileInfo, err := file.Stat(); err == nil {
		if fileInfo.Size() == 0 {
			return new(v0Format)
		}
	}

	preamble, err := pbReader.ReadPreamble()
	if preamble != 0 {
		return new(v0Format)
	}

	var formatInfo pb.FileInfo
	err = pbReader.Read(&formatInfo)
	if err != nil {
		return &unknownFormat{err}
	}

	switch formatInfo.FormatVersion {
	case 1:
		return new(v1Format)
	default:
		return &unknownFormat{fmt.Errorf("%s has an unknown file format", filename)}
	}
}

type unknownFormat struct {
	err error
}

func (f *unknownFormat) ReadAll(pbReader) ([]core.QSO, *core.Station, *core.Contest, error) {
	return nil, nil, nil, f.err
}

func (f *unknownFormat) WriteQSO(pbWriter, core.QSO) error {
	return f.err
}

func (f *unknownFormat) WriteStation(pbWriter, core.Station) error {
	return f.err
}

func (f *unknownFormat) WriteContest(pbWriter, core.Contest) error {
	return f.err
}

func (f *unknownFormat) Clear(pbWriter) error {
	return f.err
}

type v0Format struct {
	filename string
}

func (f *v0Format) ReadAll(r pbReader) ([]core.QSO, *core.Station, *core.Contest, error) {
	qsos := []core.QSO{}
	var pbQSO pb.QSO
	for {
		err := r.Read(&pbQSO)
		if err == io.EOF {
			return qsos, nil, nil, nil
		} else if err != nil {
			return nil, nil, nil, err
		}
		qso, err := pbToQSO(pbQSO)
		if err != nil {
			return nil, nil, nil, err
		}
		qsos = append(qsos, qso)
	}
}

func (f *v0Format) WriteQSO(w pbWriter, qso core.QSO) error {
	pbQSO := qsoToPB(qso)
	return w.Write(&pbQSO)
}

func (f *v0Format) WriteStation(pbWriter, core.Station) error {
	log.Println("The V0 file format cannot store station data.")
	return nil
}

func (f *v0Format) WriteContest(pbWriter, core.Contest) error {
	log.Println("The V0 file format cannot store contest data.")
	return nil
}

func (f *v0Format) Clear(pbWriter) error {
	return nil
}

type v1Format struct {
	filename string
}

func (f *v1Format) ReadAll(r pbReader) ([]core.QSO, *core.Station, *core.Contest, error) {
	var (
		pbFormatInfo pb.FileInfo
		pbEntry      pb.Entry
	)
	_, err := r.ReadPreamble()
	if err != nil {
		return nil, nil, nil, err
	}
	err = r.Read(&pbFormatInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	var qsos []core.QSO
	var station *core.Station
	var contest *core.Contest
	for {
		err := r.Read(&pbEntry)
		if err == io.EOF {
			return qsos, station, contest, nil
		} else if err != nil {
			return nil, nil, nil, err
		}

		if pbQSO := pbEntry.GetQso(); pbQSO != nil {
			qso, err := pbToQSO(*pbQSO)
			if err != nil {
				return nil, nil, nil, err
			}
			qsos = append(qsos, qso)
		}
		if pbStation := pbEntry.GetStation(); pbStation != nil {
			s, err := pbToStation(*pbStation)
			station = &s
			if err != nil {
				return nil, nil, nil, err
			}
		}
		if pbContest := pbEntry.GetContest(); pbContest != nil {
			c, err := pbToContest(*pbContest)
			contest = &c
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}
}

func (f *v1Format) WriteQSO(w pbWriter, qso core.QSO) error {
	pbQSO := qsoToPB(qso)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Qso{Qso: &pbQSO},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteStation(w pbWriter, station core.Station) error {
	pbStation := stationToPB(station)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Station{Station: &pbStation},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteContest(w pbWriter, contest core.Contest) error {
	pbContest := contestToPB(contest)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Contest{Contest: &pbContest},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) Clear(w pbWriter) error {
	err := w.WritePreamble()
	if err != nil {
		return err
	}
	return w.Write(&pb.FileInfo{
		FormatVersion: 1,
	})
}
