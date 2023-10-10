package comborpc

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
func tcpSend(endpoint string, data []byte) ([]byte, error) {
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
	_, err = conn.Write(data)
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
func tcpListen(s *Router) {
	server, err := net.Listen("tcp", s.endpoint)
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
func tcpProcess(s *Router, conn net.Conn) error {
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
	var requestList []Request
	err = json.Unmarshal(buf[:n], &requestList)
	if err != nil {
		return err
	}
	var responseList []Response
	var wg sync.WaitGroup
	wg.Add(len(requestList))
	for i := 0; i < len(requestList); i++ {
		responseList = append(responseList, Response{})
		if s.router[requestList[i].Method] == nil {
			responseList[i].Error = "no method found"
			wg.Done()
			continue
		}
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
			defer cancel()
			res := s.router[requestList[i].Method](ctx, requestList[i].Data)
			handleErr := recover()
			if handleErr != nil {
				responseList[i].Error = fmt.Sprintf("%v", handleErr)
			} else {
				responseList[i].Data = res
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	marshal, err := json.Marshal(responseList)
	if err != nil {
		return err
	}
	_, err = conn.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
