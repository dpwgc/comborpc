package ebrpc

import (
	"context"
	"time"
)

type SubscriberServerModel struct {
	serverEndpoint   string
	subscriberRouter map[string]map[string]func(ctx context.Context, message string) any
	processTimeout   time.Duration
	close            bool
}

type eventModel struct {
	Topic   string `json:"t"`
	Message string `json:"m"`
}

type responseModel struct {
	Error map[string]string `json:"e"`
	Data  map[string]any    `json:"d"`
}
