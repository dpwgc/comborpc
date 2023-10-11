package comborpc

import (
	"encoding/json"
	"time"
)

// NewComboRequestBuilder
// create a new composite request builder
func NewComboRequestBuilder(endpoint string, timeout time.Duration) *ComboRequestBuilder {
	return &ComboRequestBuilder{
		endpoint: endpoint,
		timeout:  timeout,
	}
}

// AddRequest
// append the request body
func (b *ComboRequestBuilder) AddRequest(request Request) *ComboRequestBuilder {
	b.requestList = append(b.requestList, request)
	return b
}

// Send
// perform a send operation
func (b *ComboRequestBuilder) Send() ([]Response, error) {
	data, err := json.Marshal(b.requestList)
	if err != nil {
		return nil, err
	}
	res, err := tcpSend(b.endpoint, data, b.timeout)
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
func NewSingleRequestBuilder(endpoint string, timeout time.Duration) *SingleRequestBuilder {
	return &SingleRequestBuilder{
		endpoint: endpoint,
		timeout:  timeout,
	}
}

// SetRequest
// set a request body
func (b *SingleRequestBuilder) SetRequest(request Request) *SingleRequestBuilder {
	if len(b.requestList) == 0 {
		b.requestList = append(b.requestList, request)
	} else {
		b.requestList[0] = request
	}
	return b
}

// Send
// perform a send operation
func (b *SingleRequestBuilder) Send() (Response, error) {
	data, err := json.Marshal(b.requestList)
	if err != nil {
		return Response{}, err
	}
	res, err := tcpSend(b.endpoint, data, b.timeout)
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
