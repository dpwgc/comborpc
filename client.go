package comborpc

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"time"
)

// NewComboRequestClient
// create a new composite request client
func NewComboRequestClient(endpoint string, timeout time.Duration) *ComboRequestClient {
	return &ComboRequestClient{
		endpoint: endpoint,
		timeout:  timeout,
	}
}

func (c *ComboRequestClient) EditEndpoint(endpoint string) *ComboRequestClient {
	c.endpoint = endpoint
	return c
}

func (c *ComboRequestClient) EditTimeout(timeout time.Duration) *ComboRequestClient {
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
	data, err := json.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	res, err := tcpRequest(c.endpoint, c.timeout, data)
	if err != nil {
		return nil, err
	}
	var resList []Response
	err = json.Unmarshal(res, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}

// NewSingleRequestClient
// create a new single request client
func NewSingleRequestClient(endpoint string, timeout time.Duration) *SingleRequestClient {
	return &SingleRequestClient{
		endpoint: endpoint,
		timeout:  timeout,
	}
}

func (c *SingleRequestClient) EditEndpoint(endpoint string) *SingleRequestClient {
	c.endpoint = endpoint
	return c
}

func (c *SingleRequestClient) EditTimeout(timeout time.Duration) *SingleRequestClient {
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
	data, err := json.Marshal(c.requests)
	if err != nil {
		return Response{}, err
	}
	res, err := tcpRequest(c.endpoint, c.timeout, data)
	if err != nil {
		return Response{}, err
	}
	var resList []Response
	err = json.Unmarshal(res, &resList)
	if err != nil {
		return Response{}, err
	}
	if len(resList) == 0 {
		return Response{}, nil
	}
	return resList[0], nil
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

func tcpRequest(endpoint string, timeout time.Duration, data []byte) ([]byte, error) {
	c, err := newConnect(endpoint, timeout)
	if err != nil {
		return nil, err
	}
	defer c.close()
	err = c.send(data)
	if err != nil {
		return nil, err
	}
	return c.read()
}
