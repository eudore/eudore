package eudore


import (
	"fmt"
	"sync"
	"time"
)

type (
	Session interface {
		Component
		SessionLoad(Context) SessionData
		SessionSave(SessionData)
		SessionFlush(string)
	}
	SessionData interface {
		Set(key string, value interface{}) error //set session value
		Get(key string) interface{}  //get session value
		Del(key string) error     //delete session value
		SessionID() string                //back current sessionID
	}
	// SessionMap会将数据存储到sync.Map放在内存，不会数据持久化。
	SessionMap struct {
		KeyFync	func(Context) string	`set:"keyfunc"`
		mem		sync.Map
	}
	SessionCacheConfig struct {
		Cache	Cache	`set:"cache"`
		Lifetime	time.Duration	`set:"lifetime"`
	}
	// SessionCache会将数据存储到eudore.Cache
	SessionCache struct {
		Cache		Cache	`set:"cache"`
		Lifetime	time.Duration		`set:"lifetime"`
		KeyFync	func(Context) string	`set:"keyfunc"`
	}
	SessionDataStd struct {
		Id		string
		Data	map[string]interface{}
	}
)

var (
	DefaultSession Session
)

func NewSession(name string, arg interface{}) (Session, error) {
	name = ComponentPrefix(name, "session")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	l, ok := c.(Session)
	if ok {
		return l, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Session type", name)
}


func NewSessionMap(interface{}) (Session, error) {
	return &SessionMap{
		KeyFync:	func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}, nil
}

func SessionLoad(ctx Context) SessionData {
	return DefaultSession.SessionLoad(ctx)
}

func SessionSave(sess SessionData) {
	DefaultSession.SessionSave(sess)
}

func SessionFlush(id string) {
	DefaultSession.SessionFlush(id)
}



func (store *SessionMap) SessionLoad(ctx Context) SessionData {
	key := store.KeyFync(ctx)
	sess, ok := store.mem.Load(key)
	if ok {
		return sess.(SessionData)
	}
	return &SessionDataStd{
		Id:		key,
		Data:	make(map[string]interface{}),
	}
}

func (store *SessionMap) SessionSave(sess SessionData) {
	store.mem.Store(sess.SessionID(), sess)
}

func (store *SessionMap) SessionFlush(id string) {
	store.mem.Delete(id)
}

func (*SessionMap) GetName() string {
	return ComponentSessionMapName
}

func (*SessionMap) Version() string {
	return ComponentSessionMapVersion
}




func NewSessionCache(i interface{}) (Session, error) {
	config := &SessionCacheConfig{
		Lifetime:	60 * time.Minute,
	}
	ConvertTo(i, config)
	return &SessionCache{
		Cache:		config.Cache,
		Lifetime:	config.Lifetime,
		KeyFync:    func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}, nil
}

func (store *SessionCache) SessionLoad(ctx Context) SessionData {
	key := store.KeyFync(ctx)
	sess := store.Cache.Get(key)
	if sess != nil {
		return sess.(SessionData)
	}
	return &SessionDataStd{
		Id:		key,
		Data:	make(map[string]interface{}),
	}
}

func (store *SessionCache) SessionSave(sess SessionData) {
	store.Cache.Set(sess.SessionID(), sess, store.Lifetime)
}

func (store *SessionCache) SessionFlush(id string) {
	store.Cache.Delete(id)
}

func (*SessionCache) GetName() string {
	return ComponentSessionCacheName
}

func (*SessionCache) Version() string {
	return ComponentSessionCacheVersion
}



func (sess *SessionDataStd) Set(key string, val interface{}) error {
	sess.Data[key] = val
	return nil
}

func (sess *SessionDataStd) Get(key string) interface{} {
	return sess.Data[key]
}

func (sess *SessionDataStd) Del(key string) error {
	delete(sess.Data, key)
	return nil
}

func (sess *SessionDataStd) SessionID() string {
	return sess.Id
}
