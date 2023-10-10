package comborpc

import (
	"context"
	"net"
	"time"
)

type Router struct {
	endpoint    string
	router      map[string]func(ctx context.Context, data string) string
	queue       chan net.Conn
	consumerNum int
	timeout     time.Duration
	listener    net.Listener
	close       bool
}

type ComboRequestBuilder struct {
	endpoint    string
	requestList []Request
}

type SingleRequestBuilder struct {
	endpoint    string
	requestList []Request
}

type Request struct {
	Method string `json:"m"`
	Data   string `json:"d"`
}

type Response struct {
	Error string `json:"e"`
	Data  string `json:"d"`
}
