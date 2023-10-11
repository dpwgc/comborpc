package comborpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const TCPHeaderLen int = 8

// tcp发送
func tcpSend(endpoint string, body []byte, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	bodyLen := len(body)
	bodyLenBytes := int64ToBytes(int64(bodyLen), TCPHeaderLen)
	// 发送消息头（数据长度）
	binLen, err := conn.Write(bodyLenBytes)
	if err != nil {
		return nil, err
	}
	if binLen != TCPHeaderLen {
		return nil, errors.New("header len not match")
	}
	// 发送消息体（数据包）
	binLen, err = conn.Write(body)
	if err != nil {
		return nil, err
	}
	if binLen != bodyLen {
		return nil, errors.New("body len not match")
	}
	return tcpRead(conn)
}

func tcpRead(conn net.Conn) ([]byte, error) {
	// read header
	header := make([]byte, TCPHeaderLen)
	binLen, err := conn.Read(header)
	if err != nil {
		return nil, err
	}
	if binLen != TCPHeaderLen {
		return nil, errors.New("header len not match")
	}
	bodyLen := bytesToInt64(header)
	// read body
	body := make([]byte, bodyLen)
	binLen, err = conn.Read(body)
	if err != nil {
		return nil, err
	}
	if int64(binLen) != bodyLen {
		return nil, errors.New("body len not match")
	}
	return body, nil
}

// tcp服务监听
func enableTcpListener(r *Router) {
	server, err := net.Listen("tcp", r.endpoint)
	if err != nil {
		panic(err)
	}
	defer func(server net.Listener) {
		if r.close {
			return
		}
		err = server.Close()
		r.close = true
		if err != nil {
			panic(err)
		}
	}(server)
	r.listener = server
	for {
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			return
		}
		err = conn.SetDeadline(time.Now().Add(r.timeout))
		if err != nil {
			return
		}
		if err != nil {
			if r.close {
				return
			}
			log.Println(err)
		} else {
			r.queue <- conn
		}
	}
}

func enableTcpConsumer(r *Router) {
	for {
		consumeErr := recover()
		if consumeErr != nil {
			log.Println(consumeErr)
		}
		conn, ok := <-r.queue
		if !ok {
			if r.close {
				return
			}
		}
		err := tcpProcess(r, conn)
		if err != nil {
			log.Println(err)
		}
	}
}

// tcp处理函数
func tcpProcess(r *Router, conn net.Conn) error {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	body, err := tcpRead(conn)
	if err != nil {
		return err
	}
	var requestList []Request
	err = json.Unmarshal(body, &requestList)
	if err != nil {
		return err
	}
	var responseList []Response
	var wg sync.WaitGroup
	wg.Add(len(requestList))
	for i := 0; i < len(requestList); i++ {
		responseList = append(responseList, Response{})
		if r.router[requestList[i].Method] == nil {
			responseList[i].Error = "no method found"
			wg.Done()
			continue
		}
		go func(i int) {
			res := r.router[requestList[i].Method](requestList[i].Data)
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
	resultBody, err := json.Marshal(responseList)
	if err != nil {
		return err
	}
	resultBodyLen := len(resultBody)
	resultBodyLenBytes := int64ToBytes(int64(resultBodyLen), TCPHeaderLen)
	// 发送消息头（数据长度）
	binLen, err := conn.Write(resultBodyLenBytes)
	if err != nil {
		return err
	}
	if binLen != TCPHeaderLen {
		return errors.New("header len not match")
	}
	// 发送消息体（数据包）
	binLen, err = conn.Write(resultBody)
	if err != nil {
		return err
	}
	if binLen != resultBodyLen {
		return errors.New("body len not match")
	}
	return nil
}
