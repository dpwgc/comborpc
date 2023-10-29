package comborpc

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
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
func (c *ComboCall) AddRequest(method string, data any) *ComboCall {
	c.requests = append(c.requests, Request{
		Method: method,
		Data:   data,
	})
	return c
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
	return tcpCall(c.loadBalance(c.endpoints), c.timeout, c.requests)
}

func (c *ComboCall) Broadcast() ([]BroadcastResponse, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return nil, err
	}
	return tcpBroadcast(c.endpoints, c.timeout, c.requests), nil
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
func (c *SingleCall) SetRequest(method string, data any) *SingleCall {
	if len(c.requests) == 0 {
		c.requests = append(c.requests, Request{
			Method: method,
			Data:   data,
		})
	} else {
		c.requests[0] = Request{
			Method: method,
			Data:   data,
		}
	}
	return c
}

// Do
// perform a send operation
func (c *SingleCall) Do() (Response, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return Response{}, err
	}
	resList, err := tcpCall(c.loadBalance(c.endpoints), c.timeout, c.requests)
	if err != nil {
		return Response{}, err
	}
	if len(resList) == 0 {
		return Response{}, nil
	}
	return resList[0], nil
}

func (c *SingleCall) DoAndBind(v any) error {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return err
	}
	resList, err := tcpCall(c.loadBalance(c.endpoints), c.timeout, c.requests)
	if err != nil {
		return err
	}
	if len(resList) == 0 {
		return nil
	}
	return resList[0].Bind(v)
}

func (c *SingleCall) Broadcast() ([]BroadcastResponse, error) {
	err := requestValid(c.requests, c.endpoints)
	if err != nil {
		return nil, err
	}
	return tcpBroadcast(c.endpoints, c.timeout, c.requests), nil
}

func (r *Response) Bind(v any) error {
	if len(r.Error) > 0 {
		return errors.New(fmt.Sprintf("response error: %s", r.Error))
	}
	bytes, err := msgpack.Marshal(r.Data)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(bytes, v)
}

func (r *Response) Success() bool {
	if len(r.Error) > 0 {
		return false
	}
	return true
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

func tcpBroadcast(endpoints []string, timeout time.Duration, requests []Request) []BroadcastResponse {
	var bcResList = make([]BroadcastResponse, len(endpoints))
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for i := 0; i < len(endpoints); i++ {
		bcResList[i].Endpoint = endpoints[i]
		go func(i int) {
			defer wg.Done()
			resList, err := tcpCall(endpoints[i], timeout, requests)
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

func tcpCall(endpoint string, timeout time.Duration, requests []Request) ([]Response, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint nil")
	}
	data, err := msgpack.Marshal(requests)
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
