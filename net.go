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

func newConnect(endpoint string, timeout time.Duration) (*connect, error) {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return nil, err
	}
	return &connect{
		object: conn,
	}, nil
}

func convertedConnect(conn net.Conn) *connect {
	return &connect{
		object: conn,
	}
}

// tcp请求发送，并获取响应结果
func (c *connect) sendAndGetResponse(body []byte) ([]byte, error) {
	bodyLen := len(body)
	bodyLenBytes := int64ToBytes(int64(bodyLen), TCPHeaderLen)
	// 发送消息头（数据长度）
	binLen, err := c.object.Write(bodyLenBytes)
	if err != nil {
		return nil, err
	}
	if binLen != TCPHeaderLen {
		return nil, errors.New("header len not match")
	}
	// 发送消息体（数据包）
	binLen, err = c.object.Write(body)
	if err != nil {
		return nil, err
	}
	if binLen != bodyLen {
		return nil, errors.New("body len not match")
	}
	return c.read()
}

// tcp请求发送
func (c *connect) send(body []byte) error {
	bodyLen := len(body)
	bodyLenBytes := int64ToBytes(int64(bodyLen), TCPHeaderLen)
	// 发送消息头（数据长度）
	binLen, err := c.object.Write(bodyLenBytes)
	if err != nil {
		return err
	}
	if binLen != TCPHeaderLen {
		return errors.New("header len not match")
	}
	// 发送消息体（数据包）
	binLen, err = c.object.Write(body)
	if err != nil {
		return err
	}
	if binLen != bodyLen {
		return errors.New("body len not match")
	}
	return nil
}

func (c *connect) read() ([]byte, error) {
	// read header
	header := make([]byte, TCPHeaderLen)
	binLen, err := c.object.Read(header)
	if err != nil {
		return nil, err
	}
	if binLen != TCPHeaderLen {
		return nil, errors.New("header len not match")
	}
	bodyLen := bytesToInt64(header)
	// read body
	body := make([]byte, bodyLen)
	binLen, err = c.object.Read(body)
	if err != nil {
		return nil, err
	}
	if int64(binLen) != bodyLen {
		return nil, errors.New("body len not match")
	}
	return body, nil
}

func (c *connect) close() {
	err := c.object.Close()
	if err != nil {
		log.Println(err)
	}
}

func newBgService(r *Router) *bgService {
	return &bgService{
		object: r,
	}
}

// tcp服务监听
func (s *bgService) enableListener() {
	server, err := net.Listen("tcp", s.object.endpoint)
	if err != nil {
		panic(err)
	}
	defer func(server net.Listener) {
		if s.object.close {
			return
		}
		err = server.Close()
		s.object.close = true
		if err != nil {
			panic(err)
		}
	}(server)
	s.object.listener = server
	for {
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			if s.object.close {
				return
			}
			log.Println(err)
			continue
		}
		err = conn.SetDeadline(time.Now().Add(s.object.timeout))
		if err != nil {
			log.Println(err)
			continue
		}
		s.object.queue <- convertedConnect(conn)
	}
}

func (s *bgService) enableConsumer() {
	for {
		c, ok := <-s.object.queue
		if !ok && s.object.close {
			return
		}
		func(c *connect) {
			defer func() {
				catchErr := recover()
				if catchErr != nil {
					log.Println(catchErr)
				}
			}()
			err := s.processConnect(c)
			if err != nil {
				log.Println(err)
			}
		}(c)
	}
}

// tcp处理函数
func (s *bgService) processConnect(c *connect) error {
	defer c.close()
	body, err := c.read()
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
		if s.object.router[requestList[i].Method] == nil {
			responseList[i].Error = "no method found"
			wg.Done()
			continue
		}
		go func(i int) {
			defer func() {
				handleErr := recover()
				if handleErr != nil {
					responseList[i].Error = fmt.Sprintf("%v", handleErr)
				}
				wg.Done()
			}()
			ctx := Context{
				input:   requestList[i].Data,
				index:   0,
				methods: s.object.middlewares,
			}
			if len(s.object.middlewares) > 0 {
				ctx.methods = append(ctx.methods, s.object.router[requestList[i].Method])
				for ctx.index < len(ctx.methods) {
					ctx.methods[ctx.index](&ctx)
					ctx.index++
				}
			} else {
				s.object.router[requestList[i].Method](&ctx)
			}
			responseList[i].Data = ctx.output
		}(i)
	}
	wg.Wait()
	resultBody, err := json.Marshal(responseList)
	if err != nil {
		return err
	}
	return c.send(resultBody)
}
