package session

import (
	"encoding/gob"
	"github.com/eudore/eudore"
	"sync"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
}

type (
	// Session 定义会话管理对象。
	Session interface {
		GetSessionId(eudore.Context) string
		SessionLoad(eudore.Context) map[string]interface{}
		SessionSave(eudore.Context, map[string]interface{})
		SessionFlush(eudore.Context)
	}
	// SessionStore 定义存储对象接口。
	SessionStore interface {
		Load(string) interface{}
		Save(string, interface{})
		Delete(string)
	}
	// SessionStd 定义默认使用的Session。
	SessionStd struct {
		SessionStore
		Maxage  int
		KeyFunc func(eudore.Context) string `set:"keyfunc"`
		SetFunc func(eudore.Context, string, int)
	}
	// StoreMap 使用sync.Map实现的SessionStore。
	StoreMap struct {
		data sync.Map
	}
	// ContextSession 是使用Session实现的Context扩展。
	ContextSession struct {
		eudore.Context
		Session
	}
)

// NewSessionMap 创建一个SessionMap，使用sync.Map保存数据。
func NewSessionMap() Session {
	return NewSessionStd(&StoreMap{})
}

// NewSessionStd 创建一个使用Store为存储的Session对象。
func NewSessionStd(store SessionStore) Session {
	return &SessionStd{
		SessionStore: store,
		Maxage:       3600,
		KeyFunc: func(ctx eudore.Context) string {
			return eudore.GetDefaultString(ctx.GetCookie("sessionid"), "99")
		},
		SetFunc: func(ctx eudore.Context, key string, age int) {
			ctx.SetCookieValue("sessionid", key, age)
		},
	}
}

// GetSessionId 方法获取请求上下文的sessionid。
func (session *SessionStd) GetSessionId(ctx eudore.Context) string {
	return session.KeyFunc(ctx)
}

// SessionLoad 方法实现加载一个会话数据，
func (session *SessionStd) SessionLoad(ctx eudore.Context) map[string]interface{} {
	key := session.GetSessionId(ctx)
	data := session.SessionStore.Load(key)
	if data != nil {
		return data.(map[string]interface{})
	}
	session.SetFunc(ctx, key, session.Maxage)
	return make(map[string]interface{})
}

// SessionSave 方法实现将一个会话数据保存。
func (session *SessionStd) SessionSave(ctx eudore.Context, data map[string]interface{}) {
	session.SessionStore.Save(session.GetSessionId(ctx), data)
	session.SetFunc(ctx, session.GetSessionId(ctx), session.Maxage)
}

// SessionFlush 方法实现使用一个sessionid删除一个会话数据。
func (session *SessionStd) SessionFlush(ctx eudore.Context) {
	session.SessionStore.Delete(session.GetSessionId(ctx))
	session.SetFunc(ctx, session.GetSessionId(ctx), -1)
}

// Load 方法加载数据。
func (store *StoreMap) Load(key string) interface{} {
	data, ok := store.data.Load(key)
	if ok {
		return data
	}
	return nil
}

// Save 方法保存数据。
func (store *StoreMap) Save(key string, val interface{}) {
	store.data.Store(key, val)
}

// Delete 方法删除一个数据。
func (store *StoreMap) Delete(key string) {
	store.data.Delete(key)
}

// DeleteSession 方法删除当前会话数据
func (ctx ContextSession) DeleteSession() {
	ctx.Session.SessionFlush(ctx.Context)
}

// GetSession 获取会话数据。
func (ctx ContextSession) GetSession() map[string]interface{} {
	return ctx.Session.SessionLoad(ctx.Context)
}

// SetSession 方法设置当前会话的数据
func (ctx ContextSession) SetSession(data map[string]interface{}) {
	ctx.Session.SessionSave(ctx.Context, data)
}

// GetSessionBool 方法获取会话数据转换成bool。
func (ctx ContextSession) GetSessionBool(key string) bool {
	return eudore.GetDefaultBool(ctx.GetSession()[key], false)
}

// GetSessionDefaultBool 方法获取会话数据转换成bool，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultBool(key string, b bool) bool {
	return eudore.GetDefaultBool(ctx.GetSession()[key], b)
}

// GetSessionInt 方法获取会话数据转换成int。
func (ctx ContextSession) GetSessionInt(key string) int {
	return eudore.GetDefaultInt(ctx.GetSession()[key], 0)
}

// GetSessionDefaultInt 方法获取会话数据转换成int，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultInt(key string, i int) int {
	return eudore.GetDefaultInt(ctx.GetSession()[key], i)
}

// GetSessionFloat32 方法获取会话数据转换成float32。
func (ctx ContextSession) GetSessionFloat32(key string) float32 {
	return eudore.GetDefaultFloat32(ctx.GetSession()[key], 0)
}

// GetSessionDefaultFloat32 方法获取会话数据转换成float32，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultFloat32(key string, f float32) float32 {
	return eudore.GetDefaultFloat32(ctx.GetSession()[key], f)
}

// GetSessionFloat64 方法获取会话数据转换成float64。
func (ctx ContextSession) GetSessionFloat64(key string) float64 {
	return eudore.GetDefaultFloat64(ctx.GetSession()[key], 0)
}

// GetSessionDefaultFloat64 方法获取会话数据转换成float64，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultFloat64(key string, f float64) float64 {
	return eudore.GetDefaultFloat64(ctx.GetSession()[key], f)
}

// GetSessionString 方法获取会话数据转换成string。
func (ctx ContextSession) GetSessionString(key string) string {
	return eudore.GetDefaultString(ctx.GetSession()[key], "")
}

// GetSessionDefaultString 方法获取会话数据转换成string，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultString(key, str string) string {
	return eudore.GetDefaultString(ctx.GetSession()[key], str)
}
