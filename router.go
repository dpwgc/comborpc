package comborpc

import (
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
func (r *Router) Run() {
	s := newTcpServe(r)
	go s.enableListener()
	for i := 0; i < r.consumerNum; i++ {
		go s.enableConsumer()
	}
}

// Close
// turn off service routing
func (r *Router) Close() error {
	r.close = true
	err := r.listener.Close()
	if err != nil {
		return err
	}
	close(r.queue)
	return nil
}
