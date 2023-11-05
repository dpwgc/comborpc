package comborpc

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"math/rand"
	"sync"
	"time"
)

// ----- ComboCall -----

// NewComboCall
// create a new composite call
func NewComboCall(options CallOptions) *ComboCall {
	c := &ComboCall{
		callBase: callBase{
			headers:     make(map[string]string, 3),
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
func (c *ComboCall) AddRequest(method string, obj any) *ComboCall {
	data, err := msgpack.Marshal(obj)
	if err != nil {
		c.buildError = err
	}
	c.requests = append(c.requests, request{
		Method: method,
		Data:   data,
	})
	return c
}

func (c *ComboCall) PutHeader(key string, value string) *ComboCall {
	c.headers[key] = value
	return c
}

func (c *ComboCall) RemoveHeader(key string) *ComboCall {
	delete(c.headers, key)
	return c
}

// Do
// perform a send operation
func (c *ComboCall) Do() ([]Response, error) {
	err := requestValid(c.requests, c.endpoints, c.buildError)
	if err != nil {
		return nil, err
	}
	return tcpCall(c.loadBalance(c.endpoints), c.timeout, c.requests, c.headers)
}

// ----- SingleCall -----

// NewSingleCall
// create a new single call
func NewSingleCall(options CallOptions) *SingleCall {
	c := &SingleCall{
		callBase: callBase{
			headers:     make(map[string]string, 3),
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
func (c *SingleCall) SetRequest(method string, obj any) *SingleCall {
	data, err := msgpack.Marshal(obj)
	if err != nil {
		c.buildError = err
	}
	if len(c.requests) == 0 {
		c.requests = append(c.requests, request{
			Method: method,
			Data:   data,
		})
	} else {
		c.requests[0] = request{
			Method: method,
			Data:   data,
		}
	}
	return c
}

func (c *SingleCall) PutHeader(key string, value string) *SingleCall {
	c.headers[key] = value
	return c
}

func (c *SingleCall) RemoveHeader(key string) *SingleCall {
	delete(c.headers, key)
	return c
}

// Do
// perform a send operation
func (c *SingleCall) Do() (Response, error) {
	err := requestValid(c.requests, c.endpoints, c.buildError)
	if err != nil {
		return Response{}, err
	}
	resList, err := tcpCall(c.loadBalance(c.endpoints), c.timeout, c.requests, c.headers)
	if err != nil {
		return Response{}, err
	}
	if len(resList) == 0 {
		return Response{}, nil
	}
	return resList[0], nil
}

func (c *SingleCall) DoAndBind(v any) error {
	res, err := c.Do()
	if err != nil {
		return err
	}
	return res.Bind(v)
}

// ----- callBase -----

func (c *callBase) Broadcast() ([]BroadcastResponse, error) {
	err := requestValid(c.requests, c.endpoints, c.buildError)
	if err != nil {
		return nil, err
	}
	return tcpBroadcast(c.endpoints, c.timeout, c.requests, c.headers), nil
}

// ----- Response -----

func (r *Response) Bind(v any) error {
	if !r.Success() {
		return errors.New(fmt.Sprintf("response error: %s", r.Error))
	}
	return msgpack.Unmarshal(r.Data, v)
}

func (r *Response) Success() bool {
	if len(r.Error) > 0 {
		return false
	}
	return true
}

// ----- Other -----

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

func requestValid(requests []request, endpoints []string, buildError error) error {
	if buildError != nil {
		return buildError
	}
	if len(requests) == 0 {
		return errors.New("requests len = 0")
	}
	if len(endpoints) == 0 {
		return errors.New("endpoints len = 0")
	}
	return nil
}

func tcpBroadcast(endpoints []string, timeout time.Duration, requests []request, headers map[string]string) []BroadcastResponse {
	var bcResList = make([]BroadcastResponse, len(endpoints))
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for i := 0; i < len(endpoints); i++ {
		bcResList[i].Endpoint = endpoints[i]
		go func(i int) {
			defer wg.Done()
			resList, err := tcpCall(endpoints[i], timeout, requests, headers)
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

func tcpCall(endpoint string, timeout time.Duration, requests []request, headers map[string]string) ([]Response, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint nil")
	}
	data, err := msgpack.Marshal(fullRequest{
		Headers:  headers,
		Requests: requests,
	})
	if err != nil {
		return nil, err
	}
	c, err := newConnect(endpoint, timeout)
	if err != nil {
		return nil, err
	}
	defer c.close()
	err = c.send(data)
	if err != nil {
		return nil, err
	}
	res, err := c.read()
	if err != nil {
		return nil, err
	}
	var resList []Response
	err = msgpack.Unmarshal(res, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}
