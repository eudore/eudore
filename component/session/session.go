package session

import (
	"fmt"
	"sync"
	"time"
)

type (
	// Session 定义会话管理对象。
	Session interface {
		SessionLoad(Context) SessionData
		SessionSave(SessionData)
		SessionFlush(string)
	}
	// SessionData 是=定义一个会话的数据。
	SessionData interface {
		Set(key string, value interface{}) error //set session value
		Get(key string) interface{}              //get session value
		Del(key string) error                    //delete session value
		SessionID() string                       //back current sessionID
	}
	// SessionStore
	SessionStore interface {
		Load(string) interface{}
		Save(string, interface{})
		Delete(string)
	}

	// SessionMap 会将数据存储到sync.Map放在内存，不会数据持久化。
	SessionMap struct {
		KeyFync func(Context) string `set:"keyfunc"`
		mem     sync.Map
	}
	SessionStd struct {
		SessionStore
	}
	// SessionDataStd 定义默认使用的会话数据。
	SessionDataStd struct {
		Id   string
		Data map[string]interface{}
	}
)

// DefaultSession 定义默认的会话管理
var (
	DefaultSession Session
)

// NewSessionMap 创建一个SessionMap，使用sync.Map保存数据。
func NewSessionMap() Session {
	return &SessionMap{
		KeyFync: func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}
}

// SessionLoad 获取sessionid加载用户会话数据。
func (store *SessionMap) SessionLoad(ctx Context) SessionData {
	key := store.KeyFync(ctx)
	sess, ok := store.mem.Load(key)
	if ok {
		return sess.(SessionData)
	}
	return &SessionDataStd{
		Id:   key,
		Data: make(map[string]interface{}),
	}
}

// SessionSave 方法实现将一个会话数据保存。
func (store *SessionMap) SessionSave(sess SessionData) {
	store.mem.Store(sess.SessionID(), sess)
}

// SessionFlush 方法实现使用一个sessionid删除一个会话数据。
func (store *SessionMap) SessionFlush(id string) {
	store.mem.Delete(id)
}

// NewSessionCache 创建一个使用Cache作为存储的Session对象。
func NewSessionStd(store SessionStore) Session {
	return &SessionStd{
		SessionStore: store,
		KeyFync: func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}
}

// SessionLoad 方法实现加载一个会话数据，
func (store *SessionStd) SessionLoad(ctx Context) SessionData {
	key := store.KeyFync(ctx)
	sess := store.SessionStore.Load(key)
	if sess != nil {
		return sess.(SessionData)
	}
	return &SessionDataStd{
		Id:   key,
		Data: make(map[string]interface{}),
	}
}

// SessionSave 方法实现将一个会话数据保存。
func (store *SessionStd) SessionSave(sess SessionData) {
	store.SessionStore.Save(sess.SessionID(), sess)
}

// SessionFlush 方法实现使用一个sessionid删除一个会话数据。
func (store *SessionStd) SessionFlush(id string) {
	store.SessionStore.Delete(id)
}

// Set 方法实现SessionDataStd设置数据。
func (sess *SessionDataStd) Set(key string, val interface{}) error {
	sess.Data[key] = val
	return nil
}

// Get 方法从SessionDataStd获得一个数据。
func (sess *SessionDataStd) Get(key string) interface{} {
	return sess.Data[key]
}

// Del 方法删除一个数据。
func (sess *SessionDataStd) Del(key string) error {
	delete(sess.Data, key)
	return nil
}

// SessionID 返回当前会话数据的sessionid。
func (sess *SessionDataStd) SessionID() string {
	return sess.Id
}
