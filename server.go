package comborpc

import (
	"context"
	"time"
)

// NewRouter
// create a new rpc service route
func NewRouter(endpoint string, timeout time.Duration) *Router {
	return &Router{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, data string) string),
		timeout:  timeout,
		close:    true,
	}
}

// Add
// append the processing method to the service route
func (r *Router) Add(methodName string, methodFunc func(ctx context.Context, data string) string) *Router {
	r.router[methodName] = methodFunc
	return r
}

// Listen
// start the routing listening service
func (r *Router) Listen() {
	r.close = false
	go tcpListen(r)
}

// Close
// turn off service routing
func (r *Router) Close() {
	r.close = true
}
