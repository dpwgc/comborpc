package ebrpc

import "context"

type SubscriberServerModel struct {
	serverEndpoint   string
	subscriberRouter map[string]map[string]func(ctx context.Context, message string) any
}

type eventModel struct {
	Topic   string `json:"t"`
	Message string `json:"m"`
}
