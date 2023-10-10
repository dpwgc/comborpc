package comborpc

import (
	"context"
	"time"
)

// NewRPCRouter
// create a new rpc service route
func NewRPCRouter(endpoint string, timeout time.Duration) *RPCRouter {
	return &RPCRouter{
		endpoint: endpoint,
		router:   make(map[string]func(ctx context.Context, data string) string),
		timeout:  timeout,
		close:    true,
	}
}

// Add
// append the processing method to the service route
func (r *RPCRouter) Add(methodName string, methodFunc func(ctx context.Context, data string) string) *RPCRouter {
	r.router[methodName] = methodFunc
	return r
}

// Listen
// start the routing listening service
func (r *RPCRouter) Listen() {
	r.close = false
	go tcpListen(r)
}

// Close
// turn off service routing
func (r *RPCRouter) Close() {
	r.close = true
}
