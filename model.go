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
	Endpoint    string
	QueueLen    int
	ConsumerNum int
	Timeout     time.Duration
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

type LoadBalanceFunc func(endpoints []string) string

type Context struct {
	remoteAddr string
	localAddr  string
	callMethod string
	input      string
	output     string
	index      int
	methods    []MethodFunc
	cache      any
}

type CallOptions struct {
	Endpoints   []string
	Timeout     time.Duration
	LoadBalance LoadBalanceFunc
}

type callBase struct {
	requests    []Request
	endpoints   []string
	timeout     time.Duration
	loadBalance LoadBalanceFunc
}

type ComboCall struct {
	callBase
}

type SingleCall struct {
	callBase
}

type Request struct {
	Method string `yaml:"m"`
	Data   string `yaml:"d,omitempty"`
}

type Response struct {
	Error string `yaml:"e,omitempty"`
	Data  string `yaml:"d,omitempty"`
}

type BroadcastResponse struct {
	Endpoint  string
	Error     error
	Responses []Response
}
