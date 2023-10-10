package comborpc

import (
	"encoding/json"
)

// NewComboRequestBuilder
// create a new composite request builder
func NewComboRequestBuilder(endpoint string) *ComboRequestBuilder {
	return &ComboRequestBuilder{
		endpoint: endpoint,
	}
}

// Add
// append the request body
func (c *ComboRequestBuilder) Add(request Request) *ComboRequestBuilder {
	c.requestList = append(c.requestList, request)
	return c
}

// Send
// perform a send operation
func (c *ComboRequestBuilder) Send() ([]Response, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return nil, err
	}
	res, err := tcpSend(c.endpoint, marshal)
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

// NewSingleRequestBuilder
// create a new single request builder
func NewSingleRequestBuilder(endpoint string) *SingleRequestBuilder {
	return &SingleRequestBuilder{
		endpoint: endpoint,
	}
}

// Set
// create a request body
func (c *SingleRequestBuilder) Set(request Request) *SingleRequestBuilder {
	if len(c.requestList) == 0 {
		c.requestList = append(c.requestList, request)
	} else {
		c.requestList[0] = request
	}
	return c
}

// Send
// perform a send operation
func (c *SingleRequestBuilder) Send() (Response, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return Response{}, err
	}
	res, err := tcpSend(c.endpoint, marshal)
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
