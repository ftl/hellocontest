// The package persist allows to store data between sessions of hellocontest.
package session

import (
	fmt "fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"

	"github.com/ftl/hellocontest/core/cfg"
)

type Session struct {
	filename string
	virgin   bool

	state *State
}

func NewDefaultSession() *Session {
	return &Session{
		filename: filepath.Join(cfg.Directory(), "hellocontest.session"),
		virgin:   true,

		state: &State{
			LastFilename:   "current.log",
			SendSpotsToTci: true,
			Esm:            true,
		},
	}
}

func (s *Session) Virgin() bool {
	return s.virgin
}

func (s *Session) LastFilename() string {
	return s.state.LastFilename
}

func (s *Session) SetLastFilename(lastFilename string) error {
	s.state.LastFilename = lastFilename
	return s.Store()
}

func (s *Session) SendSpotsToTci() bool {
	return s.state.SendSpotsToTci
}

func (s *Session) SetSendSpotsToTci(sendSpotsToTci bool) error {
	s.state.SendSpotsToTci = sendSpotsToTci
	return s.Store()
}

func (s *Session) Radio1() string {
	return s.state.Radio1
}

func (s *Session) SetRadio1(value string) error {
	s.state.Radio1 = value
	return s.Store()
}

func (s *Session) Keyer1() string {
	return s.state.Keyer1
}

func (s *Session) SetKeyer1(value string) error {
	s.state.Keyer1 = value
	return s.Store()
}

func (s *Session) ESM() bool {
	return s.state.Esm
}

func (s *Session) SetESM(enabled bool) error {
	s.state.Esm = enabled
	return s.Store()
}

func (s *Session) ESMEnabled(enabled bool) {
	err := s.SetESM(enabled)
	if err != nil {
		log.Printf("cannot store ESM state in the session: %v", err)
	}
}

func (s *Session) Restore() error {
	f, err := os.Open(s.filename)
	if err != nil {
		return fmt.Errorf("cannot open session state for reading: %v", err)
	}
	defer f.Close()

	state, err := readState(f)
	if err != nil {
		return err
	}

	s.state = state
	s.virgin = false

	return nil
}

func readState(r io.Reader) (*State, error) {
	buffer, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read session state: %v", err)
	}
	result := new(State)
	err = proto.Unmarshal(buffer, result)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal session state: %v", err)
	}

	return result, nil
}

func (s *Session) Store() error {
	f, err := os.OpenFile(s.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open session state for writing: %v", err)
	}
	defer f.Close()

	return writeState(f, s.state)
}

func writeState(w io.Writer, state *State) error {
	buffer, err := proto.Marshal(state)
	if err != nil {
		return fmt.Errorf("cannot marshal session state: %v", err)
	}

	_, err = w.Write(buffer)
	if err != nil {
		return fmt.Errorf("cannot write session state: %v", err)
	}

	return nil
}
