package comborpc

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
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
	gzipBody, err := doGzip(body)
	if err != nil {
		return err
	}
	bodyLen := len(gzipBody)
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
	binLen, err = c.conn.Write(gzipBody)
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
	unGzipBody, err := unGzip(body)
	if err != nil {
		return nil, err
	}
	return unGzipBody, nil
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
func (s *tcpServe) enableListener() error {
	server, err := net.Listen("tcp", s.router.endpoint)
	if err != nil {
		return err
	}
	s.router.listener = server
	for {
		// 接收tcp数据
		conn, err := server.Accept()
		if err != nil {
			if s.router.close {
				return nil
			}
			log.Println(err)
			continue
		}
		err = conn.SetDeadline(time.Now().Add(s.router.timeout))
		if err != nil {
			if s.router.close {
				return nil
			}
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
		s.router.limit <- true
		go func(c *tcpConnect) {
			defer func() {
				catchErr := recover()
				if catchErr != nil {
					log.Println(catchErr)
				}
				<-s.router.limit
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
	var fr fullRequest
	err = msgpack.Unmarshal(body, &fr)
	if err != nil {
		return err
	}
	requestsLen := len(fr.Requests)
	var responseList = make([]Response, requestsLen)
	var wg sync.WaitGroup
	wg.Add(requestsLen)
	for i := 0; i < requestsLen; i++ {
		go func(i int) {
			defer func() {
				handleErr := recover()
				if handleErr != nil {
					responseList[i].Error = fmt.Sprintf("%v", handleErr)
				}
				wg.Done()
			}()
			if s.router.router[fr.Requests[i].Method] == nil {
				responseList[i].Error = "no method found"
				return
			}
			ctx := Context{
				RemoteAddr: c.conn.RemoteAddr().String(),
				LocalAddr:  c.conn.LocalAddr().String(),
				CallMethod: fr.Requests[i].Method,
				headers:    fr.Headers,
				input:      fr.Requests[i].Data,
				index:      0,
				methods:    copyMethodFuncSlice(s.router.middlewares),
			}
			if len(s.router.middlewares) > 0 {
				ctx.methods = append(ctx.methods, s.router.router[fr.Requests[i].Method])
				for ctx.index < len(ctx.methods) {
					ctx.methods[ctx.index](&ctx)
					ctx.index++
				}
			} else {
				s.router.router[fr.Requests[i].Method](&ctx)
			}
			responseList[i].Data = ctx.output
		}(i)
	}
	wg.Wait()
	resBody, err := msgpack.Marshal(responseList)
	if err != nil {
		return err
	}
	return c.send(resBody)
}
