package ebrpc

import (
	"context"
)

func NewSubscribeServer(endpoint string) *SubscriberServerModel {
	return &SubscriberServerModel{
		serverEndpoint:   endpoint,
		subscriberRouter: make(map[string]map[string]func(ctx context.Context, message string) any),
	}
}

func (s *SubscriberServerModel) Add(topic string, subscriberName string, subscriberFunc func(ctx context.Context, message string) any) {
	if s.subscriberRouter[topic] == nil {
		var initFuncArr = make(map[string]func(ctx context.Context, message string) any)
		s.subscriberRouter[topic] = initFuncArr
	}
	s.subscriberRouter[topic][subscriberName] = subscriberFunc
}

func (s *SubscriberServerModel) Listen() {
	go tcpListen(s)
}
