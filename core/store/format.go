package store

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ftl/hellocontest/core/model"
	"github.com/ftl/hellocontest/core/pb"
)

type fileFormat interface {
	ReadAll(pbReader) ([]model.QSO, *model.Station, *model.Contest, *model.KeyerSettings, []model.QTC, error)
	WriteQSO(pbWriter, model.QSO) error
	WriteStation(pbWriter, model.Station) error
	WriteContest(pbWriter, model.Contest) error
	WriteKeyer(pbWriter, model.KeyerSettings) error
	WriteQTC(pbWriter, model.QTC) error
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

func (f *unknownFormat) ReadAll(pbReader) ([]model.QSO, *model.Station, *model.Contest, *model.KeyerSettings, []model.QTC, error) {
	return nil, nil, nil, nil, nil, f.err
}

func (f *unknownFormat) WriteQSO(pbWriter, model.QSO) error {
	return f.err
}

func (f *unknownFormat) WriteStation(pbWriter, model.Station) error {
	return f.err
}

func (f *unknownFormat) WriteContest(pbWriter, model.Contest) error {
	return f.err
}

func (f *unknownFormat) WriteKeyer(pbWriter, model.KeyerSettings) error {
	return f.err
}

func (f *unknownFormat) WriteQTC(pbWriter, model.QTC) error {
	return f.err
}

func (f *unknownFormat) Clear(pbWriter) error {
	return f.err
}

type v0Format struct {
	filename string
}

func (f *v0Format) ReadAll(r pbReader) ([]model.QSO, *model.Station, *model.Contest, *model.KeyerSettings, []model.QTC, error) {
	qsos := []model.QSO{}
	var pbQSO pb.QSO
	for {
		err := r.Read(&pbQSO)
		if err == io.EOF {
			return qsos, nil, nil, nil, nil, nil
		} else if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		qso, err := pb.ToQSO(&pbQSO)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		qsos = append(qsos, qso)
	}
}

func (f *v0Format) WriteQSO(w pbWriter, qso model.QSO) error {
	pbQSO := pb.QSOToPB(qso)
	return w.Write(pbQSO)
}

func (f *v0Format) WriteStation(pbWriter, model.Station) error {
	log.Println("The V0 file format cannot store station data.")
	return nil
}

func (f *v0Format) WriteContest(pbWriter, model.Contest) error {
	log.Println("The V0 file format cannot store contest data.")
	return nil
}

func (f *v0Format) WriteKeyer(pbWriter, model.KeyerSettings) error {
	log.Println("The V0 file format cannot store keyer data.")
	return nil
}

func (f *v0Format) WriteQTC(pbWriter, model.QTC) error {
	log.Println("The V0 file format cannot store QTC data.")
	return nil
}

func (f *v0Format) Clear(pbWriter) error {
	return nil
}

type v1Format struct {
	filename string
}

func (f *v1Format) ReadAll(r pbReader) ([]model.QSO, *model.Station, *model.Contest, *model.KeyerSettings, []model.QTC, error) {
	var (
		pbFormatInfo pb.FileInfo
		pbEntry      pb.Entry
	)
	_, err := r.ReadPreamble()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	err = r.Read(&pbFormatInfo)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var qsos []model.QSO
	var station *model.Station
	var contest *model.Contest
	var settings *model.KeyerSettings
	var qtcs []model.QTC
	for {
		err := r.Read(&pbEntry)
		if err == io.EOF {
			return qsos, station, contest, settings, qtcs, nil
		} else if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		if pbQSO := pbEntry.GetQso(); pbQSO != nil {
			qso, err := pb.ToQSO(pbQSO)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			qsos = append(qsos, qso)
		}
		if pbStation := pbEntry.GetStation(); pbStation != nil {
			s, err := pb.ToStation(pbStation)
			station = &s
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
		if pbContest := pbEntry.GetContest(); pbContest != nil {
			c, err := pb.ToContest(pbContest)
			contest = &c
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
		if pbKeyer := pbEntry.GetKeyer(); pbKeyer != nil {
			k, err := pb.ToKeyerSettings(pbKeyer)
			settings = &k
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
		if pbQTC := pbEntry.GetQtc(); pbQTC != nil {
			qtc, err := pb.ToQTC(pbQTC)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			qtcs = append(qtcs, qtc)
		}
	}
}

func (f *v1Format) WriteQSO(w pbWriter, qso model.QSO) error {
	pbQSO := pb.QSOToPB(qso)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Qso{Qso: pbQSO},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteStation(w pbWriter, station model.Station) error {
	pbStation := pb.StationToPB(station)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Station{Station: pbStation},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteContest(w pbWriter, contest model.Contest) error {
	pbContest := pb.ContestToPB(contest)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Contest{Contest: pbContest},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteKeyer(w pbWriter, settings model.KeyerSettings) error {
	pbKeyer := pb.KeyerSettingsToPB(settings)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Keyer{Keyer: pbKeyer},
	}
	return w.Write(pbEntry)
}

func (f *v1Format) WriteQTC(w pbWriter, qtc model.QTC) error {
	pbQTC := pb.QTCToPB(qtc)
	pbEntry := &pb.Entry{
		Entry: &pb.Entry_Qtc{Qtc: pbQTC},
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
