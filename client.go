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

func (c *ClientModel) Send() (map[string]string, map[string]string, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return nil, nil, err
	}
	res, err := tcpSend(c.endpoint, marshal)
	if err != nil {
		return nil, nil, err
	}
	resObj := responseModel{}
	err = json.Unmarshal(res, &resObj)
	if err != nil {
		return nil, nil, err
	}
	return resObj.Data, resObj.Error, nil
}
