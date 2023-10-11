package comborpc

import (
	"net"
	"time"
)

type connect struct {
	object net.Conn
}

type bgService struct {
	object *Router
}

type Router struct {
	endpoint    string
	router      map[string]MethodFunc
	queue       chan *connect
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
