package ebrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
)

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
	buf := make([]byte, 1024)
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
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
		}
		tcpProcess(s, conn)
	}
}

// tcp处理函数
func tcpProcess(s *SubscriberServerModel, conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	reader := bufio.NewReader(conn)
	var buf [1024]byte
	n, err := reader.Read(buf[:]) // 读取数据
	if err != nil {
		log.Println(err)
		return
	}
	event := eventModel{}
	err = json.Unmarshal(buf[:n], &event)
	if err != nil {
		log.Println(err)
		return
	}
	if s.subscriberRouter[event.Topic] == nil {
		return
	}
	var resAg = make(map[string]map[string]any)
	for subscriberName, subscriberFunc := range s.subscriberRouter[event.Topic] {
		go func(subscriberName string, subscriberFunc func(ctx context.Context, message string) any) {
			res := make(map[string]any)
			d := subscriberFunc(context.Background(), event.Message)
			err := recover()
			if err != nil {
				res["e"] = err
			} else {
				res["d"] = d
			}
			resAg[subscriberName] = res
		}(subscriberName, subscriberFunc)
	}
	marshal, err := json.Marshal(resAg)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = conn.Write(marshal)
	if err != nil {
		log.Println(err)
		return
	}
}
