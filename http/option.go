package http

import "log"

type ServerOption func(*Server)

func ServerWithLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

func ServerWithConfig(cfg Config) ServerOption {
	return func(s *Server) {
		cfg.SetDefaults()
		s.cfg = cfg
		s.name = cfg.Name
		s.addr = cfg.Addr
	}
}
