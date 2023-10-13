package comborpc

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"sync"
	"time"
)

const TCPHeaderLen int = 8

func newConnect(endpoint string, timeout time.Duration) (*tcpConnect, error) {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return nil, err
	}
	return &tcpConnect{
		conn: conn,
	}, nil
}

func convertedConnect(conn net.Conn) *tcpConnect {
	return &tcpConnect{
		conn: conn,
	}
}

// tcp请求发送
func (c *tcpConnect) send(body []byte) error {
	bodyLen := len(body)
	bodyLenBytes := int64ToBytes(int64(bodyLen), TCPHeaderLen)
	// 发送消息头（数据长度）
	binLen, err := c.conn.Write(bodyLenBytes)
	if err != nil {
		return err
	}
	if binLen != TCPHeaderLen {
		return errors.New("header len not match")
	}
	// 发送消息体（数据包）
	binLen, err = c.conn.Write(body)
	if err != nil {
		return err
	}
	if binLen != bodyLen {
		return errors.New("body len not match")
	}
	return nil
}

func (c *tcpConnect) read() ([]byte, error) {
	// read header
	header := make([]byte, TCPHeaderLen)
	binLen, err := c.conn.Read(header)
	if err != nil {
		return nil, err
	}
	if binLen != TCPHeaderLen {
		return nil, errors.New("header len not match")
	}
	bodyLen := bytesToInt64(header)
	// read body
	body := make([]byte, bodyLen)
	binLen, err = c.conn.Read(body)
	if err != nil {
		return nil, err
	}
	if int64(binLen) != bodyLen {
		return nil, errors.New("body len not match")
	}
	return body, nil
}

func (c *tcpConnect) close() {
	err := c.conn.Close()
	if err != nil {
		log.Println(err)
	}
}

func newTcpServe(r *Router) *tcpServe {
	return &tcpServe{
		router: r,
	}
}

// tcp服务监听
func (s *tcpServe) enableListener() {
	server, err := net.Listen("tcp", s.router.endpoint)
	if err != nil {
		panic(err)
	}
	defer func(server net.Listener) {
		if s.router.close {
			return
		}
		err = server.Close()
		s.router.close = true
		if err != nil {
			panic(err)
		}
	}(server)
	s.router.listener = server
	for {
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			if s.router.close {
				return
			}
			log.Println(err)
			continue
		}
		err = conn.SetDeadline(time.Now().Add(s.router.timeout))
		if err != nil {
			log.Println(err)
			continue
		}
		s.router.queue <- convertedConnect(conn)
	}
}

func (s *tcpServe) enableConsumer() {
	for {
		c, ok := <-s.router.queue
		if !ok && s.router.close {
			return
		}
		func(c *tcpConnect) {
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
func (s *tcpServe) processConnect(c *tcpConnect) error {
	defer c.close()
	body, err := c.read()
	if err != nil {
		return err
	}
	unGzipBody, err := unGzip(body)
	if err != nil {
		return err
	}
	var requestList []Request
	err = yaml.Unmarshal(unGzipBody, &requestList)
	if err != nil {
		return err
	}
	var responseList []Response
	var wg sync.WaitGroup
	wg.Add(len(requestList))
	for i := 0; i < len(requestList); i++ {
		responseList = append(responseList, Response{})
		if s.router.router[requestList[i].Method] == nil {
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
				methods: s.router.middlewares,
			}
			if len(s.router.middlewares) > 0 {
				ctx.methods = append(ctx.methods, s.router.router[requestList[i].Method])
				for ctx.index < len(ctx.methods) {
					ctx.methods[ctx.index](&ctx)
					ctx.index++
				}
			} else {
				s.router.router[requestList[i].Method](&ctx)
			}
			responseList[i].Data = ctx.output
		}(i)
	}
	wg.Wait()
	resultBody, err := yaml.Marshal(responseList)
	if err != nil {
		return err
	}
	gzipResultBody, err := doGzip(resultBody)
	if err != nil {
		return err
	}
	return c.send(gzipResultBody)
}
