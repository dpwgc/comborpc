package comborpc

import (
	"context"
	"time"
)

func NewServer(endpoint string, timeout time.Duration) *Server {
	return &Server{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, data string) string),
		timeout:  timeout,
		close:    true,
	}
}

func (s *Server) Add(methodName string, methodFunc func(ctx context.Context, data string) string) *Server {
	s.router[methodName] = methodFunc
	return s
}

func (s *Server) Listen() {
	s.close = false
	go tcpListen(s)
}

func (s *Server) Close() {
	s.close = true
}
