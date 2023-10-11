package comborpc

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"net"
	"time"
)

// NewRouter
// create a new rpc service route
func NewRouter(endpoint string, queueLen int, consumerNum int, timeout time.Duration) *Router {
	return &Router{
		endpoint:    endpoint,
		router:      make(map[string]func(ctx *Context)),
		queue:       make(chan net.Conn, queueLen),
		consumerNum: consumerNum,
		timeout:     timeout,
		close:       false,
	}
}

// AddMethod
// append the processing method to the service route
func (r *Router) AddMethod(methodName string, methodFunc func(ctx *Context)) *Router {
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

func (c *Context) ReadString() string {
	return c.input
}

func (c *Context) ReadJson(obj any) error {
	return json.Unmarshal([]byte(c.input), obj)
}

func (c *Context) ReadYaml(obj any) error {
	return yaml.Unmarshal([]byte(c.input), obj)
}

func (c *Context) WriteString(data string) {
	c.output = data
}

func (c *Context) WriteJson(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}

func (c *Context) WriteYaml(obj any) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}
