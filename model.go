package comborpc

import (
	"net"
	"time"
)

type tcpConnect struct {
	conn net.Conn
}

type tcpServe struct {
	router *Router
}

type Router struct {
	endpoint    string
	router      map[string]MethodFunc
	queue       chan *tcpConnect
	consumerNum int
	timeout     time.Duration
	listener    net.Listener
	close       bool
	middlewares []MethodFunc
}

type MethodFunc func(ctx *Context)

type Context struct {
	input   string
	output  string
	index   int
	methods []MethodFunc
}

type ComboRequestClient struct {
	endpoint string
	requests []Request
	timeout  time.Duration
}

type SingleRequestClient struct {
	endpoint string
	requests []Request
	timeout  time.Duration
}

type Request struct {
	Method string `json:"m"`
	Data   string `json:"d"`
}

type Response struct {
	Error string `json:"e"`
	Data  string `json:"d"`
}
