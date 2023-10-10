package comborpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

const TCPHeaderLen int = 8

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
	bodyLen := len(body)
	bodyLenBytes := int64ToBytes(int64(bodyLen))
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
	resultBody, err := tcpRead(conn)
	if err != nil {
		return nil, err
	}
	return resultBody, nil
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
func tcpListen(r *Router) {
	server, err := net.Listen("tcp", r.endpoint)
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
		if r.close {
			break
		}
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
		}
		err = tcpProcess(r, conn)
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
			ctx, cancel := context.WithTimeout(context.TODO(), r.timeout)
			defer cancel()
			res := r.router[requestList[i].Method](ctx, requestList[i].Data)
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
	resultBodyLenBytes := int64ToBytes(int64(resultBodyLen))
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
	_, err = conn.Write(resultBody)
	if err != nil {
		return err
	}
	return nil
}
