package comborpc

import (
	"encoding/json"
)

func NewComboRequest(endpoint string) *ComboRequestModel {
	return &ComboRequestModel{
		endpoint: endpoint,
	}
}

func (c *ComboRequestModel) Add(request RequestModel) *ComboRequestModel {
	c.requestList = append(c.requestList, request)
	return c
}

func (c *ComboRequestModel) Send() ([]ResponseModel, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return nil, err
	}
	res, err := tcpSend(c.endpoint, marshal)
	if err != nil {
		return nil, err
	}
	var resList []ResponseModel
	err = json.Unmarshal(res, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}

func NewSingleRequest(endpoint string) *SingleRequestModel {
	return &SingleRequestModel{
		endpoint: endpoint,
	}
}

func (c *SingleRequestModel) Set(request RequestModel) *SingleRequestModel {
	if len(c.requestList) == 0 {
		c.requestList = append(c.requestList, request)
	} else {
		c.requestList[0] = request
	}
	return c
}

func (c *SingleRequestModel) Send() (ResponseModel, error) {
	marshal, err := json.Marshal(c.requestList)
	if err != nil {
		return ResponseModel{}, err
	}
	res, err := tcpSend(c.endpoint, marshal)
	if err != nil {
		return ResponseModel{}, err
	}
	var resList []ResponseModel
	err = json.Unmarshal(res, &resList)
	if err != nil {
		return ResponseModel{}, err
	}
	if len(resList) == 0 {
		return ResponseModel{}, nil
	}
	return resList[0], nil
}
