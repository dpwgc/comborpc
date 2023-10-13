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

type LoadBalancingFunc func(endpoints []string) string

type Context struct {
	input   string
	output  string
	index   int
	methods []MethodFunc
}

type ComboRequestClient struct {
	endpoints     []string
	requests      []Request
	timeout       time.Duration
	loadBalancing LoadBalancingFunc
}

type SingleRequestClient struct {
	endpoints     []string
	requests      []Request
	timeout       time.Duration
	loadBalancing LoadBalancingFunc
}

type Request struct {
	Method string `yaml:"m"`
	Data   string `yaml:"d"`
}

type Response struct {
	Error string `yaml:"e"`
	Data  string `yaml:"d"`
}
