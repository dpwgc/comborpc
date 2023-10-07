package comborpc

import (
	"context"
	"time"
)

func NewServer(endpoint string, timeout time.Duration) *ServerModel {
	return &ServerModel{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, message string) any),
		timeout:  timeout,
		close:    true,
	}
}

func (s *ServerModel) Add(methodName string, methodFunc func(ctx context.Context, message string) any) {
	s.router[methodName] = methodFunc
}

func (s *ServerModel) Listen() {
	s.close = false
	go tcpListen(s)
}

func (s *ServerModel) Close() {
	s.close = true
}
