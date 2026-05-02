package remote

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
)

type ActionDispatcher interface {
	DoAction(id string) error
}

type Keyer interface {
	SendTextWithTemplate(text string) error
}

type Server struct {
	dispatcher ActionDispatcher
	keyer      Keyer
	port       int
	runner     func(func())
	httpServer *http.Server
}

func NewServer(dispatcher ActionDispatcher, keyer Keyer, port int, runner func(func())) *Server {
	return &Server{
		dispatcher: dispatcher,
		keyer:      keyer,
		port:       port,
		runner:     runner,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/do", s.handleDo)
	mux.HandleFunc("/send", s.handleSend)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("remote server error: %v", err)
		}
	}()

	log.Printf("remote server listening on %s", s.httpServer.Addr)
	return nil
}

func (s *Server) Stop() {
	if s.httpServer == nil {
		return
	}
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("remote server shutdown error: %v", err)
	}
}

func (s *Server) runOnMainThread(f func() error) error {
	result := make(chan error, 1)
	s.runner(func() {
		result <- f()
	})
	return <-result
}

func (s *Server) handleDo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	action := r.URL.Query().Get("action")
	if action == "" {
		http.Error(w, "missing action parameter", http.StatusBadRequest)
		return
	}
	if err := s.runOnMainThread(func() error { return s.dispatcher.DoAction(action) }); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "missing text parameter", http.StatusBadRequest)
		return
	}
	if err := s.runOnMainThread(func() error { return s.keyer.SendTextWithTemplate(text) }); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
