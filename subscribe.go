package ebrpc

import (
	"context"
	"time"
)

func NewSubscribeServer(serverEndpoint string, processTimeout time.Duration) *SubscriberServerModel {
	return &SubscriberServerModel{
		serverEndpoint:   serverEndpoint,
		subscriberRouter: make(map[string]map[string]func(ctx context.Context, message string) any),
		processTimeout:   processTimeout,
		close:            true,
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
	s.close = false
	go tcpListen(s)
}

func (s *SubscriberServerModel) Close() {
	s.close = true
}
