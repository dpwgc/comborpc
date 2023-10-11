package comborpc

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"time"
)

// NewRouter
// create a new rpc service route
func NewRouter(endpoint string, queueLen int, consumerNum int, timeout time.Duration) *Router {
	return &Router{
		endpoint:    endpoint,
		router:      make(map[string]MethodFunc),
		queue:       make(chan *connect, queueLen),
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
	bg := newBgService(r)
	go bg.enableListener()
	for i := 0; i < r.consumerNum; i++ {
		go bg.enableConsumer()
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
