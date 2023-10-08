package comborpc

import (
	"context"
	"time"
)

type ServerModel struct {
	endpoint string
	router   map[string]func(ctx context.Context, data string) string
	timeout  time.Duration
	close    bool
}

type ComboRequestModel struct {
	endpoint    string
	requestList []RequestModel
}

type SingleRequestModel struct {
	endpoint    string
	requestList []RequestModel
}

type RequestModel struct {
	Method string `json:"m"`
	Data   string `json:"d"`
}

type ResponseModel struct {
	Error string `json:"e"`
	Data  string `json:"d"`
}
