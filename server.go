package comborpc

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"
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

func (c *Context) GetCallMethod() string {
	return c.callMethod
}

func (c *Context) GetShareData() any {
	return c.shareData.Load()
}

func (c *Context) PutShareData(v any) {
	c.shareData.Store(v)
}

func (c *Context) ReadString() string {
	return c.input
}

func (c *Context) ReadJson(v any) error {
	return json.Unmarshal([]byte(c.input), v)
}

func (c *Context) ReadYaml(v any) error {
	return yaml.Unmarshal([]byte(c.input), v)
}

func (c *Context) ReadXml(v any) error {
	return xml.Unmarshal([]byte(c.input), v)
}

func (c *Context) WriteString(data string) {
	c.output = data
}

func (c *Context) WriteJson(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}

func (c *Context) WriteYaml(v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}

func (c *Context) WriteXml(v any) error {
	data, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}
