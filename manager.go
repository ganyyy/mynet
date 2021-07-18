package mynet

import "sync"

const (
	sessionMapNum = 32
)

//sessionMap 管理分区
type sessionMap struct {
	mutex    sync.RWMutex
	sessions map[uint64]*Session
	dispose  bool
}

//Manager 会话管理结构
type Manager struct {
	sessionMaps [sessionMapNum]*sessionMap
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

//NewManager 创建一个新的Session管理器
func NewManager() *Manager {
	var m = &Manager{}
	for i := 0; i < sessionMapNum; i++ {
		m.sessionMaps[i] = &sessionMap{
			sessions: map[uint64]*Session{},
		}
	}
	return m
}

//Dispose 关闭管理器
func (m *Manager) Dispose() {
	m.disposeOnce.Do(func() {
		for _, sm := range m.sessionMaps {
			sm.mutex.Lock()
			sm.dispose = true
			for _, ses := range sm.sessions {
				ses.Close()
			}
			sm.mutex.Unlock()
		}
		// 等待执行结束
		m.disposeWait.Wait()
	})
}

//NewSession 基于编码和缓冲队列创建一个新的Session
func (m *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	ses := newSession(m, codec, sendChanSize)
	m.putSession(ses)
	return ses
}

//putSession 加入一个Session
func (m *Manager) putSession(s *Session) {
	var smap = m.sessionMaps[s.id%sessionMapNum]
	if smap.dispose {
		s.Close()
		return
	}
	smap.mutex.Lock()
	smap.sessions[s.id] = s
	smap.mutex.Unlock()
	m.disposeWait.Add(1)
}

//delSession 删除一个Session
func (m *Manager) delSession(s *Session) {
	var smap = m.sessionMaps[s.id%sessionMapNum]
	if smap.dispose {
		return
	}

	smap.mutex.Lock()
	defer smap.mutex.Unlock()
	if _, ok := smap.sessions[s.id]; !ok {
		return
	}
	delete(smap.sessions, s.id)
	m.disposeWait.Done()
}
