package comborpc

import (
	"context"
	"time"
)

func NewServer(endpoint string, timeout time.Duration) *ServerModel {
	return &ServerModel{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, data string) string),
		timeout:  timeout,
		close:    true,
	}
}

func (s *ServerModel) Add(methodName string, methodFunc func(ctx context.Context, data string) string) *ServerModel {
	s.router[methodName] = methodFunc
	return s
}

func (s *ServerModel) Listen() {
	s.close = false
	go tcpListen(s)
}

func (s *ServerModel) Close() {
	s.close = true
}
