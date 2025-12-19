package http

import "log"

type ServerOption func(*Server)

func ServerWithLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.log = logger
	}
}
