package comborpc

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"math/rand"
	"sync"
	"time"
)

// NewComboRequestClient
// create a new composite request client
func NewComboRequestClient() *ComboRequestClient {
	return &ComboRequestClient{
		loadBalancing: loadBalancing,
	}
}

func (c *ComboRequestClient) SetLoadBalancing(loadBalancing LoadBalancingFunc) *ComboRequestClient {
	c.loadBalancing = loadBalancing
	return c
}

func (c *ComboRequestClient) SetEndpoints(endpoints ...string) *ComboRequestClient {
	c.endpoints = endpoints
	return c
}

func (c *ComboRequestClient) SetTimeout(timeout time.Duration) *ComboRequestClient {
	c.timeout = timeout
	return c
}

// AddRequest
// append the request body
func (c *ComboRequestClient) AddRequest(request Request) *ComboRequestClient {
	c.requests = append(c.requests, request)
	return c
}
func (c *ComboRequestClient) AddStringRequest(method string, data string) *ComboRequestClient {
	return c.AddRequest(Request{
		Method: method,
		Data:   data,
	})
}
func (c *ComboRequestClient) AddJsonRequest(method string, v any) *ComboRequestClient {
	data, err := json.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.AddRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *ComboRequestClient) AddYamlRequest(method string, v any) *ComboRequestClient {
	data, err := yaml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.AddRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *ComboRequestClient) AddXmlRequest(method string, v any) *ComboRequestClient {
	data, err := xml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.AddRequest(Request{
		Method: method,
		Data:   string(data),
	})
}

// AddRequests
// append the request body
func (c *ComboRequestClient) AddRequests(requests ...Request) *ComboRequestClient {
	c.requests = append(c.requests, requests...)
	return c
}

// Do
// perform a send operation
func (c *ComboRequestClient) Do() ([]Response, error) {
	if len(c.requests) == 0 {
		return nil, errors.New("requests len = 0")
	}
	if len(c.endpoints) == 0 {
		return nil, errors.New("endpoints len = 0")
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	res, err := tcpRequest(c.loadBalancing(c.endpoints), c.timeout, data)
	if err != nil {
		return nil, err
	}
	var resList []Response
	err = yaml.Unmarshal(res, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}

func (c *ComboRequestClient) Broadcast() error {
	if len(c.requests) == 0 {
		return errors.New("requests len = 0")
	}
	if len(c.endpoints) == 0 {
		return errors.New("endpoints len = 0")
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return err
	}
	var endpointsCopy []string
	copy(endpointsCopy, c.endpoints)
	wg := sync.WaitGroup{}
	wg.Add(len(endpointsCopy))
	for _, v := range endpointsCopy {
		go func(v string) {
			_, _ = tcpRequest(v, c.timeout, data)
			wg.Done()
		}(v)
	}
	wg.Wait()
	return nil
}

// ClearRequests
// clear all request
func (c *ComboRequestClient) ClearRequests() *ComboRequestClient {
	c.requests = nil
	return c
}

// NewSingleRequestClient
// create a new single request client
func NewSingleRequestClient() *SingleRequestClient {
	return &SingleRequestClient{
		loadBalancing: loadBalancing,
	}
}

func (c *SingleRequestClient) SetLoadBalancing(loadBalancing LoadBalancingFunc) *SingleRequestClient {
	c.loadBalancing = loadBalancing
	return c
}

func (c *SingleRequestClient) SetEndpoints(endpoints ...string) *SingleRequestClient {
	c.endpoints = endpoints
	return c
}

func (c *SingleRequestClient) SetTimeout(timeout time.Duration) *SingleRequestClient {
	c.timeout = timeout
	return c
}

// SetRequest
// set a request body
func (c *SingleRequestClient) SetRequest(request Request) *SingleRequestClient {
	if len(c.requests) == 0 {
		c.requests = append(c.requests, request)
	} else {
		c.requests[0] = request
	}
	return c
}
func (c *SingleRequestClient) SetStringRequest(method string, data string) *SingleRequestClient {
	return c.SetRequest(Request{
		Method: method,
		Data:   data,
	})
}
func (c *SingleRequestClient) SetJsonRequest(method string, v any) *SingleRequestClient {
	data, err := json.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.SetRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *SingleRequestClient) SetYamlRequest(method string, v any) *SingleRequestClient {
	data, err := yaml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.SetRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *SingleRequestClient) SetXmlRequest(method string, v any) *SingleRequestClient {
	data, err := xml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.SetRequest(Request{
		Method: method,
		Data:   string(data),
	})
}

// Do
// perform a send operation
func (c *SingleRequestClient) Do() (Response, error) {
	if len(c.requests) == 0 {
		return Response{}, errors.New("request nil")
	}
	if len(c.endpoints) == 0 {
		return Response{}, errors.New("endpoints len = 0")
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return Response{}, err
	}
	res, err := tcpRequest(c.loadBalancing(c.endpoints), c.timeout, data)
	if err != nil {
		return Response{}, err
	}
	var resList []Response
	err = yaml.Unmarshal(res, &resList)
	if err != nil {
		return Response{}, err
	}
	if len(resList) == 0 {
		return Response{}, nil
	}
	return resList[0], nil
}

func (c *SingleRequestClient) Broadcast() error {
	if len(c.requests) == 0 {
		return errors.New("requests len = 0")
	}
	if len(c.endpoints) == 0 {
		return errors.New("endpoints len = 0")
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return err
	}
	var endpointsCopy []string
	copy(endpointsCopy, c.endpoints)
	wg := sync.WaitGroup{}
	wg.Add(len(endpointsCopy))
	for _, v := range endpointsCopy {
		go func(v string) {
			_, _ = tcpRequest(v, c.timeout, data)
			wg.Done()
		}(v)
	}
	wg.Wait()
	return nil
}

func (r *Response) ParseJson(v any) error {
	if len(r.Error) > 0 {
		return errors.New(fmt.Sprintf("response error: %s", r.Error))
	}
	return json.Unmarshal([]byte(r.Data), v)
}
func (r *Response) ParseYaml(v any) error {
	if len(r.Error) > 0 {
		return errors.New(fmt.Sprintf("response error: %s", r.Error))
	}
	return yaml.Unmarshal([]byte(r.Data), v)
}
func (r *Response) ParseXml(v any) error {
	if len(r.Error) > 0 {
		return errors.New(fmt.Sprintf("response error: %s", r.Error))
	}
	return xml.Unmarshal([]byte(r.Data), v)
}

func loadBalancing(endpoints []string) string {
	rand.Seed(time.Now().Unix())
	return endpoints[rand.Intn(len(endpoints))]
}

func tcpRequest(endpoint string, timeout time.Duration, data []byte) ([]byte, error) {
	if timeout.Milliseconds() < 1 {
		timeout = 1 * time.Minute
	}
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint nil")
	}
	gzipData, err := doGzip(data)
	if err != nil {
		return nil, err
	}
	c, err := newConnect(endpoint, timeout)
	if err != nil {
		return nil, err
	}
	defer c.close()
	err = c.send(gzipData)
	if err != nil {
		return nil, err
	}
	res, err := c.read()
	if err != nil {
		return nil, err
	}
	return unGzip(res)
}
