package comborpc

import (
	"context"
	"time"
)

type Router struct {
	endpoint string
	router   map[string]func(ctx context.Context, data string) string
	timeout  time.Duration
	close    bool
}

type ComboRequestBuilder struct {
	endpoint    string
	requestList []Request
}

type SingleRequestBuilder struct {
	endpoint    string
	requestList []Request
}

type Request struct {
	Method string `json:"m"`
	Data   string `json:"d"`
}

type Response struct {
	Error string `json:"e"`
	Data  string `json:"d"`
}
