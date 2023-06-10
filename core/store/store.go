package store

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/ftl/hellocontest/core"
)

type latestFormat = v1Format

// NewFileStore returns a new file based Store.
func NewFileStore(filename string) *FileStore {
	format := formatFromFile(filename)
	log.Printf("Using %T for %s", format, filename)
	return &FileStore{
		filename: filename,
		format:   format,
	}
}

type FileStore struct {
	filename string
	format   fileFormat
}

func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.filename)
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func (f *FileStore) ReadAll() ([]core.QSO, *core.Station, *core.Contest, *core.KeyerSettings, error) {
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	reader := bytes.NewReader(b)
	bufferedReader := bufio.NewReader(reader)
	pbReader := &pbReadWriter{reader: bufferedReader}
	return f.format.ReadAll(pbReader)
}

func (f *FileStore) WriteQSO(qso core.QSO) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.format.WriteQSO(&pbReadWriter{writer: file}, qso)
}

func (f *FileStore) WriteStation(station core.Station) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.format.WriteStation(&pbReadWriter{writer: file}, station)
}

func (f *FileStore) WriteContest(contest core.Contest) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.format.WriteContest(&pbReadWriter{writer: file}, contest)
}

func (f *FileStore) WriteKeyer(keyer core.KeyerSettings) error {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.format.WriteKeyer(&pbReadWriter{writer: file}, keyer)
}

func (f *FileStore) Clear() error {
	file, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	f.format = new(latestFormat)
	f.format.Clear(&pbReadWriter{writer: file})
	return file.Sync()
}

type pbReader interface {
	Read(pb protoiface.MessageV1) error
	ReadPreamble() (int32, error)
}

type pbWriter interface {
	Write(pb protoiface.MessageV1) error
	WritePreamble() error
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

func (rw *pbReadWriter) ReadPreamble() (int32, error) {
	var preamble int32
	err := binary.Read(rw.reader, binary.LittleEndian, &preamble)
	if err != nil {
		return 0, err
	}
	return preamble, nil
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

func (rw *pbReadWriter) WritePreamble() error {
	preamble := int32(0)
	return binary.Write(rw.writer, binary.LittleEndian, preamble)
}
