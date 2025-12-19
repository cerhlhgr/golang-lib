package http

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

type Server struct {
	name    string
	addr    string
	log     *log.Logger
	server  *http.Server
	handler http.Handler
	mux     *http.ServeMux
	started chan struct{}
	errors  chan error

	listenAddr atomic.Pointer[net.Addr]
}

func NewServer(addr, name string, opts ...ServerOption) *Server {
	server := &Server{
		name:    name,
		addr:    addr,
		mux:     http.NewServeMux(),
		started: make(chan struct{}, 1),
		errors:  make(chan error, 1),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (h *Server) SetHandler(handler http.Handler) {
	h.handler = handler
}

func (h *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", h.addr)
	if err != nil {
		h.errors <- err
		return err
	}

	return h.Serve(ln)
}

func (h *Server) Serve(ln net.Listener) error {
	if ln == nil {
		err := errors.New("net.Listener is nil")
		h.errors <- err
		return err
	}

	addr := ln.Addr()
	h.listenAddr.Store(&addr)

	handler := h.handler
	if handler == nil {
		handler = h.mux
	}

	h.server = &http.Server{
		Addr:              h.addr,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 30,
		IdleTimeout:       time.Second * 120, // keep-alive
		ReadTimeout:       time.Minute * 5,
		WriteTimeout:      time.Minute * 5,
		MaxHeaderBytes:    1 << 20,
	}

	h.started <- struct{}{}

	err := h.server.Serve(ln)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	h.errors <- err
	return err
}

func (h *Server) WaitsForStarted() error {
	select {
	case err := <-h.errors:
		return err
	case <-h.started:
		return nil
	}
}

func (h *Server) Shutdown(ctx context.Context) error {
	if h.server != nil {
		return h.server.Shutdown(ctx)
	}

	return nil
}
