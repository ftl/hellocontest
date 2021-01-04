package store

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
	"github.com/ftl/hellocontest/core/pb"
)

// NewFileStore returns a new file based Store.
func NewFileStore(filename string) *FileStore {
	return &FileStore{
		filename: filename,
	}
}

type FileStore struct {
	filename string
}

func (f *FileStore) ReadAllQSOs() ([]core.QSO, error) {
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return []core.QSO{}, err
	}

	reader := bytes.NewReader(b)
	bufferedReader := bufio.NewReader(reader)
	pbReader := &pbReadWriter{reader: bufferedReader}
	qsos := []core.QSO{}
	var pbQSO pb.QSO
	for {
		err = pbReader.Read(&pbQSO)
		if err == io.EOF {
			return qsos, nil
		} else if err != nil {
			return nil, err
		}
		qso, err := pbToQSO(pbQSO)
		if err != nil {
			return nil, err
		}
		qsos = append(qsos, qso)
	}
}

func (f *FileStore) WriteQSO(qso core.QSO) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	pbQSO := qsoToPB(qso)
	pbWriter := &pbReadWriter{writer: file}
	return pbWriter.Write(&pbQSO)
}

func (f *FileStore) Clear() error {
	file, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	return file.Sync()
}

func pbToQSO(pbQSO pb.QSO) (core.QSO, error) {
	var qso core.QSO
	var err error
	qso.Callsign, err = callsign.Parse(pbQSO.Callsign)
	if err != nil {
		return core.QSO{}, err
	}
	qso.Time = time.Unix(pbQSO.Timestamp, 0)
	qso.Frequency = core.Frequency(pbQSO.Frequency)
	qso.Band, err = parse.Band(pbQSO.Band)
	if err != nil {
		return core.QSO{}, err
	}
	qso.Mode, err = parse.Mode(pbQSO.Mode)
	if err != nil {
		return core.QSO{}, err
	}
	qso.MyReport, err = parse.RST(pbQSO.MyReport)
	if err != nil {
		return core.QSO{}, err
	}
	qso.MyNumber = core.QSONumber(pbQSO.MyNumber)
	qso.MyXchange = pbQSO.MyXchange
	qso.TheirReport, err = parse.RST(pbQSO.TheirReport)
	if err != nil {
		return core.QSO{}, err
	}
	qso.TheirNumber = core.QSONumber(pbQSO.TheirNumber)
	qso.TheirXchange = pbQSO.TheirXchange
	qso.LogTimestamp = time.Unix(pbQSO.LogTimestamp, 0)
	return qso, nil
}

func qsoToPB(qso core.QSO) pb.QSO {
	return pb.QSO{
		Callsign:     qso.Callsign.String(),
		Timestamp:    qso.Time.Unix(),
		Frequency:    float64(qso.Frequency),
		Band:         qso.Band.String(),
		Mode:         qso.Mode.String(),
		MyReport:     qso.MyReport.String(),
		MyNumber:     int32(qso.MyNumber),
		MyXchange:    qso.MyXchange,
		TheirReport:  qso.TheirReport.String(),
		TheirNumber:  int32(qso.TheirNumber),
		TheirXchange: qso.TheirXchange,
		LogTimestamp: qso.LogTimestamp.Unix(),
	}
}

type pbReader interface {
	Read(pb protoiface.MessageV1) error
}

type pbWriter interface {
	Write(pb protoiface.MessageV1) error
}

type pbReadWriter struct {
	reader io.Reader
	writer io.Writer
}

func (rw *pbReadWriter) Read(pb protoiface.MessageV1) error {
	var length int32
	err := binary.Read(rw.reader, binary.LittleEndian, &length)
	if err != nil {
		return err
	}

	b := make([]byte, length)
	_, err = io.ReadFull(rw.reader, b)
	if err != nil {
		return err
	}

	return proto.Unmarshal(b, pb)
}

func (rw *pbReadWriter) Write(pb protoiface.MessageV1) error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	length := int32(len(b))
	err = binary.Write(rw.writer, binary.LittleEndian, length)
	if err != nil {
		return err
	}

	_, err = rw.writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}
