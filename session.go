package mynet

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrSessionClosed  = errors.New("session closed")
	ErrSessionBlocked = errors.New("session blocked")
)

var (
	globalSessionId uint64
)

//closeCallback 关闭的回调函数. 链式结构
type closeCallback struct {
	Handler interface{}
	Key     interface{}
	Func    func()
	Next    *closeCallback
}

//Session 抽象的连接对象
type Session struct {
	id       uint64           // 当前ses的id
	codec    Codec            // 编码接口
	manager  *Manager         // 持有的管理器引用
	sendChan chan interface{} // 异步的消息发送队列

	recvMutex sync.Mutex   // 数据接收锁
	sendMutex sync.RWMutex // 数据发送锁

	closeFlag  int32         // 关闭标记
	closeChan  chan struct{} // 关闭通知
	closeMutex sync.Mutex    // 关闭的锁

	firstCloseCallback *closeCallback
	lastCloseCallback  *closeCallback

	State interface{} // 当前Session的状态信息
}

//NewSession 创建一个新的Session
func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

//newSession API的封装
func newSession(m *Manager, codec Codec, sendChanSize int) *Session {
	var ses = &Session{
		id:        atomic.AddUint64(&globalSessionId, 1),
		codec:     codec,
		manager:   m,
		closeChan: make(chan struct{}),
	}

	if sendChanSize > 0 {
		// 如果大于0, 说明通过chan启动一个异步的消息处理
		ses.sendChan = make(chan interface{}, sendChanSize)
		go ses.sendLoop()

	}
	return ses
}

func (s *Session) sendLoop() {
	for {
		//TODO 返回原因的log
		select {
		case msg, ok := <-s.sendChan:
			if !ok {
				return
			}
			if err := s.codec.Send(msg); err != nil {
				return
			}
		case <-s.closeChan:
			return
		}
	}
}

func (s *Session) Send(msg interface{}) error {
	if s.sendChan == nil {
		// 非异步的Session
		if s.IsClosed() {
			return ErrSessionClosed
		}

		s.sendMutex.Lock()
		defer s.sendMutex.Unlock()
		err := s.codec.Send(msg)
		if err != nil {
			s.Close()
		}
		return err
	}

	// 使用异步chan 需要保证chan是可用的
	s.sendMutex.RLock()
	defer s.sendMutex.RUnlock()
	if s.IsClosed() {
		return ErrSessionClosed
	}

	select {
	case s.sendChan <- msg:
		return nil
	default:
		s.Close()
		return ErrSessionBlocked
	}
}

//Receive 接收一条数据
func (s *Session) Receive() (interface{}, error) {
	s.recvMutex.Lock()
	defer s.recvMutex.Unlock()
	msg, err := s.codec.Receive()
	if err != nil {
		s.Close()
	}
	return msg, err
}

//IsClosed 当前Session是否已经关闭
func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closeFlag) == 1
}

//Close 关闭当前Session
func (s *Session) Close() error {
	if !atomic.CompareAndSwapInt32(&s.closeFlag, 0, 1) {
		return ErrSessionClosed
	}

	close(s.closeChan)

	if s.sendChan != nil {
		// 这个关闭应该也不是必须的.
		s.sendMutex.Lock()
		close(s.sendChan)
		s.sendMutex.Unlock()

		// 是否有必要释放chan中的消息?
		// 感觉完全没必要. 失去引用自然就被GC了
		if clear, ok := s.codec.(ClearSendChan); ok {
			// 如果有回收逻辑的话, 清理一下
			clear.ClearSendChan(s.sendChan)
		}
	}

	go func() {
		// 关闭前的清理回调函数
		s.invokeCloseCallbacks()
		if s.manager != nil {
			s.manager.delSession(s)
		}
	}()

	err := s.codec.Close()

	return err
}

//AddCloseCallback 注册关闭回调函数
func (s *Session) AddCloseCallback(handler, key interface{}, callback func()) {
	if s.IsClosed() {
		return
	}

	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()

	var newCallback = &closeCallback{
		Handler: handler,
		Key:     key,
		Func:    callback,
	}
	if s.firstCloseCallback == nil {
		s.firstCloseCallback = newCallback
	} else {
		s.lastCloseCallback.Next = newCallback
	}
	s.lastCloseCallback = newCallback
}

//RemoveCloseCallback 移除指定key-handle 的回调函数
func (s *Session) RemoveCloseCallback(key, handle interface{}) {
	if s.IsClosed() {
		return
	}

	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()

	var prev *closeCallback
	//
	for callback := s.firstCloseCallback; callback != nil; callback = callback.Next {
		if callback.Handler == handle && callback.Key == key {
			if s.firstCloseCallback == callback {
				s.firstCloseCallback = s.firstCloseCallback.Next
			} else {
				prev.Next = callback.Next
			}
			if s.lastCloseCallback == callback {
				s.lastCloseCallback = prev
			}
			return
		}
		prev = callback
	}
}

func (s *Session) invokeCloseCallbacks() {
	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()

	for callback := s.firstCloseCallback; callback != nil; callback = callback.Next {
		callback.Func()
	}
}
