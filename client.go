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

// NewComboCall
// create a new composite call
func NewComboCall(options CallOptions) *ComboCall {
	c := &ComboCall{
		callBase: callBase{
			endpoints:   copyStringSlice(options.Endpoints),
			loadBalance: defaultLoadBalance,
			timeout:     1 * time.Minute,
		},
	}
	if options.LoadBalance != nil {
		c.loadBalance = options.LoadBalance
	}
	if options.Timeout.Milliseconds() >= 1 {
		c.timeout = options.Timeout
	}
	return c
}

// AddRequest
// append the request body
func (c *ComboCall) AddRequest(request Request) *ComboCall {
	c.requests = append(c.requests, request)
	return c
}
func (c *ComboCall) AddStringRequest(method string, data string) *ComboCall {
	return c.AddRequest(Request{
		Method: method,
		Data:   data,
	})
}
func (c *ComboCall) AddJsonRequest(method string, v any) *ComboCall {
	data, err := json.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.AddRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *ComboCall) AddYamlRequest(method string, v any) *ComboCall {
	data, err := yaml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.AddRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *ComboCall) AddXmlRequest(method string, v any) *ComboCall {
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
func (c *ComboCall) AddRequests(requests ...Request) *ComboCall {
	c.requests = append(c.requests, requests...)
	return c
}

// Do
// perform a send operation
func (c *ComboCall) Do() ([]Response, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	res, err := tcpCall(c.loadBalance(c.endpoints), c.timeout, data)
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

func (c *ComboCall) Broadcast() ([]BroadcastResponse, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	return tcpBroadcast(c.endpoints, c.timeout, data), nil
}

// NewSingleCall
// create a new single call
func NewSingleCall(options CallOptions) *SingleCall {
	c := &SingleCall{
		callBase: callBase{
			endpoints:   options.Endpoints,
			loadBalance: defaultLoadBalance,
			timeout:     1 * time.Minute,
		},
	}
	if options.LoadBalance != nil {
		c.loadBalance = options.LoadBalance
	}
	if options.Timeout.Milliseconds() >= 1 {
		c.timeout = options.Timeout
	}
	return c
}

// SetRequest
// set a request body
func (c *SingleCall) SetRequest(request Request) *SingleCall {
	if len(c.requests) == 0 {
		c.requests = append(c.requests, request)
	} else {
		c.requests[0] = request
	}
	return c
}
func (c *SingleCall) SetStringRequest(method string, data string) *SingleCall {
	return c.SetRequest(Request{
		Method: method,
		Data:   data,
	})
}
func (c *SingleCall) SetJsonRequest(method string, v any) *SingleCall {
	data, err := json.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.SetRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *SingleCall) SetYamlRequest(method string, v any) *SingleCall {
	data, err := yaml.Marshal(v)
	if err != nil {
		data = []byte("")
	}
	return c.SetRequest(Request{
		Method: method,
		Data:   string(data),
	})
}
func (c *SingleCall) SetXmlRequest(method string, v any) *SingleCall {
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
func (c *SingleCall) Do() (Response, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return Response{}, err
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return Response{}, err
	}
	res, err := tcpCall(c.loadBalance(c.endpoints), c.timeout, data)
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

func (c *SingleCall) Broadcast() ([]BroadcastResponse, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	return tcpBroadcast(c.endpoints, c.timeout, data), nil
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

func defaultLoadBalance(endpoints []string) string {
	if len(endpoints) == 1 {
		return endpoints[0]
	}
	if len(endpoints) == 0 {
		return ""
	}
	rand.Seed(time.Now().Unix())
	return endpoints[rand.Intn(len(endpoints))]
}

func requestValid(requests []Request, endpoints []string) error {
	if len(requests) == 0 {
		return errors.New("requests len = 0")
	}
	if len(endpoints) == 0 {
		return errors.New("endpoints len = 0")
	}
	return nil
}

func tcpBroadcast(endpoints []string, timeout time.Duration, data []byte) []BroadcastResponse {
	var bcResList = make([]BroadcastResponse, len(endpoints))
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for i := 0; i < len(endpoints); i++ {
		bcResList[i].Endpoint = endpoints[i]
		go func(i int) {
			defer wg.Done()
			res, err := tcpCall(endpoints[i], timeout, data)
			if err != nil {
				bcResList[i].Error = err
				return
			}
			var resList []Response
			err = yaml.Unmarshal(res, &resList)
			if err != nil {
				bcResList[i].Error = err
				return
			}
			bcResList[i].Responses = resList
		}(i)
	}
	wg.Wait()
	return bcResList
}

func tcpCall(endpoint string, timeout time.Duration, data []byte) ([]byte, error) {
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
