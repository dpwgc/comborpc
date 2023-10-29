package comborpc

import (
	"github.com/vmihailenco/msgpack/v5"
	"time"
)

// NewRouter
// create a new rpc service route
func NewRouter(options RouterOptions) *Router {
	timeout := 1 * time.Minute
	queueLen := 1000
	consumerNum := 30
	if options.Timeout.Milliseconds() >= 1 {
		timeout = options.Timeout
	}
	if options.QueueLen > 0 {
		queueLen = options.QueueLen
	}
	if options.ConsumerNum > 0 {
		consumerNum = options.ConsumerNum
	}
	return &Router{
		endpoint:    options.Endpoint,
		router:      make(map[string]MethodFunc),
		queue:       make(chan *tcpConnect, queueLen),
		consumerNum: consumerNum,
		timeout:     timeout,
		close:       false,
	}
}

// AddMethod
// append the processing method to the service route
func (r *Router) AddMethod(methodName string, methodFunc MethodFunc) *Router {
	r.router[methodName] = methodFunc
	return r
}

// AddMiddleware
// append the middleware
func (r *Router) AddMiddleware(middleware MethodFunc) *Router {
	r.middlewares = append(r.middlewares, middleware)
	return r
}

// AddMiddlewares
// append the middleware
func (r *Router) AddMiddlewares(middlewares ...MethodFunc) *Router {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Run
// start the routing listening service
func (r *Router) Run() error {
	defer func() {
		if r.close {
			return
		}
		r.Close()
	}()
	s := newTcpServe(r)
	for i := 0; i < r.consumerNum; i++ {
		go s.enableConsumer()
	}
	return s.enableListener()
}

// Close
// turn off service routing
func (r *Router) Close() {
	if r.close {
		return
	}
	r.close = true
	_ = r.listener.Close()
	close(r.queue)
}

// Next
// go to the next processing method
func (c *Context) Next() {
	c.index++
	for c.index < len(c.methods) {
		c.methods[c.index](c)
		c.index++
	}
}

// Abort
// stop continuing down execution
func (c *Context) Abort() {
	c.index = len(c.methods) + 1
}

func (c *Context) Read() any {
	return c.input
}

func (c *Context) Bind(v any) error {
	bytes, err := msgpack.Marshal(c.Read())
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(bytes, v)
}

func (c *Context) Write(data any) {
	c.output = data
}
