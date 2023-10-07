package ebrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

// tcp发送
func tcpSend(endpoint string, body []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	_, err = conn.Write(body)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 1024*1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// tcp服务监听
func tcpListen(s *SubscriberServerModel) {
	server, err := net.Listen("tcp", s.serverEndpoint)
	if err != nil {
		panic(err)
	}
	defer func(server net.Listener) {
		err = server.Close()
		if err != nil {
			log.Println(err)
		}
	}(server)
	for {
		if s.close {
			break
		}
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
		}
		err = tcpProcess(s, conn)
		if err != nil {
			log.Println(err)
		}
	}
}

// tcp处理函数
func tcpProcess(s *SubscriberServerModel, conn net.Conn) error {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	reader := bufio.NewReader(conn)
	var buf [1024 * 1024]byte
	n, err := reader.Read(buf[:]) // 读取数据
	if err != nil {
		return err
	}
	event := eventModel{}
	err = json.Unmarshal(buf[:n], &event)
	if err != nil {
		return err
	}
	if s.subscriberRouter[event.Topic] == nil {
		return err
	}
	var resAg = responseModel{
		Error: make(map[string]string),
		Data:  make(map[string]any),
	}
	var wg sync.WaitGroup
	wg.Add(len(s.subscriberRouter[event.Topic]))
	for subscriberName, subscriberFunc := range s.subscriberRouter[event.Topic] {
		go func(subscriberName string, subscriberFunc func(ctx context.Context, message string) any) {
			ctx, cancel := context.WithTimeout(context.TODO(), s.processTimeout)
			defer cancel()
			res := subscriberFunc(ctx, event.Message)
			handleErr := recover()
			if handleErr != nil {
				resAg.Error[subscriberName] = fmt.Sprintf("%v", err)
			} else {
				resAg.Data[subscriberName] = res
			}
			wg.Done()
		}(subscriberName, subscriberFunc)
	}
	wg.Wait()
	marshal, err := json.Marshal(resAg)
	if err != nil {
		return err
	}
	_, err = conn.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
