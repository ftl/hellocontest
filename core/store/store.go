package store

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
	"github.com/ftl/hellocontest/core/pb"
	"github.com/golang/protobuf/proto"
)

// New returns a new file based Store.
func New(filename string) core.Store {
	return &fileStore{filename}
}

type fileStore struct {
	filename string
}

func (f *fileStore) ReadAll() ([]core.QSO, error) {
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return []core.QSO{}, err
	}

	reader := bytes.NewReader(b)
	bufferedReader := bufio.NewReader(reader)
	qsos := []core.QSO{}
	for {
		qso, err := read(bufferedReader)
		if err == io.EOF {
			return qsos, nil
		} else if err != nil {
			return nil, err
		}
		qsos = append(qsos, qso)
		log.Printf("QSO loaded: %s", qso.String())
	}
}

func read(reader *bufio.Reader) (core.QSO, error) {
	var length int32
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return core.QSO{}, err
	}

	b := make([]byte, length)
	_, err = io.ReadFull(reader, b)
	if err != nil {
		return core.QSO{}, err
	}

	pbQSO := &pb.QSO{}
	err = proto.Unmarshal(b, pbQSO)
	if err != nil {
		return core.QSO{}, err
	}

	qso := core.QSO{}
	qso.Callsign, err = callsign.Parse(pbQSO.Callsign)
	if err != nil {
		return core.QSO{}, err
	}
	qso.Time = time.Unix(pbQSO.Timestamp, 0)
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

func (f *fileStore) Write(qso core.QSO) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return write(file, qso)
}

func write(writer io.Writer, qso core.QSO) error {
	pbQSO := &pb.QSO{
		Callsign:     qso.Callsign.String(),
		Timestamp:    qso.Time.Unix(),
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

	log.Printf("QSO written: %s", qso.String())
	return nil
}

func (f *fileStore) Clear() error {
	file, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	return file.Sync()
}
