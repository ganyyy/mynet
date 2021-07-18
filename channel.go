package mynet

import "sync"

type KEY interface{}

type Channel struct {
	mutex      sync.RWMutex
	sessionMap map[KEY]*Session

	State interface{}
}

func NewChannel() *Channel {
	return &Channel{
		sessionMap: make(map[KEY]*Session),
	}
}

func (c *Channel) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.sessionMap)
}

func (c *Channel) Fetch(callback func(*Session)) {
	c.mutex.RLock()
	for _, s := range c.sessionMap {
		callback(s)
	}
	c.mutex.RUnlock()
}

func (c *Channel) Get(key KEY) *Session {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.sessionMap[key]
}

//Put 加入一个Session
func (c *Channel) Put(key KEY, session *Session) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// 存在就先执行移除
	if old, exist := c.sessionMap[key]; exist {
		c.remove(key, old)
	}

	// 增加关闭的回调
	session.AddCloseCallback(c, key, func() {
		c.Remove(key)
	})
	c.sessionMap[key] = session

}

//Remove 移除指定key对应的Session
func (c *Channel) Remove(key KEY) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if old, exist := c.sessionMap[key]; exist {
		c.remove(key, old)
		return true
	}
	return false
}

//remove 移除一个key对应的Session. 调用这个方法之前请在外边持有写锁
func (c *Channel) remove(key KEY, session *Session) {
	session.RemoveCloseCallback(c, key)
	delete(c.sessionMap, key)
}

//FetchAndRemove 关闭所有相关的session
func (c *Channel) FetchAndRemove(callback func(*Session)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, ses := range c.sessionMap {
		c.remove(key, ses)
		callback(ses)
	}
}

//Close 关闭Channel
func (c *Channel) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, ses := range c.sessionMap {
		c.remove(key, ses)
	}
}
