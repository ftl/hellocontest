package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	logger "log"
	"os"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/pb"
	"github.com/golang/protobuf/proto"
)

// Store allows to read and write log entries.
type Store interface {
	Reader
	Writer
}

// NewFileStore returns a new file based Store.
func NewFileStore(filename string) Store {
	return &fileStore{filename}
}

type fileStore struct {
	filename string
}

func (f *fileStore) ReadAll() ([]QSO, error) {
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return []QSO{}, err
	}

	reader := bytes.NewReader(b)
	bufferedReader := bufio.NewReader(reader)
	qsos := []QSO{}
	for {
		qso, err := read(bufferedReader)
		if err == io.EOF {
			return qsos, nil
		} else if err != nil {
			return nil, err
		}
		qsos = append(qsos, qso)
		logger.Printf("QSO loaded: %s", qso.String())
	}
}

func read(reader *bufio.Reader) (QSO, error) {
	var length int32
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return QSO{}, err
	}

	b := make([]byte, length)
	_, err = io.ReadFull(reader, b)
	if err != nil {
		return QSO{}, err
	}

	pbQSO := &pb.QSO{}
	err = proto.Unmarshal(b, pbQSO)
	if err != nil {
		return QSO{}, err
	}

	qso := QSO{}
	qso.Callsign, err = callsign.Parse(pbQSO.Callsign)
	if err != nil {
		return QSO{}, err
	}
	qso.Time = time.Unix(pbQSO.Timestamp, 0)
	qso.Band, err = ParseBand(pbQSO.Band)
	if err != nil {
		return QSO{}, err
	}
	qso.Mode, err = ParseMode(pbQSO.Mode)
	if err != nil {
		return QSO{}, err
	}
	qso.MyReport, err = ParseRST(pbQSO.MyReport)
	if err != nil {
		return QSO{}, err
	}
	qso.MyNumber = QSONumber(pbQSO.MyNumber)
	qso.TheirReport, err = ParseRST(pbQSO.TheirReport)
	if err != nil {
		return QSO{}, err
	}
	qso.TheirNumber = QSONumber(pbQSO.TheirNumber)
	qso.LogTimestamp = time.Unix(pbQSO.LogTimestamp, 0)
	return qso, nil
}

func (f *fileStore) Write(qso QSO) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return write(file, qso)
}

func write(writer io.Writer, qso QSO) error {
	pbQSO := &pb.QSO{
		Callsign:     qso.Callsign.String(),
		Timestamp:    qso.Time.Unix(),
		Band:         qso.Band.String(),
		Mode:         qso.Mode.String(),
		MyReport:     qso.MyReport.String(),
		MyNumber:     int32(qso.MyNumber),
		TheirReport:  qso.TheirReport.String(),
		TheirNumber:  int32(qso.TheirNumber),
		LogTimestamp: qso.LogTimestamp.Unix(),
	}

	b, err := proto.Marshal(pbQSO)
	if err != nil {
		return err
	}

	length := int32(len(b))
	err = binary.Write(writer, binary.LittleEndian, length)
	if err != nil {
		return err
	}

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	logger.Printf("QSO written: %s", qso.String())
	return nil
}
