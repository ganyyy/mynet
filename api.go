package mynet

import (
	"errors"
	"io"
	"net"
	"time"
)

//Protocol 协议接口, 通过这个构建出一个编码/解码生成器
type Protocol interface {
	NewCodec(rw io.ReadWriter) (Codec, error)
}

//Codec 编码/解码器, 除了基础的消息编码/解码, 还承担了数据的发送/接收
type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

//ClearSendChan 异步队列使用. 用来清理剩余未发送的数据
type ClearSendChan interface {
	ClearSendChan(<-chan interface{})
}

//Listen 创建一个TCP监听的服务器
//TODO 进行抽象, 实现gRPC, websocket等相关的实现
func Listen(network, addr string, protocol Protocol, sendChanSize int, handler Handler) (*Server, error) {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	return NewServer(listener, protocol, sendChanSize, handler), nil
}

//Dial 连接一个TCP接口的服务器
func Dial(network, addr string, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}

//DialTimeout 指定连接超时的连接
func DialTimeout(network, addr string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}

//Accept 接收一个连接
func Accept(listener net.Listener) (net.Conn, error) {
	const (
		EOFErr   = "use of closed network connection"
		MaxDelay = time.Millisecond * 100
	)

	var tempDelay time.Duration

	for {
		conn, err := listener.Accept()
		if err != nil {
			// 临时性的错误, 这里会暂时忽略
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if tempDelay > MaxDelay {
					tempDelay = MaxDelay
				}

				time.Sleep(tempDelay)
				continue
			}
			if errors.Is(net.ErrClosed, err) {
				return nil, io.EOF
			}
		}
		if tempDelay != 0 {
			tempDelay = 0
		}
		return conn, nil
	}

}
