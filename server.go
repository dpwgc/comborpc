package comborpc

import (
	"net"
	"time"
)

// NewRouter
// create a new rpc service route
func NewRouter(endpoint string, queueLen int, consumerNum int, timeout time.Duration) *Router {
	return &Router{
		endpoint:    endpoint,
		router:      make(map[string]func(data string) string),
		queue:       make(chan net.Conn, queueLen),
		consumerNum: consumerNum,
		timeout:     timeout,
		close:       false,
	}
}

// AddMethod
// append the processing method to the service route
func (r *Router) AddMethod(methodName string, methodFunc func(data string) string) *Router {
	r.router[methodName] = methodFunc
	return r
}

// ListenAndServe
// start the routing listening service
func (r *Router) ListenAndServe() {
	go enableTcpListener(r)
	for i := 0; i < r.consumerNum; i++ {
		go enableTcpConsumer(r)
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
