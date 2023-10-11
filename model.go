package comborpc

import (
	"net"
	"time"
)

type Router struct {
	endpoint    string
	router      map[string]func(ctx *Context)
	queue       chan net.Conn
	consumerNum int
	timeout     time.Duration
	listener    net.Listener
	close       bool
	middlewares []func(ctx *Context)
}

type Context struct {
	input   string
	output  string
	index   int
	methods []func(ctx *Context)
}

type ComboRequestBuilder struct {
	endpoint string
	requests []Request
	timeout  time.Duration
}

type SingleRequestBuilder struct {
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
