package comborpc

import (
	"context"
	"time"
)

type ServerModel struct {
	endpoint string
	router   map[string]func(ctx context.Context, data string) any
	timeout  time.Duration
	close    bool
}

type ClientModel struct {
	endpoint    string
	requestList []requestModel
}

type requestModel struct {
	Method string `json:"m"`
	Data   string `json:"d"`
}

type responseModel struct {
	Error map[string]string `json:"e"`
	Data  map[string]any    `json:"d"`
}
