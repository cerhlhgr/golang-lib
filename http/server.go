package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	name      string
	addr      string
	logger    *log.Logger
	server    *http.Server
	handler   http.Handler
	mux       *http.ServeMux
	started   chan struct{}
	serveErr  chan error
	startOnce sync.Once
	errMu     sync.Mutex
	err       error
	cfg       Config
}

func NewServer(addr, name string, opts ...ServerOption) *Server {
	cfg := DefaultConfig()
	cfg.Addr = addr
	cfg.Name = name

	server, err := New(cfg, opts...)
	if err != nil {
		panic(err)
	}

	return server
}

func New(cfg Config, opts ...ServerOption) (*Server, error) {
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	server := &Server{
		name:     cfg.Name,
		addr:     cfg.Addr,
		mux:      http.NewServeMux(),
		started:  make(chan struct{}),
		serveErr: make(chan error, 1),
		cfg:      cfg,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

func (h *Server) SetHandler(handler http.Handler) {
	h.handler = handler
}

func (h *Server) Handle(pattern string, handler http.Handler) {
	h.mux.Handle(pattern, handler)
}

func (h *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h.mux.HandleFunc(pattern, handler)
}

func (h *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", h.addr)
	if err != nil {
		h.setErr(err)
		h.pushServeErr(err)
		return err
	}

	return h.Serve(ln)
}

func (h *Server) Serve(ln net.Listener) error {
	if ln == nil {
		err := errors.New("net.Listener is nil")
		h.setErr(err)
		h.pushServeErr(err)
		return err
	}

	handler := h.handler
	if handler == nil {
		handler = h.mux
	}

	h.server = &http.Server{
		Addr:              ln.Addr().String(),
		Handler:           handler,
		ReadTimeout:       h.cfg.ReadTimeout,
		ReadHeaderTimeout: h.cfg.ReadHeaderTimeout,
		WriteTimeout:      h.cfg.WriteTimeout,
		IdleTimeout:       h.cfg.IdleTimeout,
		MaxHeaderBytes:    h.cfg.MaxHeaderBytes,
	}

	h.markStarted()

	err := h.server.Serve(ln)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	h.setErr(err)
	h.pushServeErr(err)
	return err
}

func (h *Server) WaitsForStarted() error {
	select {
	case err := <-h.serveErr:
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

func (h *Server) Config() Config {
	return h.cfg
}

func (h *Server) setErr(err error) {
	h.errMu.Lock()
	defer h.errMu.Unlock()
	if h.err == nil {
		h.err = err
	}
}

func (h *Server) Err() error {
	h.errMu.Lock()
	defer h.errMu.Unlock()
	return h.err
}

func (h *Server) markStarted() {
	h.startOnce.Do(func() {
		close(h.started)
	})
}

func (h *Server) pushServeErr(err error) {
	select {
	case h.serveErr <- err:
	default:
		if h.logger != nil {
			h.logger.Printf("http server %s error dropped: %v", h.name, err)
		}
	}
}

type Config struct {
	Name              string
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
}

func DefaultConfig() Config {
	return Config{
		Name:              "http",
		Addr:              ":8080",
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

func (c *Config) SetDefaults() {
	def := DefaultConfig()
	if c.Name == "" {
		c.Name = def.Name
	}
	if c.Addr == "" {
		c.Addr = def.Addr
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = def.ReadTimeout
	}
	if c.ReadHeaderTimeout <= 0 {
		c.ReadHeaderTimeout = def.ReadHeaderTimeout
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = def.WriteTimeout
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = def.IdleTimeout
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = def.ShutdownTimeout
	}
	if c.MaxHeaderBytes <= 0 {
		c.MaxHeaderBytes = def.MaxHeaderBytes
	}
}

func (c Config) Validate() error {
	if c.Addr == "" {
		return errors.New("http config: addr is required")
	}
	if c.ReadHeaderTimeout > c.ReadTimeout {
		return fmt.Errorf("http config: read_header_timeout (%s) cannot be greater than read_timeout (%s)", c.ReadHeaderTimeout, c.ReadTimeout)
	}

	return nil
}
