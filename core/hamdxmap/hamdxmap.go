package hamdxmap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ftl/godxmap"
	"github.com/ftl/hellocontest/core"
)

type mapServer interface {
	Serve() error
	Close() error
	ShowPartialCall(string)
	ShowLoggedCall(string, float64)
}

type HamDXMap struct {
	server mapServer
	closed chan struct{}
}

func NewHamDXMap(port int) *HamDXMap {
	result := &HamDXMap{
		closed: make(chan struct{}),
	}

	if port > 0 {
		result.server = godxmap.NewServer(fmt.Sprintf("localhost:%d", port))
	} else {
		result.server = &nullServer{}
	}

	go func() {
		err := result.server.Serve()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("cannot serve HamDXMap server: %v", err)
		}
		close(result.closed)
	}()

	return result
}

func (m *HamDXMap) Stop() {
	err := m.server.Close()
	if err != nil {
		log.Printf("cannot close HamDXMap server: %v", err)
	}
}

func (m *HamDXMap) WhenStopped(callback func()) {
	go func() {
		<-m.closed
		callback()
	}()
}

func (m *HamDXMap) CallsignEntered(callsign string) {
	m.server.ShowPartialCall(callsign)
}

func (m *HamDXMap) CallsignLogged(callsign string, frequency core.Frequency) {
	m.server.ShowLoggedCall(callsign, float64(frequency/1000.0))
}

type nullServer struct{}

func (s *nullServer) Serve() error                   { return nil }
func (s *nullServer) Close() error                   { return nil }
func (s *nullServer) ShowPartialCall(string)         {}
func (s *nullServer) ShowLoggedCall(string, float64) {}
