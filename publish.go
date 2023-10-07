package ebrpc

import (
	"encoding/json"
)

func Publish(endpoint string, topic string, message string) (string, error) {
	e := eventModel{
		Topic:   topic,
		Message: message,
	}
	marshal, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	res, err := tcpSend(endpoint, marshal)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
