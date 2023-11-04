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

type RouterOptions struct {
	Endpoint     string
	QueueLen     int
	MaxGoroutine int
	Timeout      time.Duration
}

type Router struct {
	endpoint    string
	router      map[string]MethodFunc
	queue       chan *tcpConnect
	limit       chan bool
	timeout     time.Duration
	listener    net.Listener
	close       bool
	middlewares []MethodFunc
}

type MethodFunc func(ctx *Context)

type LoadBalanceFunc func(endpoints []string) string

type Context struct {
	RemoteAddr  string
	LocalAddr   string
	CallMethod  string
	CustomCache any
	input       []byte
	output      []byte
	index       int
	methods     []MethodFunc
}

type CallOptions struct {
	Endpoints   []string
	Timeout     time.Duration
	LoadBalance LoadBalanceFunc
}

type callBase struct {
	requests    []request
	endpoints   []string
	timeout     time.Duration
	loadBalance LoadBalanceFunc
	buildError  error
}

type ComboCall struct {
	callBase
}

type SingleCall struct {
	callBase
}

type request struct {
	Method string `msg:"m" msgpack:"m"`
	Data   []byte `msg:"d" msgpack:"d"`
}

type Response struct {
	Error string `msg:"e" msgpack:"e"`
	Data  []byte `msg:"d" msgpack:"d"`
}

type BroadcastResponse struct {
	Endpoint  string
	Error     error
	Responses []Response
}
