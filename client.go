package comborpc

import (
	"encoding/json"
)

func NewClient(endpoint string) *ClientModel {
	return &ClientModel{
		endpoint: endpoint,
	}
}

func (c *ClientModel) Add(method string, data string) *ClientModel {
	c.requestList = append(c.requestList, requestModel{
		method,
		data,
	})
	return c
}

func (c *ClientModel) Send(endpoint string) (string, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return "", err
	}
	res, err := tcpSend(endpoint, marshal)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
