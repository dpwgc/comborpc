package comborpc

import (
	"context"
	"time"
)

func NewRouter(endpoint string, timeout time.Duration) *Router {
	return &Router{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, data string) string),
		timeout:  timeout,
		close:    true,
	}
}

func (s *Router) Add(methodName string, methodFunc func(ctx context.Context, data string) string) *Router {
	s.router[methodName] = methodFunc
	return s
}

func (s *Router) Listen() {
	s.close = false
	go tcpListen(s)
}

func (s *Router) Close() {
	s.close = true
}
