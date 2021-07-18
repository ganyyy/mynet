package mynet

import (
	"net"
)

//Handler 服务器处理连接Session的接口
type Handler interface {
	HandleSession(*Session)
}

//HandlerFunc Handler接口的实现, 模仿的http.HandlerFunc
type HandlerFunc func(*Session)

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

// 接口类型检查
var _ Handler = HandlerFunc(nil)

//Server 一个服务器结构
type Server struct {
	manager      *Manager
	listener     net.Listener
	protocol     Protocol
	handler      Handler
	sendChanSize int
}

//NewServer 创建一个监听服务器
func NewServer(listener net.Listener, protocol Protocol, sendChanSize int, handler Handler) *Server {
	return &Server{
		manager:      NewManager(),
		listener:     listener,
		protocol:     protocol,
		handler:      handler,
		sendChanSize: sendChanSize,
	}
}

//Listener 获取监听的接口
func (s *Server) Listener() net.Listener {
	return s.listener
}

//Serve 处理连接
func (s *Server) Serve() error {
	for {
		conn, err := Accept(s.Listener())
		if err != nil {
			return err
		}

		go func() {
			codec, err := s.protocol.NewCodec(conn)
			if err != nil {
				conn.Close()
				return
			}
			ses := s.manager.NewSession(codec, s.sendChanSize)
			s.handler.HandleSession(ses)
		}()

	}
}
